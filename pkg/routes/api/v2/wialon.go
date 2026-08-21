// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package apiv2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	wialonapi "code.vikunja.io/api/pkg/integrations/wialon"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

const maxWialonTrackInterval = 31 * 24 * time.Hour

type wialonStatusBody struct {
	Body struct {
		Enabled    bool   `json:"enabled" doc:"Whether the Wialon integration is enabled in the Vikunja server configuration."`
		Configured bool   `json:"configured" doc:"Whether a Wialon access token is configured on the server."`
		APIURL     string `json:"api_url" doc:"The configured Wialon API host. The access token is intentionally never returned."`
	}
}

type wialonUnitsBody struct {
	Body struct {
		Units []wialonapi.Unit `json:"units" doc:"Wialon units visible to the configured token, with their latest known positions."`
	}
}

type wialonTrackBody struct {
	Body wialonapi.Track
}

var (
	wialonClientMu        sync.Mutex
	wialonClient          *wialonapi.Client
	wialonClientSignature string
)

// RegisterWialonRoutes exposes the server-side Wialon adapter to authenticated
// Vikunja users. The Wialon token remains in server configuration; these routes
// return only fleet data needed by the UI and, later, task automation.
func RegisterWialonRoutes(api huma.API) {
	tags := []string{"wialon"}

	Register(api, huma.Operation{
		OperationID: "wialon-status",
		Summary:     "Get Wialon integration status",
		Description: "Returns whether the fleet integration is enabled and configured. The Wialon token is server-side only and is never included in the response.",
		Method:      http.MethodGet,
		Path:        "/wialon/status",
		Tags:        tags,
	}, wialonStatus)

	Register(api, huma.Operation{
		OperationID: "wialon-units-list",
		Summary:     "List Wialon units",
		Description: "Returns Wialon units visible to the server token, including last known GPS position and connection state. Requires a logged-in Vikunja user.",
		Method:      http.MethodGet,
		Path:        "/wialon/units",
		Tags:        tags,
	}, wialonUnitsList)

	Register(api, huma.Operation{
		OperationID: "wialon-unit-track",
		Summary:     "Get a Wialon unit track",
		Description: "Builds a historical route from Wialon data messages containing GPS locations. The interval defaults to the previous 24 hours and is capped at 31 days. Long tracks are downsampled server-side for safe browser rendering.",
		Method:      http.MethodGet,
		Path:        "/wialon/units/{unit}/track",
		Tags:        tags,
	}, wialonUnitTrack)
}

func init() { AddRouteRegistrar(RegisterWialonRoutes) }

func wialonStatus(ctx context.Context, _ *struct{}) (*wialonStatusBody, error) {
	if _, err := requireAccessPermission(ctx, models.PermissionWialonView); err != nil {
		return nil, err
	}
	settings, err := effectiveWialonSettings()
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &wialonStatusBody{}
	out.Body.Enabled = settings.Enabled
	out.Body.Configured = strings.TrimSpace(settings.Token) != ""
	out.Body.APIURL = settings.APIURL
	return out, nil
}

func wialonUnitsList(ctx context.Context, _ *struct{}) (*wialonUnitsBody, error) {
	if _, err := requireAccessPermission(ctx, models.PermissionWialonView); err != nil {
		return nil, err
	}
	client, err := configuredWialonClient()
	if err != nil {
		return nil, wialonRouteError(err)
	}
	units, err := client.ListUnits(ctx)
	if err != nil {
		return nil, wialonRouteError(err)
	}
	out := &wialonUnitsBody{}
	out.Body.Units = units
	return out, nil
}

func wialonUnitTrack(ctx context.Context, in *struct {
	UnitID int64 `path:"unit" minimum:"1" doc:"Wialon numeric unit id."`
	From   int64 `query:"from" minimum:"0" doc:"Start of the track interval as a Unix timestamp in seconds. Defaults to 24 hours before 'to'."`
	To     int64 `query:"to" minimum:"0" doc:"End of the track interval as a Unix timestamp in seconds. Defaults to the current time."`
}) (*wialonTrackBody, error) {
	if _, err := requireAccessPermission(ctx, models.PermissionWialonView); err != nil {
		return nil, err
	}

	to := in.To
	if to == 0 {
		to = time.Now().Unix()
	}
	from := in.From
	if from == 0 {
		from = to - int64((24*time.Hour)/time.Second)
	}
	if from >= to {
		return nil, huma.Error422UnprocessableEntity("'from' must be earlier than 'to'")
	}
	if to-from > int64(maxWialonTrackInterval/time.Second) {
		return nil, huma.Error422UnprocessableEntity("Wialon track interval cannot exceed 31 days")
	}

	client, err := configuredWialonClient()
	if err != nil {
		return nil, wialonRouteError(err)
	}
	track, err := client.LoadTrack(ctx, in.UnitID, from, to)
	if err != nil {
		return nil, wialonRouteError(err)
	}
	return &wialonTrackBody{Body: *track}, nil
}

