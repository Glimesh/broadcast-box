package codecs

import (
	"log"

	"github.com/pion/webrtc/v4"
)

func RegisterCodecs(mediaEngine *webrtc.MediaEngine) {
	if err := registerVideoCodecs(mediaEngine); err != nil {
		log.Fatal(err)
	}

	if err := registerAudioCodecs(mediaEngine); err != nil {
		log.Fatal(err)
	}
}

func registerAudioCodecs(mediaEngine *webrtc.MediaEngine) []error {
	errors := []error{}
	for _, codec := range audioCodecs {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			log.Println("Error registering codec", codec.MimeType)
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		log.Println("Errors registering codecs", len(errors))
		return errors
	}

	return nil
}

func registerVideoCodecs(mediaEngine *webrtc.MediaEngine) []error {
	errors := []error{}
	for _, codec := range videoCodecs {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			log.Println("Error registering codec", codec.MimeType)
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		log.Println("Errors registering codecs", len(errors))
		return errors
	}

	return nil
}
