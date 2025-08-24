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
	STREAM_POLICY_ANYONE        = "ANYONE"
	STREAM_POLICY_WITH_RESERVED = "ANYONE_WITH_RESERVED"
	STREAM_POLICY_RESERVED_ONLY = "RESERVED"
)

func isValidStreamKey(streamKey string) bool {
	regExp := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return regExp.MatchString(streamKey)
}
func CreateProfile(streamKey string) (string, error) {

	if isValidStreamKey(streamKey) != true {
		log.Println("Authorization: Create profile failed due to invalid streamkey", streamKey)
		return "", fmt.Errorf("streamkey has invalid characters, only numbers, letters, dash and underscore allowed")
	}

	profilePath := os.Getenv(environment.STREAM_PROFILE_PATH)
	assureProfilePath()

	if hasExistingStreamKey(streamKey) {
		return "", fmt.Errorf("%s", "A profile with the stream key "+streamKey+" already exists")
	}

	token := generateToken()

	fileName := streamKey + "_" + token
	profileFilePath := filepath.Join(profilePath, fileName)
	profile := Profile{
		FileName: fileName,
		IsPublic: true,
		MOTD:     "Welcome to my stream!",
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

func RemoveProfile(streamKey string) (bool, error) {
	if isValidStreamKey(streamKey) != true {
		log.Println("Authorization: Remove profile failed due to invalid streamkey", streamKey)
		return false, fmt.Errorf("streamkey has invalid characters, only numbers, letters, dash and underscore allowed")
	}

	fileName, _ := getProfileFileNameByStreamKey(streamKey)
	if fileName == "" {
		log.Println("Authorization: RemoveProfile could not find", streamKey)
		return false, fmt.Errorf("Profile could not be found")
	}

	profilePath := os.Getenv(environment.STREAM_PROFILE_PATH)
	os.Remove(filepath.Join(profilePath, fileName))

	return true, nil
}

func GetPublicProfile(bearerToken string) (*PublicProfile, error) {
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

	var profile PublicProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		log.Println("Authorization: File", bearerToken, "could not read. File may be corrupt.")
		return nil, err
	}

	return &profile, nil
}

// Returns a slice of profiles intended for admin endpoints
func GetAdminProfilesAll() (profiles []AdminProfile, err error) {
	profilePath := os.Getenv(environment.STREAM_PROFILE_PATH)

	files, err := os.ReadDir(profilePath)
	if err != nil {
		log.Println("Authorization: Error reading profile directory", err)
		return nil, err
	}

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(profilePath, file.Name()))
		if err != nil {
			profiles = append(profiles, AdminProfile{
				StreamKey: file.Name(),
				IsPublic:  false,
				MOTD:      "Error reading profile from file: " + file.Name(),
			})

			continue
		}

		var profile Profile

		if err := json.Unmarshal(data, &profile); err != nil {
			profiles = append(profiles, AdminProfile{
				StreamKey: file.Name(),
				IsPublic:  false,
				MOTD:      "Invalid JSON in file" + file.Name(),
			})
			continue
		}

		profile.FileName = file.Name()
		profiles = append(profiles, *profile.AsAdminProfile())
	}

	return profiles, nil
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

func ResetProfileToken(streamKey string) error {
	fileName, _ := getProfileFileNameByStreamKey(streamKey)

	if fileName == "" {
		return fmt.Errorf("authorization: profile could not be found")
	}

	profilePath := os.Getenv(environment.STREAM_PROFILE_PATH)
	newFileName := streamKey + "_" + generateToken()
	currentPath := filepath.Join(profilePath, fileName)
	newPath := filepath.Join(profilePath, newFileName)

	if err := os.Rename(currentPath, newPath); err != nil {
		return fmt.Errorf("authorization: error updating profile token for %s", streamKey)
	}

	return nil
}
