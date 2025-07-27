package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCallWebhook(t *testing.T) {
	// Setup a Mock HTTP Server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(webhookResponse{StreamKey: "dummy_stream_key"})
		case "/timeout":
			time.Sleep(2 * time.Second)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/badjson":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not a json"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	tests := []struct {
		name        string
		url         string
		timeout     int
		expectedErr bool
		expectedKey string
	}{
		{"Success Case", "/ok", 1000, false, "dummy_stream_key"},
		{"Server Timeout", "/timeout", 1000, true, ""},
		{"Server Error", "/error", 1000, true, ""},
		{"Malformed JSON", "/badjson", 1000, true, ""},
		{"Not Found", "/notfound", 1000, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = "127.0.0.1"
			req.Header.Set("User-Agent", "test-agent")

			// call the function with test layers
			result, err := CallWebhook(fmt.Sprintf("%s%s", mockServer.URL, tt.url), "action", "bearerToken", tt.timeout, req)

			if tt.expectedErr && err == nil {
				t.Fatalf("expected an error but got none")
			}

			if !tt.expectedErr && err != nil {
				t.Fatalf("did not expect an error but got %v", err)
			}

			if result != tt.expectedKey {
				t.Fatalf("expected stream key %s but got %s", tt.expectedKey, result)
			}
		})
	}
}
