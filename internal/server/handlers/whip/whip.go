package whip

import (
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/server/webhook"
	"github.com/glimesh/broadcast-box/internal/webrtc"
)

func WHIPHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost && request.Method != http.MethodPatch && request.Method != http.MethodDelete {
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := request.Header.Get("Authorization")

	if authHeader == "" {
		slog.Info("Authorization was not set")
		helpers.LogHTTPError(responseWriter, "Authorization was not set", http.StatusBadRequest)
		return
	}

	token := helpers.ResolveBearerToken(authHeader)
	if token == "" {
		slog.Info("Authorization was invalid")
		helpers.LogHTTPError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)
		return
	}

	if request.Method == http.MethodDelete {
		sessionID := getSessionIDFromWHIPPath(request.URL.Path)

		if sessionID == "" {
			slog.Info("API.WHIP.Delete Error: Missing session id")
			helpers.LogHTTPError(responseWriter, "Missing session id", http.StatusBadRequest)
			return
		}

		slog.Info("API.WHIP.Delete: Removing session", "sessionID", sessionID)
		if err := deleteHandler(responseWriter, sessionID); err != nil {
			slog.Error("API.WHIP.Delete Error", "err", err)
			helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		}

		return
	}

	offer, err := io.ReadAll(request.Body)
	if err != nil || string(offer) == "" {
		slog.Info("Error reading offer")
		helpers.LogHTTPError(responseWriter, "error reading offer", http.StatusBadRequest)
		return
	}

	var userProfile authorization.PublicProfile

	// Stream profile policy
	switch os.Getenv(environment.StreamProfilePolicy) {
	// Only approved profiles are allowed to stream
	case authorization.StreamPolicyReservedOnly:
		slog.Info("Stream Policy Selected", "policy", authorization.StreamPolicyReservedOnly)
		profile, err := authorization.GetPublicProfile(token)
		if err != nil {
			slog.Info("Unauthorized login attempt", "token", token)
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}
		userProfile = *profile

	default:
		slog.Info("Stream Policy Selected", "policy", authorization.StreamPolicyWithReserved)

		// If using a streamKey check if it has been reserved
		if authorization.IsProfileReserved(token) {
			slog.Info("Unauthorized login attempt with reserved Streamkey", "token", token)
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}

		// If its a bearer token, validate and use the profile
		profile, _ := authorization.GetPublicProfile(token)
		if profile != nil {
			userProfile = *profile
		}
	}

	// Set default profile in case none is set
	if userProfile == (authorization.PublicProfile{}) {
		userProfile = authorization.PublicProfile{
			StreamKey: token,
			IsPublic:  true,
			MOTD:      "Welcome to " + token + "'s stream!",
		}
	}

	// Stream requires webhook validation
	if webhookURL := os.Getenv(environment.WebhookURL); webhookURL != "" {
		streamKey, err := webhook.CallWebhook(webhookURL, webhook.WHIPConnect, userProfile.StreamKey, request)
		if err != nil {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}

		userProfile = authorization.PublicProfile{
			StreamKey: streamKey,
			IsPublic:  true,
			MOTD:      "Welcome to " + streamKey + "'s stream!",
		}
	}

	if request.Method == http.MethodPatch {

		if contentType := request.Header.Get("Content-Type"); contentType != "application/trickle-ice-sdpfrag" {
			slog.Info("API.WHIP.Patch Error: Invalid patch request")
			helpers.LogHTTPError(responseWriter, "Invalid patch request", http.StatusBadRequest)
			return
		}

		sessionID := getSessionIDFromWHIPPath(request.URL.Path)

		if sessionID == "" {
			slog.Info("API.WHIP.Patch Error: Missing session id")
			helpers.LogHTTPError(responseWriter, "Missing session id", http.StatusBadRequest)
			return
		}

		slog.Info("API.WHIP.Patch: Patching session", "sessionID", sessionID)
		if err := patchHandler(responseWriter, request, sessionID, string(offer)); err != nil {
			slog.Error("API.WHIP.Patch Error:", "err", err)
			helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		}

		return
	}

	whipAnswer, sessionID, err := webrtc.WHIP(string(offer), userProfile)
	if err != nil {
		helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Add("Link", `<`+"/api/sse/"+sessionID+`>; rel="urn:ietf:params:whep:ext:core:server-sent-events"; events="status"`)
	responseWriter.Header().Add("Location", "/api/whip/"+sessionID)
	responseWriter.Header().Add("Content-Type", "application/sdp")
	responseWriter.WriteHeader(http.StatusCreated)

	if _, err = fmt.Fprint(responseWriter, whipAnswer); err != nil {
		slog.Error("API.WHIP Error", "err", err)
	} else {
		slog.Info("API.WHIP Completed")
	}

}

func patchHandler(res http.ResponseWriter, r *http.Request, sessionID, body string) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/trickle-ice-sdpfrag" {
		helpers.LogHTTPError(res, "invalid content type", http.StatusUnsupportedMediaType)
		return err
	}

	if err = webrtc.HandleWHIPPatch(sessionID, body); err != nil {
		return err
	}

	res.WriteHeader(http.StatusNoContent)

	return nil
}

func deleteHandler(res http.ResponseWriter, sessionID string) error {
	if err := webrtc.HandleWHIPDelete(sessionID); err != nil {
		return err
	}

	res.WriteHeader(http.StatusNoContent)

	return nil
}

func getSessionIDFromWHIPPath(path string) string {
	path = strings.Replace(path, "/api/whip", "", 1)
	segments := strings.Split(path, "/")
	return strings.TrimSpace(segments[len(segments)-1])
}
