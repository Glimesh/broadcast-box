package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type webhookPayload struct {
	Action      string            `json:"action"`
	IP          string            `json:"ip"`
	BearerToken string            `json:"bearerToken"`
	QueryParams map[string]string `json:"queryParams"`
	UserAgent   string            `json:"userAgent"`
}

type webhookResponse struct {
	StreamKey string `json:"streamKey"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
			return
		}

		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.BearerToken == "broadcastBoxRulez" {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(webhookResponse{StreamKey: payload.BearerToken}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusForbidden)
			if err := json.NewEncoder(w).Encode(webhookResponse{}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})

	log.Println("Server listening on port 8081")
	if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
