// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package apiv2

import (
	"context"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

type accessMeBody struct {
	Body struct {
		Permissions []string                 `json:"permissions"`
		Groups      []models.AccessGroupView `json:"groups"`
	}
}

type accessPermissionCatalogBody struct {
	Body struct {
		Permissions []models.PermissionDefinition `json:"permissions"`
	}
}

type accessGroupsBody struct {
	Body struct {
		Groups []models.AccessGroupView `json:"groups"`
	}
}

type accessGroupBody struct{ Body models.AccessGroupView }

type accessUsersBody struct {
	Body struct {
		Users []models.AccessUserView `json:"users"`
	}
}

type accessUserBody struct{ Body models.AccessUserView }

type accessGroupCreate struct {
	Name        string   `json:"name" minLength:"1" maxLength:"250"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type accessGroupPatch struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
}

type accessUserCreate struct {
	Username string `json:"username" valid:"length(3|250),username" minLength:"3" maxLength:"250"`
	Password string `json:"password" valid:"bcrypt_password" minLength:"8" maxLength:"72"`
}

type accessUserPatch struct {
	Name     *string      `json:"name,omitempty"`
	Email    *string      `json:"email,omitempty"`
	Phone    *string      `json:"phone,omitempty"`
	Notes    *string      `json:"notes,omitempty"`
	GroupIDs *[]int64     `json:"group_ids,omitempty"`
	IsAdmin  *bool        `json:"is_admin,omitempty"`
	Status   *user.Status `json:"status,omitempty"`
}

func RegisterAccessControlRoutes(api huma.API) {
	tags := []string{"access-control"}

	Register(api, huma.Operation{
		OperationID: "access-me", Summary: "Get effective global feature permissions",
		Method: http.MethodGet, Path: "/access/me", Tags: tags,
	}, accessMe)

	// Everything below /admin is already protected by gateV2AdminRoutes.
	Register(api, huma.Operation{
		OperationID: "admin-access-permissions", Summary: "List assignable global permissions",
		Method: http.MethodGet, Path: "/admin/access/permissions", Tags: tags,
	}, adminAccessPermissions)
	Register(api, huma.Operation{
		OperationID: "admin-access-groups", Summary: "List access groups",
		Method: http.MethodGet, Path: "/admin/access/groups", Tags: tags,
	}, adminAccessGroups)
	Register(api, huma.Operation{
		OperationID: "admin-access-groups-create", Summary: "Create an access group",
		Method: http.MethodPost, Path: "/admin/access/groups", Tags: tags,
	}, adminAccessGroupsCreate)
	Register(api, huma.Operation{
		OperationID: "admin-access-groups-update", Summary: "Update an access group",
		Method: http.MethodPatch, Path: "/admin/access/groups/{id}", Tags: tags,
	}, adminAccessGroupsUpdate)
	Register(api, huma.Operation{
		OperationID: "admin-access-groups-delete", Summary: "Delete an access group",
		Method: http.MethodDelete, Path: "/admin/access/groups/{id}", Tags: tags,
	}, adminAccessGroupsDelete)

	Register(api, huma.Operation{
		OperationID: "admin-access-users", Summary: "List users with personnel cards and groups",
		Method: http.MethodGet, Path: "/admin/access/users", Tags: tags,
	}, adminAccessUsers)
	Register(api, huma.Operation{
		OperationID: "admin-access-users-create", Summary: "Create a user with personnel card and groups",
		Method: http.MethodPost, Path: "/admin/access/users", Tags: tags,
	}, adminAccessUsersCreate)
	Register(api, huma.Operation{
		OperationID: "admin-access-users-update", Summary: "Update a user's personnel card and groups",
		Method: http.MethodPatch, Path: "/admin/access/users/{id}", Tags: tags,
	}, adminAccessUsersUpdate)
}

func init() { AddRouteRegistrar(RegisterAccessControlRoutes) }

func accessUserFromCtx(ctx context.Context) (*user.User, error) {
	auth, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return user.GetFromAuth(auth)
}

func requireAccessPermission(ctx context.Context, permission string) (*user.User, error) {
	u, err := accessUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	ok, err := models.UserHasAccessPermission(s, u, permission)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if !ok {
		return nil, huma.Error403Forbidden("You do not have permission to use this function")
	}
	return u, nil
}

func accessMe(ctx context.Context, _ *struct{}) (*accessMeBody, error) {
	u, err := accessUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()
	permissions, err := models.UserAccessPermissions(s, u)
	if err != nil {
		return nil, translateDomainError(err)
	}
	groups, err := models.UserAccessGroups(s, u.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &accessMeBody{}
	out.Body.Permissions = permissions
	out.Body.Groups = groups
	return out, nil
}

func adminAccessPermissions(_ context.Context, _ *struct{}) (*accessPermissionCatalogBody, error) {
	out := &accessPermissionCatalogBody{}
	out.Body.Permissions = models.AccessPermissionDefinitions()
	return out, nil
}

func adminAccessGroups(_ context.Context, _ *struct{}) (*accessGroupsBody, error) {
	s := db.NewSession()
	defer s.Close()
	groups, err := models.ListAccessGroups(s)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &accessGroupsBody{}
	out.Body.Groups = groups
	return out, nil
}

func adminAccessGroupsCreate(_ context.Context, in *struct{ Body accessGroupCreate }) (*accessGroupBody, error) {
	s := db.NewSession()
	defer s.Close()
	group, err := models.CreateAccessGroup(s, in.Body.Name, in.Body.Description, in.Body.Permissions)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &accessGroupBody{Body: *group}, nil
}

func adminAccessGroupsUpdate(_ context.Context, in *struct {
	ID   int64 `path:"id" minimum:"1"`
	Body accessGroupPatch
}) (*accessGroupBody, error) {
	s := db.NewSession()
	defer s.Close()
	group, err := models.UpdateAccessGroup(s, in.ID, in.Body.Name, in.Body.Description, in.Body.Permissions)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &accessGroupBody{Body: *group}, nil
}

func adminAccessGroupsDelete(_ context.Context, in *struct {
	ID int64 `path:"id" minimum:"1"`
}) (*emptyBody, error) {
	s := db.NewSession()
	defer s.Close()
	if err := models.DeleteAccessGroup(s, in.ID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func adminAccessUsers(_ context.Context, in *struct {
	Search string `query:"search"`
}) (*accessUsersBody, error) {
	s := db.NewSession()
	defer s.Close()
	users, err := models.ListAccessUsers(s, in.Search)
	if err != nil {
		return nil, translateDomainError(err)
	}
	out := &accessUsersBody{}
	out.Body.Users = users
	return out, nil
}

func adminAccessUsersCreate(ctx context.Context, in *struct{ Body accessUserCreate }) (*accessUserBody, error) {
	doer, err := adminDoerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	// CreateUserAsAdmin owns its transaction/commit to stay compatible with the
	// existing admin user-management code.
	s := db.NewSession()
	defer s.Close()
	created, err := models.CreateUserAsAdmin(s, doer, &models.CreateUserBody{
		APIUserPassword: user.APIUserPassword{
			Username: in.Body.Username,
			Password: in.Body.Password,
		},
		SkipEmailConfirm: true,
	})
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)

	read := db.NewSession()
	defer read.Close()
	view, err := models.AccessUser(read, created)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &accessUserBody{Body: *view}, nil
}

func adminAccessUsersUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id" minimum:"1"`
	Body accessUserPatch
}) (*accessUserBody, error) {
	doer, err := adminDoerFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	target := &user.User{ID: in.ID}
	has, err := s.Get(target)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if !has {
		return nil, translateDomainError(user.ErrUserDoesNotExist{UserID: in.ID})
	}

	if in.Body.IsAdmin != nil && target.IsAdmin != *in.Body.IsAdmin {
		target, err = models.SetUserAdminFlag(s, doer, target.ID, *in.Body.IsAdmin)
		if err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
	}
	if in.Body.Status != nil && target.Status != *in.Body.Status {
		if *in.Body.Status < user.StatusActive || *in.Body.Status > user.StatusAccountLocked {
			_ = s.Rollback()
			return nil, huma.Error422UnprocessableEntity("invalid user status")
		}
		target, err = models.SetUserStatusAsAdmin(s, doer, target.ID, *in.Body.Status)
		if err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
	}

	cols := make([]string, 0, 2)
	update := &user.User{ID: target.ID}
	if in.Body.Name != nil {
		update.Name = strings.TrimSpace(*in.Body.Name)
		cols = append(cols, "name")
	}
	if in.Body.Email != nil {
		update.Email = strings.TrimSpace(*in.Body.Email)
		if update.Email == "" {
			_ = s.Rollback()
			return nil, huma.Error422UnprocessableEntity("email is required")
		}
		cols = append(cols, "email")
	}
	if len(cols) > 0 {
		if _, err := s.ID(target.ID).Cols(cols...).Update(update); err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
	}

	profile, err := models.GetUserProfile(s, target.ID)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if in.Body.Phone != nil {
		profile.Phone = *in.Body.Phone
	}
	if in.Body.Notes != nil {
		profile.Notes = *in.Body.Notes
	}
	if in.Body.Phone != nil || in.Body.Notes != nil {
		if err := models.SetUserProfile(s, target.ID, profile.Phone, profile.Notes); err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
	}

	if in.Body.GroupIDs != nil {
		if err := models.SetUserAccessGroups(s, target, *in.Body.GroupIDs); err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
	}

	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)

	read := db.NewSession()
	defer read.Close()
	fresh := &user.User{ID: target.ID}
	if _, err := read.Get(fresh); err != nil {
		return nil, translateDomainError(err)
	}
	view, err := models.AccessUser(read, fresh)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &accessUserBody{Body: *view}, nil
}
