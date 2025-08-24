package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/glimesh/broadcast-box/internal/environment"
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
			MOTD:      "Welcome to " + token + "'s stream!",
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

	// Allow anyone to use streamkey has not been reserved
	case authorization.STREAM_POLICY_WITH_RESERVED:
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

	// TODO: Remove this selection, as it is the same as WITH_RESERVED
	// Anyone can stream
	case authorization.STREAM_POLICY_ANYONE:
		log.Println("Policy:", authorization.STREAM_POLICY_ANYONE)
	default:
		log.Println("Policy: None")
	}

	// Set default profile in case none is set
	if userProfile == (authorization.PublicProfile{}) {
		userProfile = authorization.PublicProfile{
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
