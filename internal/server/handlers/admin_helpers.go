package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

type SessionResponse struct {
	IsValid      bool   `json:"isValid"`
	ErrorMessage string `json:"errorMessage"`
}

// Verify that a bearer token is provided for an admin session
// A response will be written to the response writter if the session is valid
func verifyAdminSession(request *http.Request) *SessionResponse {
	token := helpers.ResolveBearerToken(request.Header.Get("Authorization"))
	if token == "" {
		log.Println("Authorization was not set")

		return &SessionResponse{
			IsValid:      false,
			ErrorMessage: "Authorization was invalid",
		}
	}

	adminApiToken := os.Getenv(environment.FRONTEND_ADMIN_TOKEN)

	if adminApiToken == "" || !strings.EqualFold(adminApiToken, token) {
		return &SessionResponse{
			IsValid:      false,
			ErrorMessage: "Authorization was invalid",
		}
	}

	return &SessionResponse{
		IsValid:      true,
		ErrorMessage: "",
	}
}

// Verify the expected method and return true or false if the method is as expected
// This will write a default METHOD NOT ALLOWED response on the responsewriter
func verifyValidMethod(expectedMethod string, responseWriter http.ResponseWriter, request *http.Request) bool {
	log.Println("Verifying request method. Expecting", expectedMethod, "got", request.Method, " - ", strings.EqualFold(expectedMethod, request.Method))

	if strings.EqualFold(expectedMethod, request.Method) == false {
		helpers.LogHttpError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		json.NewEncoder(responseWriter).Encode(&SessionResponse{
			IsValid:      false,
			ErrorMessage: "Method not allowed",
		})

		return false
	}

	return true
}
