// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package routes

import (
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// requiredGlobalFeaturePermission maps broad API areas to the new global RBAC
// layer. Existing per-project permissions are still evaluated by the handlers.
func requiredGlobalFeaturePermission(path, method string) string {
	if strings.HasPrefix(path, "/api/v2/admin/") || strings.HasSuffix(path, "/access/me") {
		return ""
	}
	read := method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
	trimmed := strings.TrimPrefix(strings.TrimPrefix(path, "/api/v1"), "/api/v2")

	if strings.HasPrefix(trimmed, "/wialon") {
		return models.PermissionWialonView
	}
	if strings.Contains(trimmed, "/time-entries") {
		return models.PermissionTimeTracking
	}
	if strings.Contains(trimmed, "/buckets") {
		return models.PermissionKanbanUse
	}

	// Nested task APIs live below /projects as well, so task detection comes first.
	if strings.Contains(trimmed, "/tasks") || strings.HasPrefix(trimmed, "/tasks") {
		if read {
			return models.PermissionTasksView
		}
		return models.PermissionTasksManage
	}
	if strings.HasPrefix(trimmed, "/projects") {
		if read {
			return models.PermissionProjectsView
		}
		return models.PermissionProjectsManage
	}
	if strings.HasPrefix(trimmed, "/labels") {
		if read {
			return models.PermissionLabelsView
		}
		return models.PermissionLabelsManage
	}
	if strings.HasPrefix(trimmed, "/teams") {
		if read {
			return models.PermissionTeamsView
		}
		return models.PermissionTeamsManage
	}
	return ""
}

// gateGlobalFeaturePermissions adds coarse feature permissions on top of the
// detailed Vikunja project/task permission model. Link shares are intentionally
// left to the existing share permission checks.
func gateGlobalFeaturePermissions() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			permission := requiredGlobalFeaturePermission(c.Request().URL.Path, c.Request().Method)
			if permission == "" {
				return next(c)
			}

			auth, err := auth2.GetAuthFromClaims(c)
			if err != nil {
				return next(c)
			}
			u, ok := auth.(*user.User)
			if !ok {
				return next(c)
			}

			s := db.NewSession()
			allowed, err := models.UserHasAccessPermission(s, u, permission)
			_ = s.Close()
			if err != nil {
				return err
			}
			if !allowed {
				return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to use this function")
			}
			return next(c)
		}
	}
}
