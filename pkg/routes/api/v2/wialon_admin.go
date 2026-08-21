// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package apiv2

import (
	"context"
	"net/http"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	wialonapi "code.vikunja.io/api/pkg/integrations/wialon"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

type adminWialonSettingsBody struct {
	Body struct {
		Enabled         bool   `json:"enabled"`
		APIURL          string `json:"api_url"`
		TokenConfigured bool   `json:"token_configured"`
		TimeoutSeconds  int    `json:"timeout_seconds"`
		TrackMaxPoints  int    `json:"track_max_points"`
	}
}

type adminWialonSettingsPatch struct {
	Enabled        *bool   `json:"enabled,omitempty"`
	APIURL         *string `json:"api_url,omitempty"`
	Token          *string `json:"token,omitempty"`
	ClearToken     bool    `json:"clear_token"`
	TimeoutSeconds *int    `json:"timeout_seconds,omitempty"`
	TrackMaxPoints *int    `json:"track_max_points,omitempty"`
}

type adminWialonTestBody struct {
	Body struct {
		OK        bool   `json:"ok"`
		UnitCount int    `json:"unit_count"`
		Message   string `json:"message"`
	}
}

func RegisterAdminWialonRoutes(api huma.API) {
	tags := []string{"admin", "wialon"}
	Register(api, huma.Operation{
		OperationID: "admin-wialon-settings-get", Summary: "Get Wialon connection settings",
		Method: http.MethodGet, Path: "/admin/wialon/settings", Tags: tags,
	}, adminWialonSettingsGet)
	Register(api, huma.Operation{
		OperationID: "admin-wialon-settings-update", Summary: "Update Wialon connection settings",
		Method: http.MethodPatch, Path: "/admin/wialon/settings", Tags: tags,
	}, adminWialonSettingsUpdate)
	Register(api, huma.Operation{
		OperationID: "admin-wialon-test", Summary: "Test the stored Wialon connection",
		Method: http.MethodPost, Path: "/admin/wialon/test", Tags: tags,
	}, adminWialonTest)
}

func init() { AddRouteRegistrar(RegisterAdminWialonRoutes) }

func renderAdminWialonSettings(settings *models.WialonSettings) *adminWialonSettingsBody {
	out := &adminWialonSettingsBody{}
	out.Body.Enabled = settings.Enabled
	out.Body.APIURL = settings.APIURL
	out.Body.TokenConfigured = strings.TrimSpace(settings.Token) != ""
	out.Body.TimeoutSeconds = settings.TimeoutSeconds
	out.Body.TrackMaxPoints = settings.TrackMaxPoints
	return out
}

func adminWialonSettingsGet(_ context.Context, _ *struct{}) (*adminWialonSettingsBody, error) {
	settings, err := effectiveWialonSettings()
	if err != nil {
		return nil, translateDomainError(err)
	}
	return renderAdminWialonSettings(settings), nil
}

func adminWialonSettingsUpdate(_ context.Context, in *struct{ Body adminWialonSettingsPatch }) (*adminWialonSettingsBody, error) {
	settings, err := effectiveWialonSettings()
	if err != nil {
		return nil, translateDomainError(err)
	}
	if in.Body.Enabled != nil {
		settings.Enabled = *in.Body.Enabled
	}
	if in.Body.APIURL != nil {
		settings.APIURL = strings.TrimSpace(*in.Body.APIURL)
	}
	if in.Body.TimeoutSeconds != nil {
		settings.TimeoutSeconds = *in.Body.TimeoutSeconds
	}
	if in.Body.TrackMaxPoints != nil {
		settings.TrackMaxPoints = *in.Body.TrackMaxPoints
	}
	if in.Body.ClearToken {
		settings.Token = ""
	} else if in.Body.Token != nil && strings.TrimSpace(*in.Body.Token) != "" {
		settings.Token = strings.TrimSpace(*in.Body.Token)
	}

	s := db.NewSession()
	defer s.Close()
	if err := models.SaveWialonSettings(s, settings); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	resetWialonClient()
	return renderAdminWialonSettings(settings), nil
}

func adminWialonTest(ctx context.Context, _ *struct{}) (*adminWialonTestBody, error) {
	settings, err := effectiveWialonSettings()
	if err != nil {
		return nil, translateDomainError(err)
	}
	token := strings.TrimSpace(settings.Token)
	if token == "" {
		return nil, huma.Error422UnprocessableEntity("Wialon token is not configured")
	}
	client, err := wialonapi.NewClient(wialonapi.Config{
		APIURL: settings.APIURL, Token: token,
		Timeout:        time.Duration(settings.TimeoutSeconds) * time.Second,
		TrackMaxPoints: settings.TrackMaxPoints,
	})
	if err != nil {
		return nil, huma.Error422UnprocessableEntity(err.Error())
	}
	units, err := client.ListUnits(ctx)
	if err != nil {
		return nil, wialonAdminRouteError(err)
	}
	out := &adminWialonTestBody{}
	out.Body.OK = true
	out.Body.UnitCount = len(units)
	out.Body.Message = "Wialon connection is working"
	return out, nil
}
