package authorization

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Profile struct {
	StreamKey string `json:"streamKey"`
	IsPublic  bool   `json:"isPublic"`
	MOTD      string `json:"motd"`
}

func CreateProfile(streamKey string) (string, error) {
	profilePath := os.Getenv("STREAM_PROFILE_PATH")
	assureProfilePath()

	if hasExistingStreamKey(streamKey) {
		return "", fmt.Errorf("%s", "A profile with the stream key "+streamKey+" already exists")
	}

	token := generateToken()

	profileFilePath := filepath.Join(profilePath, streamKey+"_"+token)
	profile := Profile{
		StreamKey: streamKey,
		IsPublic:  true,
		MOTD:      "Welcome to my stream!",
	}

	jsonData, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		log.Println("Error ocurred while trying to create profile")
		log.Println(err)
		return "", err
	}

	err = os.WriteFile(profileFilePath, jsonData, 0644)
	if err != nil {
		log.Println("Error ocurred while trying to create profile")
		log.Println(err)
		return "", err
	}

	return token, nil
}

func GetProfile(bearerToken string) (*Profile, error) {
	profilePath := os.Getenv("STREAM_PROFILE_PATH")
	assureProfilePath()

	fileName, err := getProfileFileNameByBearerToken(bearerToken)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(profilePath, fileName))
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		log.Println("File", bearerToken, "could not read. File may be corrupt.")
		return nil, err
	}

	return &profile, nil
}