func effectiveWialonSettings() (*models.WialonSettings, error) {
	s := db.NewSession()
	defer s.Close()
	settings, has, err := models.LoadWialonSettings(s)
	if err != nil {
		return nil, err
	}
	if has {
		return settings, nil
	}
	return &models.WialonSettings{
		ID:             1,
		Enabled:        config.WialonEnabled.GetBool(),
		APIURL:         config.WialonAPIURL.GetString(),
		Token:          config.WialonToken.GetString(),
		TimeoutSeconds: config.WialonTimeoutSeconds.GetInt(),
		TrackMaxPoints: config.WialonTrackMaxPoints.GetInt(),
	}, nil
}

func resetWialonClient() {
	wialonClientMu.Lock()
	defer wialonClientMu.Unlock()
	wialonClient = nil
	wialonClientSignature = ""
}

func configuredWialonClient() (*wialonapi.Client, error) {
	settings, err := effectiveWialonSettings()
	if err != nil {
		return nil, err
	}
	if !settings.Enabled {
		return nil, errors.New("Wialon integration is disabled")
	}
	token := strings.TrimSpace(settings.Token)
	if token == "" {
		return nil, errors.New("Wialon token is not configured")
	}

	apiURL := strings.TrimSpace(settings.APIURL)
	timeout := settings.TimeoutSeconds
	maxPoints := settings.TrackMaxPoints
	signature := fmt.Sprintf("%s\x00%s\x00%d\x00%d", apiURL, token, timeout, maxPoints)

	wialonClientMu.Lock()
	defer wialonClientMu.Unlock()
	if wialonClient != nil && wialonClientSignature == signature {
		return wialonClient, nil
	}

	client, err := wialonapi.NewClient(wialonapi.Config{
		APIURL:         apiURL,
		Token:          token,
		Timeout:        time.Duration(timeout) * time.Second,
		TrackMaxPoints: maxPoints,
	})
	if err != nil {
		return nil, err
	}
	wialonClient = client
	wialonClientSignature = signature
	return wialonClient, nil
}

func wialonRouteError(err error) error {
	var apiErr *wialonapi.APIError
	if errors.As(err, &apiErr) {
		return huma.Error502BadGateway(fmt.Sprintf("Wialon API returned error code %d (%s)", apiErr.Code, wialonErrorDescription(apiErr.Code)))
	}
	if strings.Contains(err.Error(), "disabled") || strings.Contains(err.Error(), "not configured") {
		return huma.Error503ServiceUnavailable(err.Error())
	}
	return huma.Error502BadGateway("Could not load data from Wialon")
}

func wialonAdminRouteError(err error) error {
	var apiErr *wialonapi.APIError
	if errors.As(err, &apiErr) {
		return huma.Error502BadGateway(fmt.Sprintf("%s: Wialon API error %d (%s)", wialonErrorStage(err), apiErr.Code, wialonErrorDescription(apiErr.Code)))
	}
	// This endpoint is administrator-only. Return the sanitized technical
	// reason so connection setup can be diagnosed without exposing the token.
	return huma.Error502BadGateway(err.Error())
}

func wialonErrorStage(err error) string {
	text := err.Error()
	for _, stage := range []string{"token/login", "core/search_items", "messages/load_interval"} {
		if strings.Contains(text, stage) {
			return stage
		}
	}
	return "Wialon"
}

func wialonErrorDescription(code int) string {
	switch code {
	case 1:
		return "invalid session or missing service"
	case 2:
		return "invalid service name"
	case 3:
		return "invalid result"
	case 4:
		return "invalid input"
	case 5:
		return "error performing request"
	case 6:
		return "unknown error"
	case 7:
		return "access denied"
	case 8:
		return "invalid user name or password"
	case 9:
		return "authorization server unavailable"
	case 10:
		return "concurrent request limit reached"
	case 1001:
		return "no messages for selected interval"
	default:
		return "see Wialon Remote API error reference"
	}
}
