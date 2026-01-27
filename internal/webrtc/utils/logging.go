package utils

import (
	"log"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
)

func DebugOutputOffer(offer string) string {
	if strings.EqualFold(os.Getenv(environment.DEBUG_PRINT_OFFER), "true") {
		log.Println(offer)
	}

	return offer
}

func DebugOutputAnswer(answer string) string {
	if strings.EqualFold(os.Getenv(environment.DEBUG_PRINT_ANSWER), "true") {
		log.Println(answer)
	}

	return answer
}
