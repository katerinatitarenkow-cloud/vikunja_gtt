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

func TestTaskChecklistV2(t *testing.T) {
	t.Run("create, audit completion and auto-complete parent", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1/checklist-items", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"total":0`)

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/checklist-items", `{"title":"Check inverter"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"total":1`)
		assert.Contains(t, rec.Body.String(), `"task_done":false`)

		rec = humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/checklist-items/1", `{"title":"Check inverter","done":true}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"completed":1`)
		assert.Contains(t, rec.Body.String(), `"task_done":true`)
		assert.Contains(t, rec.Body.String(), `"completed_by"`)
		assert.Contains(t, rec.Body.String(), `"id":1`)

		db.AssertExists(t, "task_checklist_items", map[string]interface{}{
			"id":              1,
			"task_id":         1,
			"done":            true,
			"completed_by_id": 1,
		}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": true,
		}, false)
	})

	t.Run("adding pending work reopens completed parent", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/checklist-items", `{"title":"First"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		rec = humaRequest(t, e, http.MethodPut, "/api/v2/tasks/1/checklist-items/1", `{"title":"First","done":true}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"task_done":true`)

		rec = humaRequest(t, e, http.MethodPost, "/api/v2/tasks/1/checklist-items", `{"title":"Second"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"total":2`)
		assert.Contains(t, rec.Body.String(), `"completed":1`)
		assert.Contains(t, rec.Body.String(), `"task_done":false`)
	})

	t.Run("write requires task write permission", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// task 15 is read-only for user1 in the stock webtest fixture set.
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/15/checklist-items", `{"title":"Forbidden"}`, token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("read requires task read permission", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// task 41 belongs to a project user1 cannot access.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/41/checklist-items", "", token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
