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

func TestClientCustomFieldsV2(t *testing.T) {
	t.Run("create update list delete and history", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/client/custom-fields", `{"name":"Мощность СЭС","value":"150 кВт"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"name":"Мощность СЭС"`)
		assert.Contains(t, rec.Body.String(), `"value":"150 кВт"`)

		db.AssertExists(t, "client_custom_fields", map[string]interface{}{
			"project_id": 1,
			"name":       "Мощность СЭС",
			"value":      "150 кВт",
		}, false)
		db.AssertExists(t, "client_activity_events", map[string]interface{}{
			"project_id":       1,
			"event_type":       "custom_field_created",
			"entity_type":      "custom_field",
			"system_generated": true,
		}, false)

		rec = humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/client/custom-fields", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"name":"Мощность СЭС"`)

		rec = humaRequest(t, e, http.MethodPut, "/api/v2/projects/1/client/custom-fields/1", `{"name":"Мощность объекта","value":"180 кВт"}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"name":"Мощность объекта"`)
		assert.Contains(t, rec.Body.String(), `"value":"180 кВт"`)
		db.AssertExists(t, "client_activity_events", map[string]interface{}{
			"project_id": 1,
			"event_type": "custom_field_updated",
		}, false)

		rec = humaRequest(t, e, http.MethodDelete, "/api/v2/projects/1/client/custom-fields/1", "", token, "")
		require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
		db.AssertMissing(t, "client_custom_fields", map[string]interface{}{
			"id": 1,
		})
		db.AssertExists(t, "client_activity_events", map[string]interface{}{
			"project_id": 1,
			"event_type": "custom_field_deleted",
		}, false)
	})

	t.Run("private client custom fields stay private", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/20/client/custom-fields", "", token, "")
		require.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})
}
