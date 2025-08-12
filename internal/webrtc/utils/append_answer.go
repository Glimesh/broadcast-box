package utils

import (
	"os"
	"strings"
)

func AppendAnswer(localDescriptionSFP string) string {
	if appendCandidate := os.Getenv("APPEND_CANDIDATE"); appendCandidate != "" {
		index := strings.Index(localDescriptionSFP, "a=end-of-candidates")
		localDescriptionSFP = localDescriptionSFP[:index] + appendCandidate + localDescriptionSFP[index:]
	}

	return localDescriptionSFP
}
