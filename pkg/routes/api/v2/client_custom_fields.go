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

type clientCustomFieldsBody struct {
	Body []*models.ClientCustomField
}

type clientCustomFieldBody struct {
	Body models.ClientCustomField
}

type clientCustomFieldWriteBody struct {
	Name  string `json:"name" minLength:"1" maxLength:"500" doc:"User-visible custom field name."`
	Value string `json:"value" doc:"Free-form custom field value."`
}

func RegisterClientCustomFieldRoutes(api huma.API) {
	tags := []string{"project", "client", "custom fields"}

	Register(api, huma.Operation{
		OperationID: "project-client-custom-fields-list",
		Summary:     "List custom client fields",
		Description: "Returns ordered user-defined name/value fields for the CRM client card. Public link shares are excluded.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/client/custom-fields",
		Tags:        tags,
	}, clientCustomFieldsList)

	Register(api, huma.Operation{
		OperationID:   "project-client-custom-fields-create",
		Summary:       "Create a custom client field",
		Description:   "Appends a custom name/value field to the CRM card and records the change in client history. Requires project write access.",
		Method:        http.MethodPost,
		Path:          "/projects/{project}/client/custom-fields",
		Tags:          tags,
		DefaultStatus: http.StatusCreated,
	}, clientCustomFieldsCreate)

	Register(api, huma.Operation{
		OperationID: "project-client-custom-fields-update",
		Summary:     "Update a custom client field",
		Description: "Updates the custom field name/value and records old/new values in client history. Requires project write access.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/client/custom-fields/{field}",
		Tags:        tags,
	}, clientCustomFieldsUpdate)

	Register(api, huma.Operation{
		OperationID:   "project-client-custom-fields-delete",
		Summary:       "Delete a custom client field",
		Description:   "Deletes a custom field and records the removal in client history. Requires project write access.",
		Method:        http.MethodDelete,
		Path:          "/projects/{project}/client/custom-fields/{field}",
		Tags:          tags,
		DefaultStatus: http.StatusNoContent,
	}, clientCustomFieldsDelete)
}

func init() { AddRouteRegistrar(RegisterClientCustomFieldRoutes) }

func clientCustomFieldsList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
}) (*clientCustomFieldsBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	field := &models.ClientCustomField{ProjectID: in.ProjectID}
	can, _, err := field.CanRead(s, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if !can {
		return nil, translateDomainError(models.ErrGenericForbidden{})
	}
	fields, err := models.GetClientCustomFields(s, in.ProjectID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &clientCustomFieldsBody{Body: fields}, nil
}

func clientCustomFieldsCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
	Body      clientCustomFieldWriteBody
}) (*clientCustomFieldBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	field, err := models.CreateClientCustomField(s, a, in.ProjectID, in.Body.Name, in.Body.Value)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &clientCustomFieldBody{Body: *field}, nil
}

func clientCustomFieldsUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
	FieldID   int64 `path:"field" minimum:"1"`
	Body      clientCustomFieldWriteBody
}) (*clientCustomFieldBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	field, found, err := models.UpdateClientCustomField(s, a, in.ProjectID, in.FieldID, in.Body.Name, in.Body.Value)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("custom field not found")
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &clientCustomFieldBody{Body: *field}, nil
}

func clientCustomFieldsDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project" minimum:"1"`
	FieldID   int64 `path:"field" minimum:"1"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	found, err := models.DeleteClientCustomField(s, a, in.ProjectID, in.FieldID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("custom field not found")
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
