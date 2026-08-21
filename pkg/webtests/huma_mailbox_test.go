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

func TestMailboxV2(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	token1 := humaTokenFor(t, &testuser1)
	token2 := humaTokenFor(t, &testuser2)
	token10 := humaTokenFor(t, &testuser10)

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/mailbox/messages", `{"recipient_id":2,"subject":"Hello","body":"Private body"}`, token1, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"sender_id":1`)
	assert.Contains(t, rec.Body.String(), `"recipient_id":2`)
	db.AssertExists(t, "user_mailbox_messages", map[string]interface{}{
		"sender_id":    1,
		"recipient_id": 2,
		"subject":      "Hello",
	}, false)

	t.Run("recipient sees inbox", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/mailbox/messages?folder=inbox", "", token2, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"subject":"Hello"`)
	})

	t.Run("sender sees sent", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/mailbox/messages?folder=sent", "", token1, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"subject":"Hello"`)
	})

	t.Run("third user cannot read", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/mailbox/messages/1", "", token10, "")
		require.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("recipient marks read", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v2/mailbox/messages/1/read", `{"read":true}`, token2, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotContains(t, rec.Body.String(), `"read_at":"0001-01-01`)
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/mailbox/unread-count", "", token2, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"count":0`)
	})

	t.Run("sender delete keeps recipient copy", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodDelete, "/api/v2/mailbox/messages/1", "", token1, "")
		require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/mailbox/messages/1", "", token2, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	})
}
