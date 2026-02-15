package handlers

import (
	"encoding/base64"
	"encoding/json"
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

type WhepRequest struct {
	Offer     string
	StreamKey string
}

func WhepHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" && request.Method != "PATCH" {
		return
	}

	requestBodyB64, err := io.ReadAll(request.Body)
	if err != nil {
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: Is decodedB64 neccesarry?
	decodedB64, err := base64.StdEncoding.DecodeString(string(requestBodyB64))
	if err != nil {
		log.Println("API.WHEP: Invalid B64 encoding for request")
		return
	}

	if request.Method == "PATCH" {
		if err := utils.ValidateOffer(string(decodedB64)); err != nil {
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
		if err := patchHandler(responseWriter, request, sessionId, string(decodedB64)); err != nil {
			log.Println("API.WHEP.Patch Error:", err)
			helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		}

		return
	}

	var whepRequest WhepRequest
	if err := json.Unmarshal(decodedB64, &whepRequest); err != nil {
		log.Println("API.WHEP: Could not read WHEP request")
		return
	}

	whipAnswer, sessionId, err := webrtc.WHEP(string(whepRequest.Offer), whepRequest.StreamKey)
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
