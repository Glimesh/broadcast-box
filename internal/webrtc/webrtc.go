package webrtc

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
)

const (
	videoTrackLabelDefault = "default"
)

type (
	stream struct {
		audioTrack       *webrtc.TrackLocalStaticRTP
		videoTrackLabels []string
		pliChan          chan any
		whepSessionsLock sync.RWMutex
		whepSessions     map[string]*whepSession
	}
)

var (
	streamMap     map[string]*stream
	streamMapLock sync.Mutex
	api           *webrtc.API
)

func getStream(streamKey string) (*stream, error) {
	foundStream, ok := streamMap[streamKey]
	if !ok {
		audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
		if err != nil {
			return nil, err
		}

		foundStream = &stream{
			audioTrack:   audioTrack,
			pliChan:      make(chan any, 50),
			whepSessions: map[string]*whepSession{},
		}
		streamMap[streamKey] = foundStream
	}

	return foundStream, nil
}

func deleteStream(streamKey string) {
	streamMapLock.Lock()
	defer streamMapLock.Unlock()

	delete(streamMap, streamKey)
}

func addTrack(stream *stream, rid string) error {
	streamMapLock.Lock()
	defer streamMapLock.Unlock()

	for i := range stream.videoTrackLabels {
		if rid == stream.videoTrackLabels[i] {
			return nil
		}
	}

	stream.videoTrackLabels = append(stream.videoTrackLabels, rid)
	return nil
}

func getPublicIP() string {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	ip := struct {
		Query string
	}{}
	if err = json.Unmarshal(body, &ip); err != nil {
		log.Fatal(err)
	}

	if ip.Query == "" {
		log.Fatal("Query entry was not populated")
	}

	return ip.Query
}

func populateSettingEngine(settingEngine *webrtc.SettingEngine) {
	NAT1To1IPs := []string{}

	if os.Getenv("INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP") != "" {
		NAT1To1IPs = append(NAT1To1IPs, getPublicIP())
	}

	if os.Getenv("NAT_1_TO_1_IP") != "" {
		NAT1To1IPs = append(NAT1To1IPs, os.Getenv("NAT_1_TO_1_IP"))
	}

	if len(NAT1To1IPs) != 0 {
		settingEngine.SetNAT1To1IPs(NAT1To1IPs, webrtc.ICECandidateTypeHost)
	}

	if os.Getenv("INTERFACE_FILTER") != "" {
		settingEngine.SetInterfaceFilter(func(i string) bool {
			return i == os.Getenv("INTERFACE_FILTER")
		})
	}

	if os.Getenv("UDP_MUX_PORT") != "" {
		udpPort, err := strconv.Atoi(os.Getenv("UDP_MUX_PORT"))
		if err != nil {
			log.Fatal(err)
		}

		udpMux, err := ice.NewMultiUDPMuxFromPort(udpPort)
		if err != nil {
			log.Fatal(err)
		}

		settingEngine.SetICEUDPMux(udpMux)
	}

	if os.Getenv("TCP_MUX_ADDRESS") != "" {
		tcpAddr, err := net.ResolveTCPAddr("udp", os.Getenv("TCP_MUX_ADDRESS"))
		if err != nil {
			log.Fatal(err)
		}

		tcpListener, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			log.Fatal(err)
		}

		settingEngine.SetICETCPMux(webrtc.NewICETCPMux(nil, tcpListener, 8))
	}
}

func populateMediaEngine(m *webrtc.MediaEngine) error {
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeOpus, 48000, 2, "minptime=10;useinbandfec=1", nil},
			PayloadType:        111,
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			return err
		}
	}

	// nolint
	videoRTCPFeedback := []webrtc.RTCPFeedback{{"goog-remb", ""}, {"ccm", "fir"}, {"nack", ""}, {"nack", "pli"}}
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeH264, 90000, 0, "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f", videoRTCPFeedback},
			PayloadType:        102,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{"video/rtx", 90000, 0, "apt=102", nil},
			PayloadType:        121,
		},

		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeH264, 90000, 0, "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f", videoRTCPFeedback},
			PayloadType:        127,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{"video/rtx", 90000, 0, "apt=127", nil},
			PayloadType:        120,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeH264, 90000, 0, "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f", videoRTCPFeedback},
			PayloadType:        125,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{"video/rtx", 90000, 0, "apt=125", nil},
			PayloadType:        107,
		},

		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeH264, 90000, 0, "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42e01f", videoRTCPFeedback},
			PayloadType:        108,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{"video/rtx", 90000, 0, "apt=108", nil},
			PayloadType:        109,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeH264, 90000, 0, "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f", videoRTCPFeedback},
			PayloadType:        127,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{"video/rtx", 90000, 0, "apt=127", nil},
			PayloadType:        120,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeH264, 90000, 0, "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=640032", videoRTCPFeedback},
			PayloadType:        123,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{"video/rtx", 90000, 0, "apt=123", nil},
			PayloadType:        118,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeAV1, 90000, 0, "", videoRTCPFeedback},
			PayloadType:        124,
		},
		{
			// nolint
			RTPCodecCapability: webrtc.RTPCodecCapability{"video/rtx", 90000, 0, "apt=124", nil},
			PayloadType:        125,
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}

	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
	} {
		if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}

	return nil
}

func Configure() {
	streamMap = map[string]*stream{}

	mediaEngine := &webrtc.MediaEngine{}
	if err := populateMediaEngine(mediaEngine); err != nil {
		panic(err)
	}

	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		log.Fatal(err)
	}

	settingEngine := webrtc.SettingEngine{}
	populateSettingEngine(&settingEngine)

	api = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
		webrtc.WithSettingEngine(settingEngine),
	)
}
