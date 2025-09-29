package helpers

import (
	"encoding/base64"
	"strings"
)

func ResolveBearerToken(authHeader string) string {
	const bearerPrefix = "Bearer "
	if result, ok := strings.CutPrefix(authHeader, bearerPrefix); ok {

		if base64String, err := getBase64String(strings.ReplaceAll(result, " ", "")); err == nil {
			return strings.ReplaceAll(base64String, " ", "_")
		}

		return strings.ReplaceAll(result, " ", "_")
	}

	return ""
}

// In case the bearer token is encoded in Base64, it can be resolved before return. This
// allows for UTF-8 character support in headers with bearer tokens
func getBase64String(token string) (result string, err error) {
	output, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	return string(output), err
}
