package authorization

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"os"
	"strconv"
	"time"
)

func GetTURNCredentials() (username string, credentials string) {
	turnAuthKey := os.Getenv("TURN_SERVER_AUTH_SECRET")

	if turnAuthKey == "" {
		return "BroadcastBox", "BroadcastBox"
	}

	timestamp := time.Now().Unix() + 3600
	username = strconv.FormatInt(timestamp, 10)
	secret := hmac.New(sha1.New, []byte(turnAuthKey))
	secret.Write([]byte(username))
	rawPassword := secret.Sum(nil)

	return username, base64.StdEncoding.EncodeToString(rawPassword)
}
