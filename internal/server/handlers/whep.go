package handlers

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/server/webhook"
	"github.com/glimesh/broadcast-box/internal/webrtc"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

func whepHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost && request.Method != http.MethodPatch {
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	offer, err := io.ReadAll(request.Body)
	if err != nil || string(offer) == "" {
		helpers.LogHTTPError(responseWriter, "error reading offer", http.StatusBadRequest)
		return
	}

	if request.Method == http.MethodPatch {
		if err := utils.ValidateOffer(string(offer)); err != nil {
			helpers.LogHTTPError(responseWriter, "invalid offer: "+err.Error(), http.StatusBadRequest)
			return
		}

		path := strings.Replace(request.URL.Path, "/api/whep", "", 1)
		segments := strings.Split(path, "/")
		sessionID := strings.TrimSpace(segments[len(segments)-1])

		if sessionID == "" {
			log.Println("API.WHEP.Patch Error: Missing session id")
			helpers.LogHTTPError(responseWriter, "Missing session id", http.StatusBadRequest)
			return
		}

		log.Println("API.WHEP.Patch: Patching session", sessionID)
		if err := patchHandler(responseWriter, request, sessionID, string(offer)); err != nil {
			log.Println("API.WHEP.Patch Error:", err)
			helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		}

		return
	}

	token := helpers.ResolveBearerToken(request.Header.Get("Authorization"))
	if token == "" {
		helpers.LogHTTPError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)
		return
	}

	if webhookURL := os.Getenv(environment.WebhookURL); webhookURL != "" {
		token, err = webhook.CallWebhook(webhookURL, webhook.WHEPConnect, token, request)
		if err != nil {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	whipAnswer, sessionID, err := webrtc.WHEP(string(offer), token)
	if err != nil {
		log.Println("API.WHEP: Setup Error", err.Error())
		helpers.LogHTTPError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Add("Link", `<`+"/api/sse/"+sessionID+`>; rel="urn:ietf:params:whep:ext:core:server-sent-events"; events="layers"`)
	responseWriter.Header().Add("Link", `<`+"/api/layer/"+sessionID+`>; rel="urn:ietf:params:whep:ext:core:layer"`)

	responseWriter.Header().Add("Location", "/api/whep/"+sessionID)
	responseWriter.Header().Add("Content-Type", "application/sdp")
	responseWriter.WriteHeader(http.StatusCreated)

	if _, err = fmt.Fprint(responseWriter, whipAnswer); err != nil {
		log.Println("API.WHEP:", err)
	} else {
		log.Println("API.WHEP: Completed")
	}
}

func patchHandler(res http.ResponseWriter, r *http.Request, sessionID, body string) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/trickle-ice-sdpfrag" {
		helpers.LogHTTPError(res, "invalid content type", http.StatusUnsupportedMediaType)
		return err
	}

	if err = webrtc.HandleWHEPPatch(sessionID, body); err != nil {
		return err
	}

	res.WriteHeader(http.StatusNoContent)

	return nil
}
