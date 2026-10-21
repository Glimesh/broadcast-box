package authorization

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/google/uuid"
)

func assureProfilePath() {
	profilePath := os.Getenv(environment.StreamProfilePath)

	err := os.MkdirAll(profilePath, os.ModePerm)
	if err != nil {
		slog.Error("Authorization: Error creating profile path folder folder", "err", err)
		return
	}
}

func hasExistingStreamKey(streamKey string) bool {
	profilePath := os.Getenv(environment.StreamProfilePath)
	files, err := os.ReadDir(profilePath)

	if err != nil {
		slog.Error("Authorization: Error reading profile directory", "err", err)
		return false
	}

	for _, file := range files {
		if !file.IsDir() && streamKeyFromFileName(file.Name()) == streamKey {
			return true
		}
	}

	return false
}

func hasExistingBearerToken(bearerToken string) bool {
	profilePath := os.Getenv(environment.StreamProfilePath)

	files, err := os.ReadDir(profilePath)
	if err != nil {
		slog.Error("Authorization: Error reading profile directory", "err", err)
		return false
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), bearerToken) {
			return true
		}
	}

	return false
}

func getProfileFileNameByStreamKey(streamKey string) (string, error) {
	profilePath := os.Getenv(environment.StreamProfilePath)

	files, err := os.ReadDir(profilePath)
	if err != nil {
		slog.Error("Authorization: Error reading profile directory", "err", err)
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() && strings.EqualFold(streamKey, streamKeyFromFileName(file.Name())) {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("could not find profile file")
}

func getProfileFileNameByBearerToken(bearerToken string) (string, error) {
	profilePath := os.Getenv(environment.StreamProfilePath)

	files, err := os.ReadDir(profilePath)
	if err != nil {
		slog.Error("Authorization: Error reading profile directory", "err", err)
		return "", err
	}

	separator := "_"
	for _, file := range files {
		splitIndex := strings.LastIndex(file.Name(), separator)
		fileToken := file.Name()[splitIndex+len(separator):]

		if !file.IsDir() && strings.EqualFold(bearerToken, fileToken) {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("could not find profile file")
}

// Profile files are named <streamKey>_<token>. The token never contains an
// underscore, but the stream key may, so split on the last one.
func streamKeyFromFileName(fileName string) string {
	splitIndex := strings.LastIndex(fileName, "_")
	if splitIndex == -1 {
		return fileName
	}

	return fileName[:splitIndex]
}

func generateToken() string {
	token := uuid.New().String()

	if hasExistingBearerToken(token) {
		return generateToken()
	}

	return token
}
