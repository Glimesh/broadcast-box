package authorization

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"

	"github.com/glimesh/broadcast-box/internal/environment"
)

const (
	streamPolicyAnyone       = "ANYONE"
	StreamPolicyWithReserved = "ANYONE_WITH_RESERVED"
	StreamPolicyReservedOnly = "RESERVED"
)

func isValidStreamKey(streamKey string) bool {
	regExp := regexp.MustCompile(`[\p{L}\p{N}_-]+`)
	return regExp.MatchString(streamKey)
}

// Create a new profile for the provided streamkey
func CreateProfile(streamKey string) (string, error) {

	if !isValidStreamKey(streamKey) {
		slog.Info("Authorization: Create profile failed due to invalid streamkey", "streamKey", streamKey)
		return "", fmt.Errorf("streamkey has invalid characters, only numbers, letters, dash and underscore allowed")
	}

	profilePath := os.Getenv(environment.StreamProfilePath)
	assureProfilePath()

	if hasExistingStreamKey(streamKey) {
		return "", fmt.Errorf("%s", "A profile with the stream key "+streamKey+" already exists")
	}

	token := generateToken()

	fileName := streamKey + "_" + token
	profileFilePath := filepath.Join(profilePath, fileName)
	profile := profile{
		FileName: fileName,
		IsPublic: true,
		MOTD:     "Welcome to " + streamKey + "!",
	}

	jsonData, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		slog.Error("Authorization: Error ocurred while trying to create profile", "err", err)
		return "", err
	}

	err = os.WriteFile(profileFilePath, jsonData, 0644)
	if err != nil {
		slog.Error("Authorization: Error ocurred while trying to create profile", "err", err)
		return "", err
	}

	return token, nil
}

// Update a current profile
func UpdateProfile(token string, motd string, isPublic bool) error {
	if !hasExistingBearerToken(token) {
		return fmt.Errorf("profile was not found")
	}

	profile, err := GetPersonalProfile(token)
	if err != nil {
		slog.Error("Authorization: Could not find personal profile", "err", err)
		return err
	}

	// Update properties
	profile.MOTD = motd
	profile.IsPublic = isPublic

	jsonData, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		slog.Error("Authorization: Error ocurred while trying to update profile", "err", err)
		return err
	}

	profilePath := os.Getenv(environment.StreamProfilePath)
	profileFilePath, err := getProfileFileNameByBearerToken(token)
	if err != nil {
		slog.Error("Authorization: Error ocurred while trying to update profile", "err", err)
		return err
	}

	slog.Info("Authorization: Updated Profile", "streamKey", profile.StreamKey)
	err = os.WriteFile(filepath.Join(profilePath, profileFilePath), jsonData, 0644)
	if err != nil {
		slog.Error("Authorization: Error ocurred while trying to update profile", "err", err)
		return err
	}

	return nil
}

func RemoveProfile(streamKey string) (bool, error) {
	if !isValidStreamKey(streamKey) {
		slog.Error("Authorization: Remove profile failed due to invalid streamkey", "streamKey", streamKey)
		return false, fmt.Errorf("streamkey has invalid characters, only numbers, letters, dash and underscore allowed")
	}

	fileName, _ := getProfileFileNameByStreamKey(streamKey)
	if fileName == "" {
		slog.Info("Authorization: RemoveProfile could not find", "sreamKey", streamKey)
		return false, fmt.Errorf("profile could not be found")
	}

	profilePath := os.Getenv(environment.StreamProfilePath)
	err := os.Remove(filepath.Join(profilePath, fileName))
	if err != nil {
		return false, err
	}

	return true, nil
}

// Returns the publicly available profile
func GetPublicProfile(bearerToken string) (*PublicProfile, error) {
	profilePath := os.Getenv(environment.StreamProfilePath)
	assureProfilePath()

	fileName, err := getProfileFileNameByBearerToken(bearerToken)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(profilePath, fileName))
	if err != nil {
		return nil, err
	}

	var profile profile
	if err := json.Unmarshal(data, &profile); err != nil {
		slog.Error("Authorization: could not read. File may be corrupt", "err", err, "bearerToken", bearerToken)
		return nil, err
	}
	profile.FileName = fileName

	return profile.asPublicProfile(), nil
}

// Returns the publicly available profile
func GetPersonalProfile(bearerToken string) (*PersonalProfile, error) {
	profilePath := os.Getenv(environment.StreamProfilePath)
	assureProfilePath()

	fileName, err := getProfileFileNameByBearerToken(bearerToken)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(profilePath, fileName))
	if err != nil {
		return nil, err
	}

	var profile profile
	if err := json.Unmarshal(data, &profile); err != nil {
		slog.Error("Authorization: could not read. File may be corrupt", "err", err, "bearerToken", bearerToken)
		return nil, err
	}
	profile.FileName = fileName

	return profile.asPersonalProfile(), nil
}

// Returns a slice of profiles intended for admin endpoints
func GetAdminProfilesAll() (profiles []adminProfile, err error) {
	profilePath := os.Getenv(environment.StreamProfilePath)

	files, err := os.ReadDir(profilePath)
	if err != nil {
		slog.Error("Authorization: Error reading profile directory", "err", err)
		return nil, err
	}

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(profilePath, file.Name()))
		if err != nil {
			profiles = append(profiles, adminProfile{
				StreamKey: file.Name(),
				IsPublic:  false,
				MOTD:      "Error reading profile from file: " + file.Name(),
			})

			continue
		}

		var profile profile

		if err := json.Unmarshal(data, &profile); err != nil {
			profiles = append(profiles, adminProfile{
				StreamKey: file.Name(),
				IsPublic:  false,
				MOTD:      "Invalid JSON in file" + file.Name(),
			})
			continue
		}

		profile.FileName = file.Name()
		profiles = append(profiles, *profile.asAdminProfile())
	}

	return profiles, nil
}

func IsProfileReserved(streamKey string) bool {
	assureProfilePath()

	fileName, _ := getProfileFileNameByStreamKey(streamKey)
	return fileName != ""
}

func ResetProfileToken(streamKey string) error {
	fileName, _ := getProfileFileNameByStreamKey(streamKey)

	if fileName == "" {
		return fmt.Errorf("authorization: profile could not be found")
	}

	profilePath := os.Getenv(environment.StreamProfilePath)
	newFileName := streamKey + "_" + generateToken()
	currentPath := filepath.Join(profilePath, fileName)
	newPath := filepath.Join(profilePath, newFileName)

	if err := os.Rename(currentPath, newPath); err != nil {
		return fmt.Errorf("authorization: error updating profile token for %s", streamKey)
	}

	return nil
}
