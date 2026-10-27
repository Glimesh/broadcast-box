package authorization

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glimesh/broadcast-box/internal/environment"
)

const adhocProfileDir = "adhoc"

// adhocProfile stores title and visibility for stream keys that were not
// created as reserved profiles. Keeping these files in a subdirectory means
// they do not make the stream key "reserved" for WHIP authorization.
type adhocProfile struct {
	IsPublic bool   `json:"isPublic"`
	MOTD     string `json:"motd"`
}

func adhocProfilePath() string {
	return filepath.Join(os.Getenv(environment.StreamProfilePath), adhocProfileDir)
}

func assureAdhocProfilePath() error {
	return os.MkdirAll(adhocProfilePath(), os.ModePerm)
}

func adhocProfileFilePath(streamKey string) (string, error) {
	if !isValidStreamKey(streamKey) {
		return "", fmt.Errorf("streamkey has invalid characters, only numbers, letters, dash and underscore allowed")
	}

	return filepath.Join(adhocProfilePath(), streamKey+".json"), nil
}

func getAdhocProfile(streamKey string) (*adhocProfile, error) {
	profileFilePath, err := adhocProfileFilePath(streamKey)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(profileFilePath)
	if err != nil {
		return nil, err
	}

	var profile adhocProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		log.Println("Authorization: Ad-hoc profile could not be read. File may be corrupt.")
		return nil, err
	}

	return &profile, nil
}

func setAdhocProfile(streamKey string, motd string, isPublic bool) error {
	if err := assureAdhocProfilePath(); err != nil {
		log.Println("Authorization: Error creating ad-hoc profile path:", err)
		return err
	}

	profileFilePath, err := adhocProfileFilePath(streamKey)
	if err != nil {
		return err
	}

	profile := adhocProfile{
		IsPublic: isPublic,
		MOTD:     motd,
	}

	if existing, err := getAdhocProfile(streamKey); err == nil {
		if motd == "" {
			profile.MOTD = existing.MOTD
		}
	}

	if profile.MOTD == "" {
		profile.MOTD = "Welcome to " + streamKey + "!"
	}

	jsonData, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(profileFilePath, jsonData, 0644)
}
