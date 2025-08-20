package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

func adminLoginHandler(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Verifying Admin Login")
	if isValidMethod := verifyValidMethod("POST", responseWriter, request); isValidMethod != true {
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")

	sessionResult := verifyAdminSession(request)
	if sessionResult.IsValid != true {
		helpers.LogHttpError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
	}

	json.NewEncoder(responseWriter).Encode(sessionResult)
}
