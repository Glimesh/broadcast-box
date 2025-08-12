package utils

import (
	"log"
	"os"
	"strings"
)

func DebugOutputOffer(offer string) string {
	if strings.EqualFold(os.Getenv("DEBUG_PRINT_OFFER"), "true") {
		log.Println(offer)
	}

	return offer
}

func DebugOutputAnswer(answer string) string {
	if strings.EqualFold(os.Getenv("DEBUG_PRINT_ANSWER"), "true") {
		log.Println(answer)
	}

	return answer
}
