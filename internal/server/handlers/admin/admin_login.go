package admin

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

func AdminLoginHandler(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Verifying Admin Login")
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); !isValidMethod {
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHttpError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	err := json.NewEncoder(responseWriter).Encode(sessionResult)
	if err != nil {
		log.Println("API.Admin.Login Error", err)
	}
}
