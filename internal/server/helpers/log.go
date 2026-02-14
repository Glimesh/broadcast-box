package helpers

import (
	"log"
	"net/http"
)

func LogHttpError(responseWriter http.ResponseWriter, error string, code int) {
	log.Println("LogHttpError", error)
	http.Error(responseWriter, error, code)
}
