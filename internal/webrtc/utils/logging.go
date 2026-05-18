package utils

import (
	"log/slog"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
)

func DebugOutputOffer(offer string) string {
	if strings.EqualFold(os.Getenv(environment.DebugPrintOffer), "true") {
		slog.Info("Offer", "sdp", offer)
	}

	return offer
}

func DebugOutputAnswer(answer string) string {
	if strings.EqualFold(os.Getenv(environment.DebugPrintAnswer), "true") {
		slog.Info("Answer", "sdp", answer)
	}

	return answer
}
