package codecs

import (
	"log/slog"
	"os"

	"github.com/pion/webrtc/v4"
)

func RegisterCodecs(mediaEngine *webrtc.MediaEngine) {
	if err := registerVideoCodecs(mediaEngine); err != nil {
		slog.Error("Failed to register video codecs", "err", err)
		os.Exit(1)
	}

	if err := registerAudioCodecs(mediaEngine); err != nil {
		slog.Error("Failed to register audio codecs", "err", err)
		os.Exit(1)
	}
}

func registerAudioCodecs(mediaEngine *webrtc.MediaEngine) []error {
	errors := []error{}
	for _, codec := range audioCodecs {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			slog.Error("Error registering codec", "mimeType", codec.MimeType)
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		slog.Error("Errors registering codecs", "count", len(errors))
		return errors
	}

	return nil
}

func registerVideoCodecs(mediaEngine *webrtc.MediaEngine) []error {
	errors := []error{}
	for _, codec := range videoCodecs {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			slog.Error("Error registering codec", "mimeType", codec.MimeType)
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		slog.Error("Errors registering codecs", "count", len(errors))
		return errors
	}

	return nil
}
