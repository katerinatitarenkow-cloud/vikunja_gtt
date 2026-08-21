// Copyright 2018-present Vikunja and contributors. All rights reserved.

package webtests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func setGlobalPermissions(t *testing.T, u *user.User, permissions ...string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	require.NoError(t, models.SetUserAccessGroups(s, u, nil))
	if len(permissions) > 0 {
		group, err := models.CreateAccessGroup(s, "rbac-test-"+u.Username, "", permissions)
		require.NoError(t, err)
		require.NoError(t, models.SetUserAccessGroups(s, u, []int64{group.ID}))
	}
	require.NoError(t, s.Commit())
}

func TestGlobalProjectAccess(t *testing.T) {
	t.Run("view reads unshared project and CRM but cannot manage", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		setGlobalPermissions(t, &testuser1, models.PermissionProjectsView)
		token := humaTokenFor(t, &testuser1)

		for _, path := range []string{
			"/api/v1/projects/20",
			"/api/v2/projects",
			"/api/v2/projects/20",
			"/api/v2/projects/20/client",
			"/api/v2/projects/20/client/history",
		} {
			res := humaRequest(t, e, http.MethodGet, path, "", token, "")
			require.Equal(t, http.StatusOK, res.Code, "%s: %s", path, res.Body.String())
		}

		for method, path := range map[string]string{
			http.MethodPost:   "/api/v2/projects",
			http.MethodPut:    "/api/v2/projects/20",
			http.MethodDelete: "/api/v2/projects/20",
		} {
			res := humaRequest(t, e, method, path, `{"title":"blocked"}`, token, "")
			require.Equal(t, http.StatusForbidden, res.Code, "%s %s: %s", method, path, res.Body.String())
		}
	})

	for _, groupCount := range []int{2, 3} {
		t.Run(fmt.Sprintf("project list stays unique with %d access groups", groupCount), func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			s := db.NewSession()
			defer s.Close()
			groupIDs := make([]int64, 0, groupCount)
			for i := 0; i < groupCount; i++ {
				group, err := models.CreateAccessGroup(s, fmt.Sprintf("project-list-%d-%d", groupCount, i), "", []string{models.PermissionProjectsView})
				require.NoError(t, err)
				groupIDs = append(groupIDs, group.ID)
			}
			require.NoError(t, models.SetUserAccessGroups(s, &testuser1, groupIDs))
			require.NoError(t, s.Commit())

			res := humaRequest(t, e, http.MethodGet, "/api/v2/projects?per_page=100", "", humaTokenFor(t, &testuser1), "")
			require.Equal(t, http.StatusOK, res.Code, res.Body.String())
			var page struct {
				Items []models.Project `json:"items"`
			}
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &page))
			seen := make(map[int64]struct{}, len(page.Items))
			for _, project := range page.Items {
				_, duplicate := seen[project.ID]
				require.False(t, duplicate, "project id %d occurred more than once", project.ID)
				seen[project.ID] = struct{}{}
			}
			require.Contains(t, seen, int64(20))
		})
	}

	t.Run("without view project reads are forbidden", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		setGlobalPermissions(t, &testuser1)
		token := humaTokenFor(t, &testuser1)
		for _, path := range []string{"/api/v2/projects", "/api/v2/projects/1"} {
			res := humaRequest(t, e, http.MethodGet, path, "", token, "")
			require.Equal(t, http.StatusForbidden, res.Code, "%s: %s", path, res.Body.String())
		}
	})

	t.Run("project CRM does not require tasks view", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		setGlobalPermissions(t, &testuser1, models.PermissionProjectsView)
		token := humaTokenFor(t, &testuser1)

		client := humaRequest(t, e, http.MethodGet, "/api/v2/projects/20/client", "", token, "")
		require.Equal(t, http.StatusOK, client.Code, client.Body.String())
		tasks := humaRequest(t, e, http.MethodGet, "/api/v2/projects/20/tasks", "", token, "")
		require.Equal(t, http.StatusForbidden, tasks.Code, tasks.Body.String())
	})
}
