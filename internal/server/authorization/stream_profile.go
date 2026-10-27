package authorization

import (
	"encoding/json"
	"fmt"
	"log"
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
	regExp := regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)
	return regExp.MatchString(streamKey)
}

// Create a new profile for the provided streamkey
func CreateProfile(streamKey string) (string, error) {

	if !isValidStreamKey(streamKey) {
		log.Println("Authorization: Create profile failed due to invalid streamkey", streamKey)
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
		IsPublic: false,
		MOTD:     "Welcome to " + streamKey + "!",
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

// CreateProfileWithToken creates a reserved profile using a caller-provided token.
// This is used by user accounts so the account owner keeps a stable publish token.
func CreateProfileWithToken(streamKey string, token string, motd string, isPublic bool) error {
	if !isValidStreamKey(streamKey) {
		return fmt.Errorf("streamkey has invalid characters, only numbers, letters, dash and underscore allowed")
	}

	if token == "" || hasExistingBearerToken(token) {
		return fmt.Errorf("stream token is invalid or already in use")
	}

	profilePath := os.Getenv(environment.StreamProfilePath)
	assureProfilePath()

	if hasExistingStreamKey(streamKey) {
		return fmt.Errorf("a profile with the stream key %s already exists", streamKey)
	}

	fileName := streamKey + "_" + token
	profileFilePath := filepath.Join(profilePath, fileName)
	newProfile := profile{
		FileName: fileName,
		IsPublic: isPublic,
		MOTD:     motd,
	}

	jsonData, err := json.MarshalIndent(newProfile, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(profileFilePath, jsonData, 0644)
}

// SetProfileToken replaces the bearer token used by a reserved profile.
func SetProfileToken(streamKey string, token string) error {
	if token == "" || hasExistingBearerToken(token) {
		return fmt.Errorf("stream token is invalid or already in use")
	}

	fileName, err := getProfileFileNameByStreamKey(streamKey)
	if err != nil {
		return fmt.Errorf("profile could not be found")
	}

	profilePath := os.Getenv(environment.StreamProfilePath)
	currentPath := filepath.Join(profilePath, fileName)
	newPath := filepath.Join(profilePath, streamKey+"_"+token)

	return os.Rename(currentPath, newPath)
}

// Update a current profile
func UpdateProfile(token string, motd string, isPublic bool) error {
	if !hasExistingBearerToken(token) {
		return fmt.Errorf("profile was not found")
	}

	profile, err := GetPersonalProfile(token)
	if err != nil {
		log.Println("Authorization: Could not find personal profile")
		log.Println(err)
		return err
	}

	// Update properties
	profile.MOTD = motd
	profile.IsPublic = isPublic

	jsonData, err := json.MarshalIndent(profile, "", " ")
	if err != nil {
		log.Println("Authorization: Error ocurred while trying to update profile")
		log.Println(err)
		return err
	}

	profilePath := os.Getenv(environment.StreamProfilePath)
	profileFilePath, err := getProfileFileNameByBearerToken(token)
	if err != nil {
		log.Println("Authorization: Error ocurred while trying to update profile")
		log.Println(err)
		return err
	}

	log.Println("Authorization: Updated Profile", profile.StreamKey)
	err = os.WriteFile(filepath.Join(profilePath, profileFilePath), jsonData, 0644)
	if err != nil {
		log.Println("Authorization: Error ocurred while trying to update profile")
		log.Println(err)
		return err
	}

	return nil
}

// GetPersonalProfileByStreamKey returns a profile by stream key without reserving ad-hoc keys
func GetPersonalProfileByStreamKey(streamKey string) (*PersonalProfile, error) {
	profilePath := os.Getenv(environment.StreamProfilePath)
	assureProfilePath()

	fileName, err := getProfileFileNameByStreamKey(streamKey)
	if err != nil {
		// Ad-hoc streams are not reserved. Keep their title/visibility in a
		// separate store so editing them does not require a publish token.
		if adhoc, adhocErr := getAdhocProfile(streamKey); adhocErr == nil {
			return &PersonalProfile{
				StreamKey: streamKey,
				IsPublic:  adhoc.IsPublic,
				MOTD:      adhoc.MOTD,
			}, nil
		}

		return &PersonalProfile{
			StreamKey: streamKey,
			IsPublic:  false,
			MOTD:      "Welcome to " + streamKey + "!",
		}, nil
	}

	data, err := os.ReadFile(filepath.Join(profilePath, fileName))
	if err != nil {
		return nil, err
	}

	var p profile
	if err := json.Unmarshal(data, &p); err != nil {
		log.Println("Authorization: File could not read. File may be corrupt.")
		return nil, err
	}
	p.FileName = fileName

	return p.asPersonalProfile(), nil
}

// UpdateProfileByStreamKey updates a profile's visibility (and optionally title/motd) by stream key
func UpdateProfileByStreamKey(streamKey string, motd string, isPublic bool) error {
	profilePath := os.Getenv(environment.StreamProfilePath)
	assureProfilePath()

	fileName, err := getProfileFileNameByStreamKey(streamKey)
	if err != nil {
		// Ad-hoc streams stay unreserved. Save their settings separately so
		// the next OBS connection can still publish with the same stream key.
		return setAdhocProfile(streamKey, motd, isPublic)
	}

	data, err := os.ReadFile(filepath.Join(profilePath, fileName))
	if err != nil {
		return err
	}

	var p profile
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	p.FileName = fileName
	p.IsPublic = isPublic
	if motd != "" {
		p.MOTD = motd
	}

	jsonData, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(profilePath, fileName), jsonData, 0644)
}

func RemoveProfile(streamKey string) (bool, error) {
	if !isValidStreamKey(streamKey) {
		log.Println("Authorization: Remove profile failed due to invalid streamkey", streamKey)
		return false, fmt.Errorf("streamkey has invalid characters, only numbers, letters, dash and underscore allowed")
	}

	fileName, _ := getProfileFileNameByStreamKey(streamKey)
	if fileName == "" {
		log.Println("Authorization: RemoveProfile could not find", streamKey)
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
		if adhoc, adhocErr := getAdhocProfile(bearerToken); adhocErr == nil {
			return &PublicProfile{
				StreamKey: bearerToken,
				IsPublic:  adhoc.IsPublic,
				MOTD:      adhoc.MOTD,
			}, nil
		}

		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(profilePath, fileName))
	if err != nil {
		return nil, err
	}

	var profile profile
	if err := json.Unmarshal(data, &profile); err != nil {
		log.Println("Authorization: File", bearerToken, "could not read. File may be corrupt.")
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
		if adhoc, adhocErr := getAdhocProfile(bearerToken); adhocErr == nil {
			return &PersonalProfile{
				StreamKey: bearerToken,
				IsPublic:  adhoc.IsPublic,
				MOTD:      adhoc.MOTD,
			}, nil
		}

		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(profilePath, fileName))
	if err != nil {
		return nil, err
	}

	var profile profile
	if err := json.Unmarshal(data, &profile); err != nil {
		log.Println("Authorization: File", bearerToken, "could not read. File may be corrupt.")
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
		log.Println("Authorization: Error reading profile directory", err)
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

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
