package admin

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

type MusicPlaylist struct {
	Title         string   `json:"title"`
	Subtitle      string   `json:"subtitle"`
	VideoIDs      []string `json:"videoIds"`
	DefaultVolume int      `json:"defaultVolume"`
}

var musicVideoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{6,32}$`)

func MusicPlaylistPath() string {
	if path := os.Getenv(environment.MusicPlaylistPath); path != "" {
		return path
	}
	return "music.json"
}

func DefaultMusicPlaylist() MusicPlaylist {
	return MusicPlaylist{
		Title:         "frutiger aero radio",
		Subtitle:      "nintendo mix · eShop / Wii / 3DS / DS",
		VideoIDs:      []string{"v3gOQbh32Pg", "0yiwxIuXmdk", "ZUgG3LGGbb4", "frIxIZyko80", "M6CmkoF-1oM"},
		DefaultVolume: 30,
	}
}

func LoadMusicPlaylist() (MusicPlaylist, error) {
	data, err := os.ReadFile(MusicPlaylistPath())
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultMusicPlaylist(), nil
		}
		return MusicPlaylist{}, err
	}

	var playlist MusicPlaylist
	if err := json.Unmarshal(data, &playlist); err != nil {
		return MusicPlaylist{}, err
	}

	if len(playlist.VideoIDs) == 0 {
		playlist = DefaultMusicPlaylist()
	}
	if playlist.DefaultVolume < 0 || playlist.DefaultVolume > 100 {
		playlist.DefaultVolume = 30
	}
	return playlist, nil
}

func SaveMusicPlaylist(playlist MusicPlaylist) error {
	data, err := json.MarshalIndent(playlist, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(MusicPlaylistPath(), data, 0644)
}

// MusicHandler lets admins read and update the site music playlist.
func MusicHandler(responseWriter http.ResponseWriter, request *http.Request) {
	sessionResult := verifyAdminSession(request)
	if !sessionResult.IsValid {
		helpers.LogHTTPError(responseWriter, sessionResult.ErrorMessage, http.StatusUnauthorized)
		return
	}

	switch request.Method {
	case http.MethodGet:
		playlist, err := LoadMusicPlaylist()
		if err != nil {
			helpers.LogHTTPError(responseWriter, "Could not load music playlist", http.StatusInternalServerError)
			return
		}
		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(playlist)

	case http.MethodPost:
		var playlist MusicPlaylist
		if err := json.NewDecoder(request.Body).Decode(&playlist); err != nil {
			helpers.LogHTTPError(responseWriter, "Invalid payload", http.StatusBadRequest)
			return
		}

		cleanIDs := []string{}
		for _, id := range playlist.VideoIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if !musicVideoIDPattern.MatchString(id) {
				helpers.LogHTTPError(responseWriter, "Invalid YouTube video id: "+id, http.StatusBadRequest)
				return
			}
			cleanIDs = append(cleanIDs, id)
		}
		if len(cleanIDs) == 0 {
			helpers.LogHTTPError(responseWriter, "At least one video id is required", http.StatusBadRequest)
			return
		}
		playlist.VideoIDs = cleanIDs
		if playlist.Title == "" {
			playlist.Title = DefaultMusicPlaylist().Title
		}
		if playlist.Subtitle == "" {
			playlist.Subtitle = DefaultMusicPlaylist().Subtitle
		}
		if playlist.DefaultVolume < 0 || playlist.DefaultVolume > 100 {
			playlist.DefaultVolume = 30
		}

		if err := SaveMusicPlaylist(playlist); err != nil {
			helpers.LogHTTPError(responseWriter, "Could not save music playlist", http.StatusInternalServerError)
			return
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(responseWriter).Encode(playlist)

	default:
		helpers.LogHTTPError(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
