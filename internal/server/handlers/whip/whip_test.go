package whip

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/stretchr/testify/require"
)

type whipWebhookPayload struct {
	Action      string `json:"action"`
	BearerToken string `json:"bearerToken"`
}

func TestWHIPHandlerWebhookUsesResolvedStreamKey(t *testing.T) {
	t.Setenv(environment.StreamProfilePath, t.TempDir())
	t.Setenv(environment.StreamProfilePolicy, authorization.StreamPolicyReservedOnly)

	const streamKey = "test_stream"
	bearerToken, err := authorization.CreateProfile(streamKey)
	require.NoError(t, err)

	payloads := make(chan whipWebhookPayload, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()

		var payload whipWebhookPayload
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))

		payloads <- payload
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	t.Setenv(environment.WebhookURL, server.URL)

	req := httptest.NewRequest(http.MethodPost, "/api/whip", strings.NewReader("v=0"))
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp := httptest.NewRecorder()
	WHIPHandler(resp, req)

	require.Equal(t, http.StatusUnauthorized, resp.Code)

	select {
	case payload := <-payloads:
		require.Equal(t, "whip-connect", payload.Action)
		require.Equal(t, streamKey, payload.BearerToken)
	default:
		require.Fail(t, "expected webhook to be called")
	}
}
