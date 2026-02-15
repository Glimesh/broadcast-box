package utils

import "github.com/pion/sdp/v3"

func ValidateOffer(offer string) error {
	var parsed sdp.SessionDescription
	return parsed.Unmarshal([]byte(offer))
}
