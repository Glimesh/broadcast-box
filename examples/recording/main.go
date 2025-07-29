package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264writer"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
)

type webhookPayload struct {
	Action      string            `json:"action"`
	IP          string            `json:"ip"`
	BearerToken string            `json:"bearerToken"`
	QueryParams map[string]string `json:"queryParams"`
	UserAgent   string            `json:"userAgent"`
}

type webhookResponse struct {
	StreamKey string `json:"streamKey"`
}

const (
	whepServerUrl  = "http://127.0.0.1:8080/api/whep"
	fileNameLength = 16
	readTimeout    = time.Second * 5
)

func startRecording(streamKey string) {
	s := webrtc.SettingEngine{}
	s.SetNetworkTypes([]webrtc.NetworkType{
		webrtc.NetworkTypeUDP4,
		webrtc.NetworkTypeUDP6,
		webrtc.NetworkTypeTCP4,
		webrtc.NetworkTypeTCP6,
	})

	peerConnection, err := webrtc.NewAPI(webrtc.WithSettingEngine(s)).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		panic(err)
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
		panic(err)
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
		panic(err)
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", whepServerUrl, bytes.NewBuffer([]byte(offer.SDP)))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+streamKey)
	req.Header.Set("Content-Type", "application/sdp")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close() // nolint

	if resp.StatusCode != 201 {
		panic(fmt.Sprintf("unexpected HTTP StatusCode %d", resp.StatusCode))
	}

	if resp.Header.Get("Content-Type") != "application/sdp" {
		panic(fmt.Sprintf("unexpected HTTP Content-Type %s", resp.Header.Get("Content-Type")))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	if err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  string(respBody),
	}); err != nil {
		panic(err)
	}

	prefix, audioWriter, videoWriter := createFiles()

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Recording %s Connection State has changed to %s \n", streamKey, state)
		switch state {
		case webrtc.PeerConnectionStateFailed:
			_ = peerConnection.Close()
		case webrtc.PeerConnectionStateClosed:
			_ = audioWriter.Close()
			_ = videoWriter.Close()
		}
	})

	peerConnection.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		if strings.EqualFold(track.Codec().MimeType, webrtc.MimeTypeOpus) {
			fmt.Printf("Got Opus track, saving to disk as %s.ogg (48 kHz, 2 channels)\n", prefix)
			saveToDisk(audioWriter, track)
		} else if strings.EqualFold(track.Codec().MimeType, webrtc.MimeTypeH264) {
			fmt.Printf("Got H264 track, saving to disk as %s.h264\n", prefix)
			saveToDisk(videoWriter, track)
		}
	})
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
			return
		}

		var payload webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(webhookResponse{StreamKey: payload.BearerToken}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if payload.Action == "whip-connect" {
			startRecording(payload.BearerToken)
		}
	})

	log.Println("Server listening on port 8081")
	if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func createFiles() (string, media.Writer, media.Writer) {
	prefix := make([]rune, fileNameLength)
	for i := range prefix {
		prefix[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	audioFile, err := oggwriter.New(string(prefix)+".ogg", 48000, 2)
	if err != nil {
		panic(err)
	}
	videoFile, err := h264writer.New(string(prefix) + ".h264")
	if err != nil {
		panic(err)
	}

	return string(prefix), audioFile, videoFile
}

func saveToDisk(writer media.Writer, track *webrtc.TrackRemote) {
	defer func() {
		if err := writer.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		_ = track.SetReadDeadline(time.Now().Add(readTimeout))
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := writer.WriteRTP(rtpPacket); err != nil {
			fmt.Println(err)
			return
		}
	}
}
