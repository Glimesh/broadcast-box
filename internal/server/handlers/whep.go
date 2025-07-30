package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/webrtc"
)

func whepHandler(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("WhepHandler called")
	if request.Method == "DELETE" {
		return
	}

	authHeader := request.Header.Get("Authorization")

	if authHeader == "" {
		log.Println("Authorization was not set")
		helpers.LogHttpError(responseWriter, "Authorization was not set", http.StatusBadRequest)
	}

	token := helpers.ResolveBearerToken(authHeader)
	if token == "" {
		log.Println("Authorization was invalid")
		helpers.LogHttpError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)
		return
	}

	offer, err := io.ReadAll(request.Body)
	if err != nil {
		log.Println(err.Error())
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	whipAnswer, sessionId, err := webrtc.WHEP(string(offer), token)
	if err != nil {
		log.Println("WHEP Error", err.Error())
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Add("Link", `<`+"/api/sse/"+sessionId+`>; rel="urn:ietf:params:whep:ext:core:server-sent-events"; events="layers"`)
	responseWriter.Header().Add("Link", `<`+"/api/layer/"+sessionId+`>; rel="urn:ietf:params:whep:ext:core:layer"`)

	responseWriter.Header().Add("Location", "/api/whep")
	responseWriter.Header().Add("Content-Type", "application/sdp")
	responseWriter.WriteHeader(http.StatusCreated)

	if _, err = fmt.Fprint(responseWriter, whipAnswer); err != nil {
		log.Println("Error", err)
	} else {
		log.Println("API.WHEP Completed")
	}
}
