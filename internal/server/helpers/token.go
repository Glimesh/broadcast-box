package helpers

import (
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// Resolve Bearer token.
// This supports both a B64 token as well as a regular ASCII token to allow for
// using special characters for stream keys that are not tokens
func ResolveBearerToken(authHeader string) string {
	const bearerPrefix = "Bearer "
	if result, ok := strings.CutPrefix(authHeader, bearerPrefix); ok {

		if base64String, err := getBase64String(result); err == nil {
			return strings.ReplaceAll(base64String, " ", "_")
		}

		return strings.ReplaceAll(result, " ", "_")
	}

	return ""
}

// In case the bearer token is encoded in Base64, it can be resolved before return. This
// allows for UTF-8 character support in headers with bearer tokens
func getBase64String(token string) (result string, err error) {
	log.Println("Checking B64 for string", token)
	if !regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`).MatchString(token) {
		return "", fmt.Errorf("string is not valid base64")
	}

	output, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	return string(output), err
}
