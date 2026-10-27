package utils

import (
	"log"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
)

func DebugOutputOffer(offer string) string {
	if strings.EqualFold(os.Getenv(environment.DebugPrintOffer), "true") {
		log.Println(offer)
	}

	return offer
}

func DebugOutputAnswer(answer string) string {
	if strings.EqualFold(os.Getenv(environment.DebugPrintAnswer), "true") {
		log.Println(answer)
	}

	return answer
}
