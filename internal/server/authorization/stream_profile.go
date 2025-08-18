package authorization

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glimesh/broadcast-box/internal/environment"
)

type Profile struct {
	StreamKey string `json:"streamKey"`
	IsActive  bool   `json:"isActive"`
	IsPublic  bool   `json:"isPublic"`
	MOTD      string `json:"motd"`
}

const (
	STREAM_POLICY_ANYONE        = "ANYONE"
	STREAM_POLICY_WITH_RESERVED = "ANYONE_WITH_RESERVED"
	STREAM_POLICY_RESERVED_ONLY = "RESERVED"
)

func CreateProfile(streamKey string) (string, error) {
	profilePath := os.Getenv(environment.STREAM_PROFILE_PATH)
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
		log.Println("Authorization: Error ocurred while trying to create profile")
		log.Println(err)
		return "", err
	}

	err = os.WriteFile(profileFilePath, jsonData, 0644)
	if err != nil {
		log.Println("Authorization: Error ocurred while trying to create profile")
		log.Println(err)
		return "", err
	}

	return token, nil
}

func GetProfile(bearerToken string) (*Profile, error) {
	profilePath := os.Getenv(environment.STREAM_PROFILE_PATH)
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
		log.Println("Authorization: File", bearerToken, "could not read. File may be corrupt.")
		return nil, err
	}

	return &profile, nil
}

func IsProfileReserved(streamKey string) bool {
	assureProfilePath()

	fileName, _ := getProfileFileNameByStreamKey(streamKey)
	if fileName != "" {
		log.Println("Authorization: Profile is reserved")
		return true
	}

	log.Println("Authorization: Profile is not reserved")
	return false
}
