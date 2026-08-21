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

type workGroupsBody struct {
	Body struct {
		Groups []models.WorkGroupView `json:"groups"`
	}
}

type workGroupBody struct{ Body models.WorkGroupView }

type workGroupCreateBody struct {
	Name         string  `json:"name" minLength:"1" maxLength:"250"`
	Description  string  `json:"description"`
	LeaderUserID int64   `json:"leader_user_id" minimum:"0"`
	MemberIDs    []int64 `json:"member_ids"`
}

type workGroupPatchBody struct {
	Name         *string  `json:"name,omitempty" minLength:"1" maxLength:"250"`
	Description  *string  `json:"description,omitempty"`
	LeaderUserID *int64   `json:"leader_user_id,omitempty" minimum:"0"`
	MemberIDs    *[]int64 `json:"member_ids,omitempty"`
}

type taskWorkGroupAssigneesBody struct {
	Body struct {
		Groups []models.WorkGroupView `json:"groups"`
	}
}

type taskWorkGroupAssignBody struct {
	GroupID int64 `json:"group_id" minimum:"1"`
}

type taskWorkGroupAssignResultBody struct {
	Body models.WorkGroupTaskAssignmentResult
}

func RegisterWorkGroupRoutes(api huma.API) {
	tags := []string{"work-groups"}

	Register(api, huma.Operation{
		OperationID: "work-groups-list",
		Summary:     "List operational user groups",
		Description: "Lists temporary operational groups. Group membership does not grant any system or project permissions.",
		Method:      http.MethodGet,
		Path:        "/work-groups",
		Tags:        tags,
	}, workGroupsList)

	Register(api, huma.Operation{
		OperationID: "admin-work-groups-create",
		Summary:     "Create an operational user group",
		Method:      http.MethodPost,
		Path:        "/admin/work-groups",
		Tags:        tags,
	}, adminWorkGroupsCreate)
	Register(api, huma.Operation{
		OperationID: "admin-work-groups-update",
		Summary:     "Update an operational user group",
		Method:      http.MethodPatch,
		Path:        "/admin/work-groups/{id}",
		Tags:        tags,
	}, adminWorkGroupsUpdate)
	Register(api, huma.Operation{
		OperationID: "admin-work-groups-delete",
		Summary:     "Delete an operational user group",
		Method:      http.MethodDelete,
		Path:        "/admin/work-groups/{id}",
		Tags:        tags,
	}, adminWorkGroupsDelete)

	Register(api, huma.Operation{
		OperationID: "task-work-group-assignees-list",
		Summary:     "List work groups assigned to a task",
		Method:      http.MethodGet,
		Path:        "/tasks/{projecttask}/group-assignees",
		Tags:        []string{"assignees", "work-groups"},
	}, taskWorkGroupAssigneesList)
	Register(api, huma.Operation{
		OperationID: "task-work-group-assignees-create",
		Summary:     "Assign a work group to a task",
		Description: "Assigns the group semantically and materializes its eligible members as normal Vikunja task assignees. Membership itself never grants project access.",
		Method:      http.MethodPost,
		Path:        "/tasks/{projecttask}/group-assignees",
		Tags:        []string{"assignees", "work-groups"},
	}, taskWorkGroupAssigneesCreate)
	Register(api, huma.Operation{
		OperationID: "task-work-group-assignees-delete",
		Summary:     "Remove a work group from a task",
		Method:      http.MethodDelete,
		Path:        "/tasks/{projecttask}/group-assignees/{group}",
		Tags:        []string{"assignees", "work-groups"},
	}, taskWorkGroupAssigneesDelete)
}

func init() { AddRouteRegistrar(RegisterWorkGroupRoutes) }

func workGroupsList(ctx context.Context, in *struct {
	Search string `query:"search"`
}) (*workGroupsBody, error) {
	if _, err := requireAccessPermission(ctx, models.PermissionTasksView); err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	groups, err := models.ListWorkGroups(s, in.Search)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &workGroupsBody{}
	out.Body.Groups = groups
	return out, nil
}

func adminWorkGroupsCreate(ctx context.Context, in *struct{ Body workGroupCreateBody }) (*workGroupBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	group, err := models.CreateWorkGroup(s, in.Body.Name, in.Body.Description, in.Body.LeaderUserID, in.Body.MemberIDs, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &workGroupBody{Body: *group}, nil
}

func adminWorkGroupsUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id" minimum:"1"`
	Body workGroupPatchBody
}) (*workGroupBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	group, err := models.UpdateWorkGroup(s, in.ID, in.Body.Name, in.Body.Description, in.Body.LeaderUserID, in.Body.MemberIDs, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &workGroupBody{Body: *group}, nil
}

func adminWorkGroupsDelete(ctx context.Context, in *struct {
	ID int64 `path:"id" minimum:"1"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	if err := models.DeleteWorkGroup(s, in.ID, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func taskWorkGroupAssigneesList(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask" minimum:"1"`
}) (*taskWorkGroupAssigneesBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	groups, err := models.ListTaskWorkGroupAssignees(s, in.TaskID, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &taskWorkGroupAssigneesBody{}
	out.Body.Groups = groups
	return out, nil
}

func taskWorkGroupAssigneesCreate(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask" minimum:"1"`
	Body   taskWorkGroupAssignBody
}) (*taskWorkGroupAssignResultBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	result, err := models.AssignWorkGroupToTask(s, in.TaskID, in.Body.GroupID, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &taskWorkGroupAssignResultBody{Body: *result}, nil
}

func taskWorkGroupAssigneesDelete(ctx context.Context, in *struct {
	TaskID  int64 `path:"projecttask" minimum:"1"`
	GroupID int64 `path:"group" minimum:"1"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	if err := models.UnassignWorkGroupFromTask(s, in.TaskID, in.GroupID, a); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
