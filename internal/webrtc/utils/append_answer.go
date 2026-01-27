package utils

import (
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
)

// Appends a candidate to the list of candidates that are sent back to the client in the answer
func AppendCandidateToAnswer(localDescriptionSFP string) string {
	if appendCandidate := os.Getenv(environment.APPEND_CANDIDATE); appendCandidate != "" {
		index := strings.Index(localDescriptionSFP, "a=end-of-candidates")
		localDescriptionSFP = localDescriptionSFP[:index] + appendCandidate + localDescriptionSFP[index:]
	}

	return localDescriptionSFP
}
