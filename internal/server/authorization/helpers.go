package authorization

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
)

func assureProfilePath() {
	profilePath := os.Getenv("STREAM_PROFILE_PATH")

	err := os.MkdirAll(profilePath, os.ModePerm)
	if err != nil {
		log.Println("Error creating profile path folder folder:", err)
		return
	}
}

func IsValidStreamBearerToken(bearerToken string) bool {
	return hasExistingBearerToken(bearerToken)
}

func hasExistingStreamKey(streamKey string) bool {
	profilePath := os.Getenv("STREAM_PROFILE_PATH")
	files, err := os.ReadDir(profilePath)

	if err != nil {
		log.Println("Error reading profile directory", err)
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
	profilePath := os.Getenv("STREAM_PROFILE_PATH")

	files, err := os.ReadDir(profilePath)
	if err != nil {
		log.Println("Error reading profile directory", err)
		return false
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), bearerToken) {
			return true
		}
	}

	return false
}

func getProfileFileNameByBearerToken(bearerToken string) (string, error) {
	profilePath := os.Getenv("STREAM_PROFILE_PATH")

	files, err := os.ReadDir(profilePath)
	if err != nil {
		log.Println("Error reading profile directory", err)
		return "", err
	}

	for _, file := range files {
		fileToken := strings.SplitAfter(file.Name(), "_")

		if !file.IsDir() && strings.EqualFold(bearerToken, fileToken[len(fileToken)-1]) {
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
