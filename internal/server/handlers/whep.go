package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc"
)

type WhepRequest struct {
	Offer     string
	StreamKey string
}

func WhepHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "DELETE" {
		return
	}

	requestBodyB64, err := io.ReadAll(request.Body)
	if err != nil {
		log.Println(err.Error())
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	decodedB64, err := base64.StdEncoding.DecodeString(string(requestBodyB64))
	if err != nil {
		log.Println("API.WHEP: Invalid B64 encoding for request")
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

	responseWriter.Header().Add("Location", "/api/whep")
	responseWriter.Header().Add("Content-Type", "application/sdp")
	responseWriter.WriteHeader(http.StatusCreated)

	if _, err = fmt.Fprint(responseWriter, whipAnswer); err != nil {
		log.Println("API.WHEP:", err)
	} else {
		log.Println("API.WHEP: Completed")
	}
}
