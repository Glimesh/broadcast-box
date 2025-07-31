package webrtc

import (
	"log"
	"os"
	"strings"
)

func debugOutputOffer(offer string) string {
	if strings.EqualFold(os.Getenv("DEBUG_PRINT_OFFER"), "true") {
		log.Println(offer)
	}

	return offer
}

func debugOutputAnswer(answer string) string {
	if strings.EqualFold(os.Getenv("DEBUG_PRINT_ANSWER"), "true") {
		log.Println(answer)
	}

	return answer
}
