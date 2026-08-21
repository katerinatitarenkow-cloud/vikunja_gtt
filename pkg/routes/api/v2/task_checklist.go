// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

type taskChecklistStateBody struct {
	Body models.TaskChecklistState
}

type taskChecklistCreateBody struct {
	Title string `json:"title" minLength:"1" maxLength:"1000" doc:"The checklist item text."`
}

type taskChecklistUpdateBody struct {
	Title string `json:"title" minLength:"1" maxLength:"1000" doc:"The checklist item text."`
	Done  bool   `json:"done" doc:"Whether the checklist item is completed."`
}

func RegisterTaskChecklistRoutes(api huma.API) {
	tags := []string{"tasks", "checklist"}

	Register(api, huma.Operation{
		OperationID: "task-checklist-list",
		Summary:     "List structured checklist items",
		Description: "Returns the ordered checklist inside a task, including who completed each finished item and when. Requires read access to the task.",
		Method:      http.MethodGet,
		Path:        "/tasks/{task}/checklist-items",
		Tags:        tags,
	}, taskChecklistList)

	Register(api, huma.Operation{
		OperationID: "task-checklist-create",
		Summary:     "Create a checklist item",
		Description: "Appends a pending checklist item to the task. Requires write access. Adding pending work reopens a completed task.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/checklist-items",
		Tags:        tags,
	}, taskChecklistCreate)

	Register(api, huma.Operation{
		OperationID: "task-checklist-update",
		Summary:     "Update a checklist item",
		Description: "Changes the item text or completion state. The server records the authenticated user and timestamp on completion. Completing the last pending item completes the parent task.",
		Method:      http.MethodPut,
		Path:        "/tasks/{task}/checklist-items/{item}",
		Tags:        tags,
	}, taskChecklistUpdate)

	Register(api, huma.Operation{
		OperationID: "task-checklist-delete",
		Summary:     "Delete a checklist item",
		Description: "Deletes a checklist item and recomputes the parent task completion state. Requires write access.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{task}/checklist-items/{item}",
		Tags:        tags,
	}, taskChecklistDelete)
}

func init() { AddRouteRegistrar(RegisterTaskChecklistRoutes) }

func taskChecklistCanRead(s *xorm.Session, a web.Auth, taskID int64) error {
	task := &models.Task{ID: taskID}
	can, _, err := task.CanRead(s, a)
	if err != nil {
		return err
	}
	if !can {
		return models.ErrGenericForbidden{}
	}
	return nil
}

func taskChecklistCanWrite(s *xorm.Session, a web.Auth, taskID int64) error {
	task := &models.Task{ID: taskID}
	can, err := task.CanWrite(s, a)
	if err != nil {
		return err
	}
	if !can {
		return models.ErrGenericForbidden{}
	}
	return nil
}

func taskChecklistList(ctx context.Context, in *struct {
	TaskID int64 `path:"task" minimum:"1" doc:"The parent task id."`
}) (*taskChecklistStateBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := taskChecklistCanRead(s, a, in.TaskID); err != nil {
		return nil, translateDomainError(err)
	}
	state, err := models.GetTaskChecklistState(s, in.TaskID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &taskChecklistStateBody{Body: *state}, nil
}

func taskChecklistCreate(ctx context.Context, in *struct {
	TaskID int64 `path:"task" minimum:"1" doc:"The parent task id."`
	Body   taskChecklistCreateBody
}) (*taskChecklistStateBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := taskChecklistCanWrite(s, a, in.TaskID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if _, err := models.CreateTaskChecklistItem(s, in.TaskID, in.Body.Title); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if _, err := models.SyncTaskDoneFromChecklist(s, a, in.TaskID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	state, err := models.GetTaskChecklistState(s, in.TaskID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &taskChecklistStateBody{Body: *state}, nil
}

func taskChecklistUpdate(ctx context.Context, in *struct {
	TaskID int64 `path:"task" minimum:"1" doc:"The parent task id."`
	ItemID int64 `path:"item" minimum:"1" doc:"The checklist item id."`
	Body   taskChecklistUpdateBody
}) (*taskChecklistStateBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := taskChecklistCanWrite(s, a, in.TaskID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	_, found, err := models.UpdateTaskChecklistItem(s, a, in.TaskID, in.ItemID, in.Body.Title, in.Body.Done)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("checklist item not found")
	}
	if _, err := models.SyncTaskDoneFromChecklist(s, a, in.TaskID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	state, err := models.GetTaskChecklistState(s, in.TaskID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &taskChecklistStateBody{Body: *state}, nil
}

func taskChecklistDelete(ctx context.Context, in *struct {
	TaskID int64 `path:"task" minimum:"1" doc:"The parent task id."`
	ItemID int64 `path:"item" minimum:"1" doc:"The checklist item id."`
}) (*taskChecklistStateBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	if err := taskChecklistCanWrite(s, a, in.TaskID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	found, err := models.DeleteTaskChecklistItem(s, in.TaskID, in.ItemID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if !found {
		_ = s.Rollback()
		return nil, huma.Error404NotFound("checklist item not found")
	}
	if _, err := models.SyncTaskDoneFromChecklist(s, a, in.TaskID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	state, err := models.GetTaskChecklistState(s, in.TaskID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &taskChecklistStateBody{Body: *state}, nil
}
