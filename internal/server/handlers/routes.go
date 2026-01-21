package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/glimesh/broadcast-box/internal/chat"
	"github.com/glimesh/broadcast-box/internal/environment"
	adminHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/admin"
	whipHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/whip"
)

// ChatManager is the global chat manager instance, initialized in main.go
var ChatManager *chat.Manager

func GetServeMuxHandler() http.HandlerFunc {
	serverMux := http.NewServeMux()

	if os.Getenv(environment.FrontendDisabled) == "" {
		serverMux.HandleFunc("/", frontendHandler)
	}

	// WHIP/WHEP shared endpoints
	serverMux.HandleFunc("/api/whep", corsHandler(whepHandler))
	serverMux.HandleFunc("/api/whep/", corsHandler(whepHandler))
	serverMux.HandleFunc("/api/sse/", corsHandler(sseHandler))

	// WHIP session endpoints
	serverMux.HandleFunc("/api/whip", corsHandler(whipHandlers.WHIPHandler))
	serverMux.HandleFunc("/api/whip/", corsHandler(whipHandlers.WHIPHandler))
	serverMux.HandleFunc("/api/whip/profile", corsHandler(whipHandlers.ProfileHandler))

	// WHEP session endpoints
	serverMux.HandleFunc("/api/layer/", corsHandler(layerChangeHandler))

	// Logging and status endpoints
	serverMux.HandleFunc("/api/log", corsHandler(logHandler))
	serverMux.HandleFunc("/api/status", corsHandler(statusHandler))

	// Admin endpoints
	serverMux.HandleFunc("/api/admin/login", corsHandler(adminHandlers.LoginHandler))
	serverMux.HandleFunc("/api/admin/status", corsHandler(adminHandlers.StatusHandler))
	serverMux.HandleFunc("/api/admin/logging", corsHandler(adminHandlers.LoggingHandler))
	serverMux.HandleFunc("/api/admin/profiles", corsHandler(adminHandlers.ProfilesHandler))
	serverMux.HandleFunc("/api/admin/profiles/reset-token", corsHandler(adminHandlers.ProfilesResetTokenHandler))
	serverMux.HandleFunc("/api/admin/profiles/add-profile", corsHandler(adminHandlers.ProfileAddHandler))
	serverMux.HandleFunc("/api/admin/profiles/remove-profile", corsHandler(adminHandlers.ProfileRemoveHandler))

	// Chat endpoints
	serverMux.HandleFunc("/api/chat/connect", corsHandler(chatConnectHandler))
	serverMux.HandleFunc("/api/chat/sse/", corsHandler(chatSSEHandler))
	serverMux.HandleFunc("/api/chat/send/", corsHandler(chatSendHandler))

	// Path middleware
	debugOutputWebRequests := os.Getenv(environment.DebugIncomingAPIRequest)
	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if strings.EqualFold(debugOutputWebRequests, "TRUE") {
			log.Println("Calling path", request.URL.Path)
			_, pattern := serverMux.Handler(request)

			if pattern == "" {
				log.Println("Unmatched path:", request.URL.Path)
			} else {
				log.Println("Found pattern", pattern)
			}
		}

		serverMux.ServeHTTP(responseWriter, request)
	})

	return handler
}

func RedirectToHttpsHandler(httpWriter http.ResponseWriter, request *http.Request) {
	http.Redirect(httpWriter, request, "https://"+request.Host+request.URL.String(), http.StatusMovedPermanently)
}

func corsHandler(next func(responseWriter http.ResponseWriter, request *http.Request)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Access-Control-Allow-Origin", "*")
		response.Header().Set("Access-Control-Allow-Methods", "*")
		response.Header().Set("Access-Control-Allow-Headers", "*")
		response.Header().Set("Access-Control-Expose-Headers", "*")

		if request.Method != http.MethodOptions {
			next(response, request)
		}
	}
}

func frontendHandler(response http.ResponseWriter, request *http.Request) {
	frontendFilePath := environment.GetFrontendPath()

	fileSystem := http.Dir(frontendFilePath)
	fileServer := http.FileServer(fileSystem)
	_, err := fileSystem.Open(path.Clean(request.URL.Path))

	if errors.Is(err, os.ErrNotExist) {
		http.ServeFile(response, request, frontendFilePath+"/index.html")
	} else {
		fileServer.ServeHTTP(response, request)
	}
}
