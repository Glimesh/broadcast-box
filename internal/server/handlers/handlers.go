package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
)

func GetServeMuxHandler() http.HandlerFunc {
	serverMux := http.NewServeMux()

	if os.Getenv(environment.FRONTEND_DISABLED) == "" {
		serverMux.HandleFunc("/", frontendHandler)
	}

	serverMux.HandleFunc("/api/whip", corsHandler(whipHandler))
	serverMux.HandleFunc("/api/whep", corsHandler(WhepHandler))
	serverMux.HandleFunc("/api/sse/", corsHandler(sseHandler))
	serverMux.HandleFunc("/api/layer/", corsHandler(layerChangeHandler))
	serverMux.HandleFunc("/api/log", corsHandler(logHandler))
	serverMux.HandleFunc("/api/status", corsHandler(statusHandler))
	serverMux.HandleFunc("/api/ice-servers", corsHandler(clientICEHandler))
	serverMux.HandleFunc("/api/admin/login", corsHandler(adminLoginHandler))
	serverMux.HandleFunc("/api/admin/status", corsHandler(adminStatusHandler))

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
	frontendFilePath := "./web/build"
	fileSystem := http.Dir(frontendFilePath)
	fileServer := http.FileServer(fileSystem)
	_, err := fileSystem.Open(path.Clean(request.URL.Path))

	if errors.Is(err, os.ErrNotExist) {
		http.ServeFile(response, request, "./web/build/index.html")
	} else {
		fileServer.ServeHTTP(response, request)
	}
}
