package handlers

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

func WhepHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" && request.Method != "PATCH" {
		return
	}

	offer, err := io.ReadAll(request.Body)
	if err != nil || string(offer) == "" {
		helpers.LogHttpError(responseWriter, "error reading offer", http.StatusBadRequest)
		return
	}

	if request.Method == "PATCH" {
		if err := utils.ValidateOffer(string(offer)); err != nil {
			helpers.LogHttpError(responseWriter, "invalid offer: "+err.Error(), http.StatusBadRequest)
			return
		}

		path := strings.Replace(request.URL.Path, "/api/whep", "", 1)
		segments := strings.Split(path, "/")
		sessionId := strings.TrimSpace(segments[len(segments)-1])

		if sessionId == "" {
			log.Println("API.WHEP.Patch Error: Missing session id")
			helpers.LogHttpError(responseWriter, "Missing session id", http.StatusBadRequest)
			return
		}

		log.Println("API.WHEP.Patch: Patching session", sessionId)
		if err := patchHandler(responseWriter, request, sessionId, string(offer)); err != nil {
			log.Println("API.WHEP.Patch Error:", err)
			helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		}

		return
	}

	token := helpers.ResolveBearerToken(request.Header.Get("Authorization"))
	if token == "" {
		helpers.LogHttpError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)
		return
	}

	whipAnswer, sessionId, err := webrtc.WHEP(string(offer), token)
	if err != nil {
		log.Println("API.WHEP: Setup Error", err.Error())
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Add("Link", `<`+"/api/sse/"+sessionId+`>; rel="urn:ietf:params:whep:ext:core:server-sent-events"; events="layers"`)
	responseWriter.Header().Add("Link", `<`+"/api/layer/"+sessionId+`>; rel="urn:ietf:params:whep:ext:core:layer"`)

	responseWriter.Header().Add("Location", "/api/whep/"+sessionId)
	responseWriter.Header().Add("Content-Type", "application/sdp")
	responseWriter.WriteHeader(http.StatusCreated)

	if _, err = fmt.Fprint(responseWriter, whipAnswer); err != nil {
		log.Println("API.WHEP:", err)
	} else {
		log.Println("API.WHEP: Completed")
	}
}

func patchHandler(res http.ResponseWriter, r *http.Request, sessionId, body string) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/trickle-ice-sdpfrag" {
		helpers.LogHttpError(res, "invalid content type", http.StatusUnsupportedMediaType)
		return err
	}

	if err = webrtc.HandleWhepPatch(sessionId, body); err != nil {
		return err
	}

	res.WriteHeader(http.StatusNoContent)

	return nil
}
