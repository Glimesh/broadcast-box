package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/glimesh/broadcast-box/internal/server/webhook"
	"github.com/glimesh/broadcast-box/internal/webrtc"
)

func whipHandler(responseWriter http.ResponseWriter, request *http.Request) {
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

	var userProfile authorization.Profile

	// Stream requires webhook validation
	if webhookUrl := os.Getenv("WEBHOOK_URL"); webhookUrl != "" {
		streamKey, err := webhook.CallWebhook(webhookUrl, webhook.WhipConnect, token, request)
		if err != nil {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}

		userProfile = authorization.Profile{
			StreamKey: streamKey,
			IsPublic:  true,
			MOTD:      "Welcome to " + token + "'s stream!",
		}
	}

	// Stream requires profile
	if requiresStreamProfile := os.Getenv("STREAM_PROFILE_ACTIVE"); strings.EqualFold(requiresStreamProfile, "TRUE") {
		profile, err := authorization.GetProfile(token)
		if err != nil {
			log.Println("Unauthorized login attempt with bearer", token)
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}
		userProfile = *profile
	}

	if userProfile == (authorization.Profile{}) {
		userProfile = authorization.Profile{
			StreamKey: token,
			IsPublic:  true,
			MOTD:      "Welcome to " + token + "'s stream!",
		}
	}

	whipAnswer, sessionId, err := webrtc.WHIP(string(offer), userProfile)
	if err != nil {
		log.Println("WHIP Error", err.Error())
		helpers.LogHttpError(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	responseWriter.Header().Add("Link", `<`+"/api/sse/"+sessionId+`>; rel="urn:ietf:params:whep:ext:core:server-sent-events"; events="status"`)

	responseWriter.Header().Add("Location", "/api/whip")
	responseWriter.Header().Add("Content-Type", "application/sdp")
	responseWriter.WriteHeader(http.StatusCreated)

	if _, err = fmt.Fprint(responseWriter, whipAnswer); err != nil {
		log.Println("API.WHIP Error", err)
	} else {
		log.Println("API.WHIP Completed")
	}

}
