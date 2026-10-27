package admin

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

const maxEmoteSize = 8 << 20 // 8 MiB

var emoteNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

var allowedEmoteExtensions = map[string]bool{
	".png":  true,
	".gif":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".apng": true,
}

// EmoteInfo describes an uploaded custom emote.
type EmoteInfo struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	URL      string `json:"url"`
	Animated bool   `json:"animated"`
}

// EmotesPath returns the directory where custom emotes are stored.
func EmotesPath() string {
	if path := os.Getenv(environment.EmotePath); path != "" {
		return path
	}

	return "emotes"
}

// ListEmotes returns all uploaded emotes sorted by name.
func ListEmotes() ([]EmoteInfo, error) {
	emotesPath := EmotesPath()
	if err := os.MkdirAll(emotesPath, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(emotesPath)
	if err != nil {
		return nil, err
	}

	emotes := make([]EmoteInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if !allowedEmoteExtensions[extension] {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		emotes = append(emotes, EmoteInfo{
			Name:     name,
			File:     entry.Name(),
			URL:      "/emotes/" + url.PathEscape(entry.Name()),
			Animated: extension == ".gif" || extension == ".webp" || extension == ".apng",
		})
	}

	sort.Slice(emotes, func(i, j int) bool {
		return strings.ToLower(emotes[i].Name) < strings.ToLower(emotes[j].Name)
	})

	return emotes, nil
}

func writeJSON(responseWriter http.ResponseWriter, value any) {
	responseWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(responseWriter).Encode(value); err != nil {
		log.Println("Admin.Emotes Error:", err)
	}
}

// EmotesHandler handles admin listing, upload, and deletion of custom emotes.
func EmotesHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if !VerifyAdminSession(request) {
		helpers.LogHTTPError(responseWriter, "Authorization was invalid", http.StatusUnauthorized)
		return
	}

	switch request.Method {
	case http.MethodGet:
		emotes, err := ListEmotes()
		if err != nil {
			helpers.LogHTTPError(responseWriter, "Could not list emotes", http.StatusInternalServerError)
			return
		}
		writeJSON(responseWriter, emotes)

	case http.MethodPost:
		uploadEmote(responseWriter, request)

	case http.MethodDelete:
		deleteEmote(responseWriter, request)

	default:
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func uploadEmote(responseWriter http.ResponseWriter, request *http.Request) {
	if err := request.ParseMultipartForm(maxEmoteSize); err != nil {
		helpers.LogHTTPError(responseWriter, "Upload too large or malformed", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(request.FormValue("name"))
	file, header, err := request.FormFile("file")
	if err != nil {
		helpers.LogHTTPError(responseWriter, "Missing emote file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if name == "" {
		name = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	}

	if !emoteNamePattern.MatchString(name) {
		helpers.LogHTTPError(responseWriter, "Emote name may only contain letters, numbers, dash, and underscore (max 32)", http.StatusBadRequest)
		return
	}

	extension := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedEmoteExtensions[extension] {
		helpers.LogHTTPError(responseWriter, "Unsupported emote format. Use PNG, GIF, JPG, WEBP, or APNG", http.StatusBadRequest)
		return
	}

	emotesPath := EmotesPath()
	if err := os.MkdirAll(emotesPath, 0755); err != nil {
		helpers.LogHTTPError(responseWriter, "Could not create emote directory", http.StatusInternalServerError)
		return
	}

	// Replace any existing emote with the same name, regardless of extension.
	if existing, err := ListEmotes(); err == nil {
		for _, emote := range existing {
			if strings.EqualFold(emote.Name, name) {
				_ = os.Remove(filepath.Join(emotesPath, emote.File))
			}
		}
	}

	fileName := name + extension
	destinationPath := filepath.Join(emotesPath, fileName)
	destination, err := os.Create(destinationPath)
	if err != nil {
		helpers.LogHTTPError(responseWriter, "Could not save emote", http.StatusInternalServerError)
		return
	}
	defer destination.Close()

	if _, err := io.Copy(destination, io.LimitReader(file, maxEmoteSize)); err != nil {
		helpers.LogHTTPError(responseWriter, "Could not save emote", http.StatusInternalServerError)
		return
	}

	writeJSON(responseWriter, EmoteInfo{
		Name:     name,
		File:     fileName,
		URL:      "/emotes/" + url.PathEscape(fileName),
		Animated: extension == ".gif" || extension == ".webp" || extension == ".apng",
	})
}

func deleteEmote(responseWriter http.ResponseWriter, request *http.Request) {
	name := strings.TrimSpace(request.URL.Query().Get("name"))
	if !emoteNamePattern.MatchString(name) {
		helpers.LogHTTPError(responseWriter, "Invalid emote name", http.StatusBadRequest)
		return
	}

	emotes, err := ListEmotes()
	if err != nil {
		helpers.LogHTTPError(responseWriter, "Could not list emotes", http.StatusInternalServerError)
		return
	}

	for _, emote := range emotes {
		if strings.EqualFold(emote.Name, name) {
			if err := os.Remove(filepath.Join(EmotesPath(), emote.File)); err != nil {
				helpers.LogHTTPError(responseWriter, "Could not delete emote", http.StatusInternalServerError)
				return
			}
			responseWriter.WriteHeader(http.StatusNoContent)
			return
		}
	}

	helpers.LogHTTPError(responseWriter, "Emote not found", http.StatusNotFound)
}
