// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package webtests

import (
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientActivityV2(t *testing.T) {
	t.Run("manual activity create list and delete", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		body := `{"event_type":"call","occurred_at":"2026-08-20T09:31:00Z","title":"Расчёт СЭС","description":"Нужен расчёт СЭС 150 кВт.","metadata":{"direction":"incoming","duration_minutes":4,"result":"Ждёт расчёт"}}`
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/client/history", body, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"event_type":"call"`)
		assert.Contains(t, rec.Body.String(), `"system_generated":false`)

		db.AssertExists(t, "client_activity_events", map[string]interface{}{
			"project_id":       1,
			"event_type":       "call",
			"actor_user_id":    1,
			"system_generated": false,
		}, false)

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/client/history", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"Расчёт СЭС"`)
		assert.Contains(t, rec.Body.String(), `"duration_minutes":4`)

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/projects/1/client/history/1", "", token, "")
		require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("task creation is written automatically", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/tasks", `{"title":"Получить технические условия"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/client/history?type=task_created", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"event_type":"task_created"`)
		assert.Contains(t, rec.Body.String(), `"task_title":"Получить технические условия"`)
		assert.Contains(t, rec.Body.String(), `"system_generated":true`)
	})

	t.Run("private project history stays private", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/20/client/history", "", token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
