package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultTimeout = time.Second * 5

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

func CallWebhook(url, action, bearerToken string, r *http.Request) (string, error) {
	start := time.Now()

	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	jsonPayload, err := json.Marshal(webhookPayload{
		Action:      action,
		IP:          getIPAddress(r),
		BearerToken: bearerToken,
		QueryParams: queryParams,
		UserAgent:   r.UserAgent(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{
		Timeout: defaultTimeout,
	}).Do(req)
	if err != nil {
		return "", fmt.Errorf("webhook request failed after %v: %w", time.Since(start), err)
	}
	defer resp.Body.Close() //nolint

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("webhook returned non-200 Status: %v", resp.StatusCode)
	}

	response := webhookResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.StreamKey, nil
}

func getIPAddress(r *http.Request) string {
	if r.Header.Get("X-Forwarded-For") != "" {
		return r.Header.Get("X-Forwarded-For")
	}
	return r.RemoteAddr
}
