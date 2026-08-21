// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package apiv2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	vikfiles "code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
	webfiles "code.vikunja.io/api/pkg/web/files"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/gabriel-vasile/mimetype"
	"xorm.io/xorm"
)

type clientProfileBody struct {
	Body models.ClientProfile
}

type clientGeocodeBody struct {
	Body clientGeocodeResult
}

type clientGeocodeResult struct {
	DisplayName string  `json:"display_name"`
	PostalCode  string  `json:"postal_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// clientProposalUploadInput accepts exactly one PDF proposal. Byte-level MIME
// validation still runs in the handler; octet-stream is accepted here for clients
// that do not set a precise multipart part type.
type clientProposalUploadInput struct {
	ProjectID int64 `path:"project" minimum:"1"`
	RawBody   huma.MultipartFormFiles[struct {
		Proposal huma.FormFile `form:"proposal" contentType:"application/pdf,application/octet-stream" required:"true" doc:"Commercial proposal PDF."`
	}]
}

var (
	nominatimMu          sync.Mutex
	nominatimLastRequest time.Time
)

func RegisterClientProfileRoutes(api huma.API) {
	tags := []string{"project", "client"}

	Register(api, huma.Operation{
		OperationID: "project-client-read",
		Summary:     "Get the client card for a project",
		Description: "Returns the CRM client/company card attached one-to-one to a project. Requires project read access. Public link shares are excluded because the card contains private contact data.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/client",
		Tags:        tags,
	}, clientProfileRead)

	Register(api, huma.Operation{
		OperationID: "project-client-update",
		Summary:     "Save the client card for a project",
		Description: "Saves the CRM client card, legal details, five address records and structured contact persons. The responsible employee must already have access to the project. Requires project write access.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/client",
		Tags:        tags,
	}, clientProfileUpdate)

	Register(api, huma.Operation{
		OperationID: "project-client-geocode",
		Summary:     "Resolve a client address",
		Description: "Uses OpenStreetMap Nominatim to resolve coordinates and postal code for a client address. Requires project read access.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/client/geocode",
		Tags:        tags,
	}, clientAddressGeocode)

	Register(api, huma.Operation{
		OperationID:   "project-client-proposal-upload",
		Summary:       "Upload a commercial proposal PDF",
		Description:   "Stores or replaces the single PDF commercial proposal attached to the client card. Requires project write access.",
		Method:        http.MethodPut,
		Path:          "/projects/{project}/client/proposal",
		Tags:          tags,
		MaxBodyBytes:  (int64(config.GetMaxFileSizeInMBytes()) + 2) * 1024 * 1024, //nolint:gosec // configured limit is bounded in practice
		DefaultStatus: http.StatusOK,
	}, clientProposalUpload)

	Register(api, huma.Operation{
		OperationID: "project-client-proposal-download",
		Summary:     "Download the commercial proposal PDF",
		Description: "Streams the proposal PDF for the client card. Requires project read access; public link shares are excluded.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/client/proposal",
		Tags:        tags,
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The commercial proposal PDF.",
				Content: map[string]*huma.MediaType{
					"application/pdf": {Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"}},
				},
			},
		},
	}, clientProposalDownload)

	Register(api, huma.Operation{
		OperationID: "project-client-proposal-delete",
		Summary:     "Delete the commercial proposal PDF",
		Description: "Removes the proposal file from the client card and storage. Requires project write access.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/client/proposal",
		Tags:        tags,
	}, clientProposalDelete)
}

func init() { AddRouteRegistrar(RegisterClientProfileRoutes) }

func clientProfileCanRead(s *xorm.Session, a web.Auth, projectID int64) error {
	if _, isLinkShare := a.(*models.LinkSharing); isLinkShare {
		return models.ErrGenericForbidden{}
	}
	can, _, err := (&models.Project{ID: projectID}).CanRead(s, a)
	if err != nil {
		return err
	}
	if !can {
		return models.ErrGenericForbidden{}
	}
	return nil
}

func clientProfileCanWrite(s *xorm.Session, a web.Auth, projectID int64) error {
	if _, isLinkShare := a.(*models.LinkSharing); isLinkShare {
		return models.ErrGenericForbidden{}
	}
	can, err := (&models.Project{ID: projectID}).CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !can {
		return models.ErrGenericForbidden{}
	}
	return nil
}

func clientProfileRead(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
}) (*clientProfileBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := clientProfileCanRead(s, a, in.ProjectID); err != nil {
		return nil, translateDomainError(err)
	}
	profile, err := models.GetClientProfile(s, in.ProjectID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &clientProfileBody{Body: *profile}, nil
}

func clientProfileUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
	Body      models.ClientProfile
}) (*clientProfileBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := clientProfileCanWrite(s, a, in.ProjectID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	profile, err := models.SaveClientProfile(s, a, in.ProjectID, &in.Body)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &clientProfileBody{Body: *profile}, nil
}

type nominatimSearchResult struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Address     struct {
		Postcode string `json:"postcode"`
	} `json:"address"`
}

// waitForNominatim serializes requests from this Vikunja instance. Public
// Nominatim asks clients to stay at or below one request per second.
func waitForNominatim(ctx context.Context) error {
	nominatimMu.Lock()
	defer nominatimMu.Unlock()

	if wait := time.Until(nominatimLastRequest.Add(time.Second)); wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	nominatimLastRequest = time.Now()
	return nil
}

func clientAddressGeocode(ctx context.Context, in *struct {
	ProjectID int64  `path:"project" minimum:"1"`
	Query     string `query:"q" minLength:"3" maxLength:"2000"`
}) (*clientGeocodeBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := clientProfileCanRead(s, a, in.ProjectID); err != nil {
		return nil, translateDomainError(err)
	}

	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, huma.Error400BadRequest("address is required")
	}
	if err := waitForNominatim(ctx); err != nil {
		return nil, huma.Error503ServiceUnavailable("address lookup was cancelled")
	}

	endpoint := "https://nominatim.openstreetmap.org/search?format=jsonv2&addressdetails=1&limit=1&q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, huma.Error502BadGateway("could not build geocoding request")
	}
	req.Header.Set("User-Agent", "Vikunja-GreenTech-CRM/1.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, huma.Error502BadGateway("address lookup service is unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, huma.Error502BadGateway(fmt.Sprintf("address lookup returned HTTP %d", resp.StatusCode))
	}

	var rows []nominatimSearchResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rows); err != nil {
		return nil, huma.Error502BadGateway("could not decode address lookup response")
	}
	if len(rows) == 0 {
		return nil, huma.Error404NotFound("address was not found")
	}

	lat, err := strconv.ParseFloat(rows[0].Lat, 64)
	if err != nil {
		return nil, huma.Error502BadGateway("address lookup returned an invalid latitude")
	}
	lon, err := strconv.ParseFloat(rows[0].Lon, 64)
	if err != nil {
		return nil, huma.Error502BadGateway("address lookup returned an invalid longitude")
	}

	return &clientGeocodeBody{Body: clientGeocodeResult{
		DisplayName: rows[0].DisplayName,
		PostalCode:  rows[0].Address.Postcode,
		Latitude:    lat,
		Longitude:   lon,
	}}, nil
}

func clientProposalUpload(ctx context.Context, in *clientProposalUploadInput) (*clientProfileBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := clientProfileCanWrite(s, a, in.ProjectID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	file := in.RawBody.Data().Proposal
	defer func() { _ = file.Close() }()
	mime, err := mimetype.DetectReader(file)
	if err != nil {
		_ = s.Rollback()
		return nil, huma.Error400BadRequest("could not inspect proposal file")
	}
	if !mime.Is("application/pdf") {
		_ = s.Rollback()
		return nil, huma.Error400BadRequest("commercial proposal must be a PDF file")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		_ = s.Rollback()
		return nil, huma.Error400BadRequest("could not read proposal file")
	}

	profile, err := models.SetClientCommercialProposal(s, a, in.ProjectID, file, file.Filename, uint64(file.Size))
	if err != nil {
		_ = s.Rollback()
		if vikfiles.IsErrFileIsTooLarge(err) {
			return nil, huma.NewError(http.StatusRequestEntityTooLarge, "commercial proposal exceeds the configured file size limit")
		}
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	return &clientProfileBody{Body: *profile}, nil
}

func clientProposalDownload(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
}) (*huma.StreamResponse, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := clientProfileCanRead(s, a, in.ProjectID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	profile := &models.ClientProfile{ProjectID: in.ProjectID}
	has, err := s.ID(in.ProjectID).Get(profile)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !has || profile.CommercialProposalFileID == 0 {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("commercial proposal is not configured")
	}
	file := &vikfiles.File{ID: profile.CommercialProposalFileID}
	has, err = s.ID(file.ID).Get(file)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !has {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("commercial proposal file is missing")
	}
	if err := file.LoadFileByID(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		_ = file.File.Close()
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		defer func() { _ = file.File.Close() }()
		c := humaecho.Unwrap(hctx)
		webfiles.WriteFileDownload((*c).Response(), (*c).Request(), file)
	}}, nil
}

func clientProposalDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
}) (*clientProfileBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := clientProfileCanWrite(s, a, in.ProjectID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := models.DeleteClientCommercialProposal(s, a, in.ProjectID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	profile, err := models.GetClientProfile(s, in.ProjectID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	return &clientProfileBody{Body: *profile}, nil
}
