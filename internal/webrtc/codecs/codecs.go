package codecs

import (
	"fmt"
	"strings"

	"github.com/pion/webrtc/v4"
)

const (
	AudioTrackLabelDefault = "Audio"
	VideoTrackLabelDefault = "Video"

	VideoTrackCodecH264 = iota + 1
	// VideoTrackCodecH264 videoTrackCodec = iota + 1
	VideoTrackCodecH265
	VideoTrackCodecVP8
	VideoTrackCodecVP9
	VideoTrackCodecAV1

	AudioTrackCodecOpus
)

var VideoRTCPFeedback = []webrtc.RTCPFeedback{
	{Type: "goog-remb", Parameter: ""},
	{Type: "ccm", Parameter: "fir"},
	{Type: "nack", Parameter: ""},
	{Type: "nack", Parameter: "pli"},
}

var VideoCodecs = []webrtc.RTPCodecParameters{
	{
		PayloadType: 96,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 102,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 103,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 104,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 106,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 108,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42e01f",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 39,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH264,
			ClockRate:    90000,
			SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=4d001f",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 45,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeAV1,
			ClockRate:    90000,
			SDPFmtpLine:  "",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 98,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP9,
			ClockRate:    90000,
			SDPFmtpLine:  "profile-id=0",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 100,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeVP9,
			ClockRate:    90000,
			SDPFmtpLine:  "profile-id=2",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
	{
		PayloadType: 113,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH265,
			ClockRate:    90000,
			SDPFmtpLine:  "level-id=93;profile-id=1;tier-flag=0;tx-mode=SRST",
			RTCPFeedback: VideoRTCPFeedback,
		},
	},
}

var AudioCodecs = []webrtc.RTPCodecParameters{
	{
		PayloadType: 111,
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeOpus,
			ClockRate:    48_000,
			Channels:     2,
			SDPFmtpLine:  "minptime=10;useinbandfec=1",
			RTCPFeedback: nil,
		},
	},
}

func GetVideoTrackCodec(codec string) int {
	lowerCase := strings.ToLower(codec)

	switch {
	case strings.Contains(lowerCase, strings.ToLower(webrtc.MimeTypeH264)):
		return VideoTrackCodecH264

	case strings.Contains(lowerCase, strings.ToLower(webrtc.MimeTypeVP8)):
		return VideoTrackCodecVP8

	case strings.Contains(lowerCase, strings.ToLower(webrtc.MimeTypeVP9)):
		return VideoTrackCodecVP9

	case strings.Contains(lowerCase, strings.ToLower(webrtc.MimeTypeAV1)):
		return VideoTrackCodecAV1

	case strings.Contains(lowerCase, strings.ToLower(webrtc.MimeTypeH265)):
		return VideoTrackCodecH265
	}

	return 0
}

func GetVideoTrackCodecParameters(codec int) (webrtc.RTPCodecParameters, error) {

	for audioCodec := range AudioCodecs {
		lookup := GetVideoTrackCodec(AudioCodecs[audioCodec].MimeType)
		if lookup == codec {
			return AudioCodecs[audioCodec], nil
		}
	}

	for videoCodec := range VideoCodecs {
		lookup := GetVideoTrackCodec(VideoCodecs[videoCodec].MimeType)
		if lookup == codec {
			return VideoCodecs[videoCodec], nil
		}
	}

	return webrtc.RTPCodecParameters{}, fmt.Errorf("Could not find a matching codec")
}

func GetAudioTrackCodec(codec string) int {
	lowerCase := strings.ToLower(codec)

	switch {
	case strings.Contains(lowerCase, strings.ToLower(webrtc.MimeTypeOpus)):
		return AudioTrackCodecOpus
	}

	return 0
}
