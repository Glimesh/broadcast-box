package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

type sessionResponse struct {
	IsValid      bool   `json:"isValid"`
	ErrorMessage string `json:"errorMessage"`
}

// Verify that a bearer token is provided for an admin session
// A response will be written to the response writter if the session is valid
func verifyAdminSession(request *http.Request) *sessionResponse {
	token := helpers.ResolveBearerToken(request.Header.Get("Authorization"))
	if token == "" {
		slog.Info("Authorization was not set")

		return &sessionResponse{
			IsValid:      false,
			ErrorMessage: "Authorization was invalid",
		}
	}

	adminAPIToken := os.Getenv(environment.FrontendAdminToken)

	if adminAPIToken == "" || !strings.EqualFold(adminAPIToken, token) {
		return &sessionResponse{
			IsValid:      false,
			ErrorMessage: "Authorization was invalid",
		}
	}

	return &sessionResponse{
		IsValid:      true,
		ErrorMessage: "",
	}
}

// Verify the expected method and return true or false if the method is as expected
// This will write a default METHOD NOT ALLOWED response on the responsewriter
func verifyValidMethod(expectedMethod string, responseWriter http.ResponseWriter, request *http.Request) bool {
	if !strings.EqualFold(expectedMethod, request.Method) {
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		err := json.NewEncoder(responseWriter).Encode(&sessionResponse{
			IsValid:      false,
			ErrorMessage: "Method not allowed",
		})

		if err != nil {
			slog.Error("Admin.Helpers Error", "err", err)
			return false
		}

		return false
	}

	return true
}
