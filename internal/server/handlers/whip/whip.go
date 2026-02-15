package whip

import (
	"fmt"
	"io"
	"log"
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

func WhipHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost && request.Method != http.MethodPatch && request.Method != http.MethodDelete {
		helpers.LogHttpError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := request.Header.Get("Authorization")

	if authHeader == "" {
		log.Println("Authorization was not set")
		helpers.LogHttpError(responseWriter, "Authorization was not set", http.StatusBadRequest)
		return
	}

	token := helpers.ResolveBearerToken(authHeader)
	if token == "" {
		log.Println("Authorization was invalid")
		helpers.LogHttpError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)
		return
	}

	if request.Method == http.MethodDelete {
		sessionId := getSessionIdFromWhipPath(request.URL.Path)

		if sessionId == "" {
			log.Println("API.WHIP.Delete Error: Missing session id")
			helpers.LogHttpError(responseWriter, "Missing session id", http.StatusBadRequest)
			return
		}

		log.Println("API.WHIP.Delete: Removing session", sessionId)
		if err := deleteHandler(responseWriter, sessionId); err != nil {
			log.Println("API.WHIP.Delete Error:", err)
			helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		}

		return
	}

	offer, err := io.ReadAll(request.Body)
	if err != nil || string(offer) == "" {
		log.Println("Error reading offer")
		helpers.LogHttpError(responseWriter, "error reading offer", http.StatusBadRequest)
		return
	}

	var userProfile authorization.PublicProfile

	// Stream requires webhook validation
	if webhookUrl := os.Getenv(environment.WEBHOOK_URL); webhookUrl != "" {
		streamKey, err := webhook.CallWebhook(webhookUrl, webhook.WhipConnect, token, request)
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

	// Stream profile policy
	switch os.Getenv(environment.STREAM_PROFILE_POLICY) {
	// Only approved profiles are allowed to stream
	case authorization.STREAM_POLICY_RESERVED_ONLY:
		log.Println("Policy:", authorization.STREAM_POLICY_RESERVED_ONLY)
		profile, err := authorization.GetPublicProfile(token)
		if err != nil {
			log.Println("Unauthorized login attempt with bearer", token)
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}
		userProfile = *profile

	default:
		log.Println("Policy:", authorization.STREAM_POLICY_WITH_RESERVED)

		// If using a streamKey check if it has been reserved
		if authorization.IsProfileReserved(token) {
			log.Println("Unauthorized login attempt with bearer", token, " - Streamkey has been reserved")
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

	if request.Method == http.MethodPatch {

		if contentType := request.Header.Get("Content-Type"); contentType != "application/trickle-ice-sdpfrag" {
			log.Println("API.WHIP.Patch Error: Invalid patch request")
			helpers.LogHttpError(responseWriter, "Invalid patch request", http.StatusBadRequest)
			return
		}

		sessionId := getSessionIdFromWhipPath(request.URL.Path)

		if sessionId == "" {
			log.Println("API.WHIP.Patch Error: Missing session id")
			helpers.LogHttpError(responseWriter, "Missing session id", http.StatusBadRequest)
			return
		}

		log.Println("API.WHIP.Patch: Patching session", sessionId)
		if err := patchHandler(responseWriter, request, sessionId, string(offer)); err != nil {
			log.Println("API.WHIP.Patch Error:", err)
			helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		}

		return
	}

	whipAnswer, sessionId, err := webrtc.WHIP(string(offer), userProfile)
	if err != nil {
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Add("Link", `<`+"/api/sse/"+sessionId+`>; rel="urn:ietf:params:whep:ext:core:server-sent-events"; events="status"`)
	responseWriter.Header().Add("Location", "/api/whip/"+sessionId)
	responseWriter.Header().Add("Content-Type", "application/sdp")
	responseWriter.WriteHeader(http.StatusCreated)

	if _, err = fmt.Fprint(responseWriter, whipAnswer); err != nil {
		log.Println("API.WHIP Error", err)
	} else {
		log.Println("API.WHIP Completed")
	}

}

func patchHandler(res http.ResponseWriter, r *http.Request, sessionId, body string) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/trickle-ice-sdpfrag" {
		helpers.LogHttpError(res, "invalid content type", http.StatusUnsupportedMediaType)
		return err
	}

	if err = webrtc.HandleWhipPatch(sessionId, body); err != nil {
		return err
	}

	res.WriteHeader(http.StatusNoContent)

	return nil
}

func deleteHandler(res http.ResponseWriter, sessionId string) error {
	if err := webrtc.HandleWhipDelete(sessionId); err != nil {
		return err
	}

	res.WriteHeader(http.StatusNoContent)

	return nil
}

func getSessionIdFromWhipPath(path string) string {
	path = strings.Replace(path, "/api/whip", "", 1)
	segments := strings.Split(path, "/")
	return strings.TrimSpace(segments[len(segments)-1])
}
