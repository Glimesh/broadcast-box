package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	accountStore "github.com/glimesh/broadcast-box/internal/accounts"
	"github.com/glimesh/broadcast-box/internal/environment"
	accountHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/accounts"
	adminHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/admin"
	whipHandlers "github.com/glimesh/broadcast-box/internal/server/handlers/whip"
)

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

	// Custom emotes
	serverMux.HandleFunc("/api/emotes", corsHandler(EmotesHandler))
	serverMux.HandleFunc("/api/chat/delete", corsHandler(ChatDeleteHandler))
	serverMux.HandleFunc("/api/guestbook", corsHandler(GuestbookHandler))
	serverMux.HandleFunc("/api/music", corsHandler(MusicHandler))
	serverMux.Handle("/emotes/", http.StripPrefix("/emotes/", http.FileServer(http.Dir(adminHandlers.EmotesPath()))))

	// User accounts
	serverMux.HandleFunc("/api/account/register", corsHandler(accountHandlers.RegisterHandler))
	serverMux.HandleFunc("/api/account/login", corsHandler(accountHandlers.LoginHandler))
	serverMux.HandleFunc("/api/account/logout", corsHandler(accountHandlers.LogoutHandler))
	serverMux.HandleFunc("/api/account/avatar", corsHandler(accountHandlers.AvatarHandler))
	serverMux.HandleFunc("/api/account/update", corsHandler(accountHandlers.UpdateHandler))
	serverMux.HandleFunc("/api/account/regenerate-token", corsHandler(accountHandlers.RegenerateTokenHandler))
	serverMux.HandleFunc("/api/account/emotes", corsHandler(accountHandlers.UserEmotesHandler))
	serverMux.HandleFunc("/api/account", corsHandler(accountHandlers.MeHandler))
	serverMux.HandleFunc("/api/accounts/", corsHandler(accountHandlers.PublicProfileHandler))
	serverMux.Handle("/avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir(accountStore.AvatarsPath()))))

	// Admin endpoints
	serverMux.HandleFunc("/api/admin/login", corsHandler(adminHandlers.LoginHandler))
	serverMux.HandleFunc("/api/admin/status", corsHandler(adminHandlers.StatusHandler))
	serverMux.HandleFunc("/api/admin/logging", corsHandler(adminHandlers.LoggingHandler))
	serverMux.HandleFunc("/api/admin/profiles", corsHandler(adminHandlers.ProfilesHandler))
	serverMux.HandleFunc("/api/admin/profiles/reset-token", corsHandler(adminHandlers.ProfilesResetTokenHandler))
	serverMux.HandleFunc("/api/admin/accounts", corsHandler(adminHandlers.AccountsHandler))
	serverMux.HandleFunc("/api/admin/accounts/delete", corsHandler(adminHandlers.AccountsDeleteHandler))
	serverMux.HandleFunc("/api/admin/accounts/permission", corsHandler(adminHandlers.AccountPermissionHandler))
	serverMux.HandleFunc("/api/admin/registration", corsHandler(adminHandlers.RegistrationHandler))
	serverMux.HandleFunc("/api/admin/profiles/add-profile", corsHandler(adminHandlers.ProfileAddHandler))
	serverMux.HandleFunc("/api/admin/profiles/remove-profile", corsHandler(adminHandlers.ProfileRemoveHandler))
	serverMux.HandleFunc("/api/admin/emotes", corsHandler(adminHandlers.EmotesHandler))
	serverMux.HandleFunc("/api/admin/music", corsHandler(adminHandlers.MusicHandler))

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
