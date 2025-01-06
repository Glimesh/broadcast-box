package webrtc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type WebhookPayload struct {
	Action      string            `json:"action"`
	StreamKey   string            `json:"streamKey"`
	IP          string            `json:"ip"`
	BearerToken string            `json:"bearerToken"`
	QueryParams map[string]string `json:"queryParams"`
	UserAgent   string            `json:"userAgent"`
}

type WebhookResponse struct {
	Username string `json:"username"`
}

func CallWebhook(url string, timeout int, payload WebhookPayload) (string, int, error) {
	start := time.Now()
	log.Printf("Starting webhook call to %s with timeout %d ms", url, timeout)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Sending webhook request...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Webhook request failed after %v: %v", time.Since(start), err)
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	response := WebhookResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Printf("Failed to decode webhook response: %v", err)
		return "", resp.StatusCode, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Received webhook response with status code %d after %v", resp.StatusCode, time.Since(start))

	return response.Username, resp.StatusCode, nil
}
