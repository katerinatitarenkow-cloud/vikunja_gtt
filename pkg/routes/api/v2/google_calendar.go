package apiv2

import (
	"context"
	"net/http"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	googlecalendar "code.vikunja.io/api/pkg/modules/googlecalendar"

	"github.com/danielgtaylor/huma/v2"
)

type googleCalendarStatusBody struct {
	Body struct {
		Enabled           bool      `json:"enabled"`
		Connected         bool      `json:"connected"`
		GoogleEmail       string    `json:"google_email,omitempty"`
		VikunjaCalendarID string    `json:"vikunja_calendar_id,omitempty"`
		ConnectedAt       time.Time `json:"connected_at,omitempty"`
	}
}

type googleCalendarConnectBody struct {
	Body struct {
		URL string `json:"url"`
	}
}

type googleCalendarCallbackInput struct {
	Code  string `json:"code" minLength:"1"`
	State string `json:"state" minLength:"1"`
}

func init() {
	AddRouteRegistrar(RegisterGoogleCalendarRoutes)
}

func RegisterGoogleCalendarRoutes(api huma.API) {
	tags := []string{"google-calendar"}

	Register(api, huma.Operation{
		OperationID: "google-calendar-status",
		Summary:     "Get Google Calendar connection status",
		Method:      http.MethodGet,
		Path:        "/integrations/google-calendar/status",
		Tags:        tags,
	}, googleCalendarStatus)

	Register(api, huma.Operation{
		OperationID: "google-calendar-connect",
		Summary:     "Start Google Calendar OAuth connection",
		Method:      http.MethodPost,
		Path:        "/integrations/google-calendar/connect",
		Tags:        tags,
	}, googleCalendarConnect)

	Register(api, huma.Operation{
		OperationID: "google-calendar-callback",
		Summary:     "Complete Google Calendar OAuth connection",
		Method:      http.MethodPost,
		Path:        "/integrations/google-calendar/callback",
		Tags:        tags,
	}, googleCalendarCallback)

	Register(api, huma.Operation{
		OperationID:   "google-calendar-disconnect",
		Summary:       "Disconnect Google Calendar",
		Method:        http.MethodDelete,
		Path:          "/integrations/google-calendar",
		Tags:          tags,
		DefaultStatus: http.StatusNoContent,
	}, googleCalendarDisconnect)
}

func googleCalendarStatus(
	ctx context.Context,
	_ *struct{},
) (*googleCalendarStatusBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	out := &googleCalendarStatusBody{}
	out.Body.Enabled = googlecalendar.IsConfigured()

	s := db.NewSession()
	defer s.Close()

	connection, found, err := models.GetGoogleCalendarConnection(s, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	if !found {
		return out, nil
	}

	out.Body.Connected =
		strings.TrimSpace(connection.RefreshTokenEncrypted) != ""

	out.Body.GoogleEmail = connection.GoogleEmail
	out.Body.VikunjaCalendarID = connection.VikunjaCalendarID
	out.Body.ConnectedAt = connection.ConnectedAt

	return out, nil
}

func googleCalendarConnect(
	ctx context.Context,
	_ *struct{},
) (*googleCalendarConnectBody, error) {
	if !googlecalendar.IsConfigured() {
		return nil, huma.Error503ServiceUnavailable(
			"Google Calendar integration is not configured",
		)
	}

	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	state, stateHash, err := googlecalendar.NewState()
	if err != nil {
		return nil, huma.Error500InternalServerError(
			"Could not create Google OAuth state",
		)
	}

	s := db.NewSession()
	defer s.Close()

	err = models.SaveGoogleCalendarOAuthState(
		s,
		a,
		stateHash,
		time.Now().Add(10*time.Minute),
	)

	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	out := &googleCalendarConnectBody{}
	out.Body.URL = googlecalendar.AuthorizationURL(state)

	return out, nil
}

func googleCalendarCallback(
	ctx context.Context,
	in *struct {
		Body googleCalendarCallbackInput
	},
) (*googleCalendarStatusBody, error) {
	if !googlecalendar.IsConfigured() {
		return nil, huma.Error503ServiceUnavailable(
			"Google Calendar integration is not configured",
		)
	}

	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	stateHash := googlecalendar.HashState(in.Body.State)

	stateSession := db.NewSession()

	valid, err := models.ValidateGoogleCalendarOAuthState(
		stateSession,
		a,
		stateHash,
		time.Now(),
	)

	if err != nil {
		_ = stateSession.Rollback()
		stateSession.Close()
		return nil, translateDomainError(err)
	}

	if !valid {
		_ = stateSession.Rollback()
		stateSession.Close()

		return nil, huma.Error400BadRequest(
			"Invalid or expired Google OAuth state",
		)
	}

	if err := stateSession.Commit(); err != nil {
		stateSession.Close()
		return nil, translateDomainError(err)
	}

	stateSession.Close()

	token, err := googlecalendar.Exchange(ctx, in.Body.Code)
	if err != nil {
		return nil, huma.Error502BadGateway(
			"Google rejected the authorization code",
		)
	}

	if strings.TrimSpace(token.RefreshToken) == "" {
		return nil, huma.Error400BadRequest(
			"Google did not return a refresh token; reconnect the account and grant access again",
		)
	}

	encryptedRefreshToken, err :=
		googlecalendar.EncryptRefreshToken(token.RefreshToken)

	if err != nil {
		return nil, huma.Error500InternalServerError(
			"Could not securely store Google credentials",
		)
	}

	s := db.NewSession()
	defer s.Close()

	if err := models.CompleteGoogleCalendarConnection(
		s,
		a,
		encryptedRefreshToken,
	); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	out := &googleCalendarStatusBody{}
	out.Body.Enabled = true
	out.Body.Connected = true
	out.Body.ConnectedAt = time.Now()

	return out, nil
}

func googleCalendarDisconnect(
	ctx context.Context,
	_ *struct{},
) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	if err := models.DeleteGoogleCalendarConnection(s, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	return &emptyBody{}, nil
}
