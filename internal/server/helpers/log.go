package helpers

import (
	"log"
	"net/http"
)

func LogHTTPError(responseWriter http.ResponseWriter, error string, code int) {
	log.Println("LogHTTPError", error)
	http.Error(responseWriter, error, code)
}
