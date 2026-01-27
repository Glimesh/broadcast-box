package helpers

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Resolve Bearer token.
// This supports both a B64 token as well as a regular ASCII token to allow for
// using special characters for stream keys that are not tokens
func ResolveBearerToken(authHeader string) (result string) {
	const bearerPrefix = "Bearer "

	// Cut the prefix
	if auth, ok := strings.CutPrefix(authHeader, bearerPrefix); ok {

		// Check for valid b64
		if base64String, err := getBase64String(auth); err == nil {
			return strings.ReplaceAll(base64String, " ", "_")

			// Invalid, handle as unicode
		} else {
			re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

			auth = strings.TrimSpace(auth)
			auth = strings.ReplaceAll(auth, " ", "_")
			if re.MatchString(auth) {
				return strings.ReplaceAll(auth, " ", "_")
			}
		}
	}

	return ""
}

// In case the bearer token is encoded in Base64, it can be resolved before return. This
// allows for UTF-8 character support in headers with bearer tokens
func getBase64String(token string) (result string, err error) {
	if !regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`).MatchString(token) {
		return "", fmt.Errorf("string is not valid base64")
	}

	output, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	if !utf8.Valid(output) {
		return token, nil
	}

	return string(output), nil
}
