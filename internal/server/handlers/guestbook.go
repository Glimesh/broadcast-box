package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/server/helpers"
	"github.com/google/uuid"
)

const guestbookPath = "guestbook.json"
const maxGuestbookEntries = 200
const maxGuestbookNameLength = 40
const maxGuestbookMessageLength = 280

type guestbookEntry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

func loadGuestbook() ([]guestbookEntry, error) {
	data, err := os.ReadFile(guestbookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []guestbookEntry{}, nil
		}
		return nil, err
	}

	var entries []guestbookEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func saveGuestbook(entries []guestbookEntry) error {
	data, err := json.MarshalIndent(entries, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(guestbookPath, data, 0644)
}

// GuestbookHandler lists and accepts guestbook entries.
func GuestbookHandler(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		entries, err := loadGuestbook()
		if err != nil {
			helpers.LogHTTPError(responseWriter, "Could not load guestbook", http.StatusInternalServerError)
			return
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].CreatedAt.After(entries[j].CreatedAt)
		})

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.Header().Set("Cache-Control", "no-store")
		if err := json.NewEncoder(responseWriter).Encode(entries); err != nil {
			log.Println("API.Guestbook Error", err)
		}

	case http.MethodPost:
		if !helpers.DefaultRateLimiter.Allow("guestbook:"+request.RemoteAddr, 5, time.Minute) {
			helpers.LogHTTPError(responseWriter, "Too many guestbook entries", http.StatusTooManyRequests)
			return
		}

		var payload struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			helpers.LogHTTPError(responseWriter, "Invalid guestbook payload", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(payload.Name)
		message := strings.TrimSpace(payload.Message)
		if name == "" {
			name = "anonymous"
		}
		if len(name) > maxGuestbookNameLength {
			name = name[:maxGuestbookNameLength]
		}
		if message == "" {
			helpers.LogHTTPError(responseWriter, "Message is required", http.StatusBadRequest)
			return
		}
		if len(message) > maxGuestbookMessageLength {
			message = message[:maxGuestbookMessageLength]
		}

		entries, err := loadGuestbook()
		if err != nil {
			helpers.LogHTTPError(responseWriter, "Could not load guestbook", http.StatusInternalServerError)
			return
		}

		entry := guestbookEntry{
			ID:        uuid.NewString(),
			Name:      name,
			Message:   message,
			CreatedAt: time.Now(),
		}

		entries = append(entries, entry)
		if len(entries) > maxGuestbookEntries {
			entries = entries[len(entries)-maxGuestbookEntries:]
		}

		if err := saveGuestbook(entries); err != nil {
			helpers.LogHTTPError(responseWriter, "Could not save guestbook", http.StatusInternalServerError)
			return
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(responseWriter).Encode(entry)

	default:
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
