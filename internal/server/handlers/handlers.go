package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
)

func GetServeMuxHandler() http.HandlerFunc {
	serverMux := http.NewServeMux()

	frontendPath := os.Getenv("FRONTEND_PATH")
	if os.Getenv("FRONTEND_ENABLED") != "" {
		serverMux.Handle("/", serveFrontend(http.Dir(frontendPath)))
	}

	serverMux.HandleFunc("/api/whip", corsHandler(whipHandler))
	serverMux.HandleFunc("/api/whep", corsHandler(whepHandler))
	serverMux.HandleFunc("/api/sse/", corsHandler(sseHandler))
	serverMux.HandleFunc("/api/layer/", corsHandler(layerChangeHandler))
	serverMux.HandleFunc("/api/status", corsHandler(statusHandler))

	// Path middleware
	debugOutputWebRequests := os.Getenv("DEBUG_INCOMING_API_REQUEST")
	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if debugOutputWebRequests != "" {
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

func serveFrontend(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, err := fs.Open(path.Clean(request.URL.Path))

		if errors.Is(err, os.ErrNotExist) {
			http.ServeFile(response, request, "./web/build/index.html")

			return
		}

		fileServer.ServeHTTP(response, request)
	})
}
