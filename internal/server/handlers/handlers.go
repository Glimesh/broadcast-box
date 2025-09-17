package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	adminHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/admin"
	whipHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/whip"
)

func GetServeMuxHandler() http.HandlerFunc {
	serverMux := http.NewServeMux()

	if os.Getenv(environment.FRONTEND_DISABLED) == "" {
		serverMux.HandleFunc("/", frontendHandler)
	}

	// Whip/Whep shared endpoints
	serverMux.HandleFunc("/api/whep", corsHandler(WhepHandler))
	serverMux.HandleFunc("/api/sse/", corsHandler(sseHandler))
	serverMux.HandleFunc("/api/ice-servers", corsHandler(clientICEHandler))

	// Whip session endpoints
	serverMux.HandleFunc("/api/whip", corsHandler(whipHandlers.WhipHandler))
	serverMux.HandleFunc("/api/whip/profile", corsHandler(whipHandlers.ProfileHandler))

	// Whep session endpoints
	serverMux.HandleFunc("/api/layer/", corsHandler(layerChangeHandler))

	// Logging and status endpoints
	serverMux.HandleFunc("/api/log", corsHandler(logHandler))
	serverMux.HandleFunc("/api/status", corsHandler(statusHandler))

	// Admin endpoints
	// serverMux.HandleFunc("/api/admin/sse/", corsHandler(adminSseHandler))
	serverMux.HandleFunc("/api/admin/login", corsHandler(adminHandlers.AdminLoginHandler))
	serverMux.HandleFunc("/api/admin/status", corsHandler(adminHandlers.AdminStatusHandler))
	serverMux.HandleFunc("/api/admin/logging", corsHandler(adminHandlers.AdminLoggingHandler))
	serverMux.HandleFunc("/api/admin/profiles", corsHandler(adminHandlers.AdminProfilesHandler))
	serverMux.HandleFunc("/api/admin/profiles/reset-token", corsHandler(adminHandlers.AdminProfilesResetTokenHandler))
	serverMux.HandleFunc("/api/admin/profiles/add-profile", corsHandler(adminHandlers.AdminProfileAddHandler))
	serverMux.HandleFunc("/api/admin/profiles/remove-profile", corsHandler(adminHandlers.AdminProfileRemoveHandler))

	// Path middleware
	debugOutputWebRequests := os.Getenv(environment.DEBUG_INCOMING_API_REQUEST)
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

	defaultFrontendPath := "./web/build"

	frontendFilePath := os.Getenv(environment.FRONTEND_PATH)

	if frontendFilePath == "" {
		frontendFilePath = defaultFrontendPath
	}

	fileSystem := http.Dir(frontendFilePath)
	fileServer := http.FileServer(fileSystem)
	_, err := fileSystem.Open(path.Clean(request.URL.Path))

	if errors.Is(err, os.ErrNotExist) {
		http.ServeFile(response, request, frontendFilePath+"/index.html")
	} else {
		fileServer.ServeHTTP(response, request)
	}
}
