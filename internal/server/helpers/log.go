package helpers

import (
	"log"
	"net/http"
)

func LogHttpError(responseWriter http.ResponseWriter, error string, code int) {
	log.Println(error)
	http.Error(responseWriter, error, code)
}
