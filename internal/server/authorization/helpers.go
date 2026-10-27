package authorization

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/google/uuid"
)

func assureProfilePath() {
	profilePath := os.Getenv(environment.StreamProfilePath)

	err := os.MkdirAll(profilePath, os.ModePerm)
	if err != nil {
		log.Println("Authorization: Error creating profile path folder folder:", err)
		return
	}
}

func hasExistingStreamKey(streamKey string) bool {
	profilePath := os.Getenv(environment.StreamProfilePath)
	files, err := os.ReadDir(profilePath)

	if err != nil {
		log.Println("Authorization: Error reading profile directory", err)
		return false
	}

	filePrefix := streamKey + "_"
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), filePrefix) {
			return true
		}
	}

	return false
}

func hasExistingBearerToken(bearerToken string) bool {
	profilePath := os.Getenv(environment.StreamProfilePath)

	files, err := os.ReadDir(profilePath)
	if err != nil {
		log.Println("Authorization: Error reading profile directory", err)
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
		log.Println("Authorization: Error reading profile directory", err)
		return "", err
	}

	for _, file := range files {
		fileToken := strings.Split(file.Name(), "_")

		if !file.IsDir() && strings.EqualFold(streamKey, fileToken[0]) {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("could not find profile file")
}

func getProfileFileNameByBearerToken(bearerToken string) (string, error) {
	profilePath := os.Getenv(environment.StreamProfilePath)

	files, err := os.ReadDir(profilePath)
	if err != nil {
		log.Println("Authorization: Error reading profile directory", err)
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

func generateToken() string {
	token := uuid.New().String()

	if hasExistingBearerToken(token) {
		return generateToken()
	}

	return token
}
