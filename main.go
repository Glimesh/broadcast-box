package main

import (
	"encoding/json"
	"log"
	"net/http"
)

var (
	streamKey = ""
)

type statusResponse struct {
	Status string `json:"status"`
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := "unconfigured"
	if streamKey != "" {
		status = "configured"
	}

	if err := json.NewEncoder(w).Encode(&statusResponse{Status: status}); err != nil {
		log.Fatal(err)
	}
}

func configureHandler(w http.ResponseWriter, r *http.Request) {}
func whipHandler(w http.ResponseWriter, r *http.Request)      {}
func whepHandler(w http.ResponseWriter, r *http.Request)      {}

func main() {
	h := http.NewServeMux()
	h.HandleFunc("/api/status", statusHandler)
	h.HandleFunc("/api/configure", configureHandler)
	h.HandleFunc("/api/whip", whipHandler)
	h.HandleFunc("/api/whep", whipHandler)

	s := &http.Server{
		Handler: h,
		Addr:    ":8080",
	}

	log.Fatal(s.ListenAndServe())
}
