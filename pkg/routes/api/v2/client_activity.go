// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

type clientActivityListResult struct {
	Items   []*models.ClientActivityEvent `json:"items" doc:"Client history events ordered newest first."`
	Total   int64                         `json:"total" doc:"Total matching history events."`
	Page    int                           `json:"page" doc:"Current page number."`
	PerPage int                           `json:"per_page" doc:"Maximum events returned per page."`
}

type clientActivityListBody struct {
	Body clientActivityListResult
}

type clientActivityBody struct {
	Body models.ClientActivityEvent
}

func RegisterClientActivityRoutes(api huma.API) {
	tags := []string{"project", "client", "history"}

	Register(api, huma.Operation{
		OperationID: "project-client-history-list",
		Summary:     "List client activity history",
		Description: "Returns the chronological CRM history for a project/client. Public link shares are excluded because the history contains private customer activity.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/client/history",
		Tags:        tags,
	}, clientActivityList)

	Register(api, huma.Operation{
		OperationID: "project-client-history-create",
		Summary:     "Record a manual client activity",
		Description: "Records a call, message, meeting, note, sent document, proposal or invoice in the client history. Requires project write access.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/client/history",
		Tags:        tags,
	}, clientActivityCreate)

	Register(api, huma.Operation{
		OperationID: "project-client-history-delete",
		Summary:     "Delete a manual client activity",
		Description: "Deletes a user-entered CRM history item. Automatically generated system history is immutable. Requires project write access.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/client/history/{event}",
		Tags:        tags,
	}, clientActivityDelete)
}

func init() { AddRouteRegistrar(RegisterClientActivityRoutes) }

func clientActivityList(ctx context.Context, in *struct {
	ProjectID int64  `path:"project" minimum:"1"`
	EventType string `query:"type" maxLength:"80"`
	Page      int    `query:"page" default:"1" minimum:"1"`
	PerPage   int    `query:"per_page" default:"50" minimum:"1" maximum:"200"`
}) (*clientActivityListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := clientProfileCanRead(s, a, in.ProjectID); err != nil {
		return nil, translateDomainError(err)
	}
	items, total, err := models.GetClientActivityEvents(s, in.ProjectID, in.EventType, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &clientActivityListBody{Body: clientActivityListResult{
		Items: items, Total: total, Page: in.Page, PerPage: in.PerPage,
	}}, nil
}

func clientActivityCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
	Body      models.ClientActivityCreate
}) (*clientActivityBody, error) {
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
	event, err := models.CreateManualClientActivity(s, a, in.ProjectID, &in.Body)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &clientActivityBody{Body: *event}, nil
}

func clientActivityDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
	EventID   int64 `path:"event" minimum:"1"`
}) (*emptyBody, error) {
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
	found, err := models.DeleteManualClientActivity(s, in.ProjectID, in.EventID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("client history event not found")
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
