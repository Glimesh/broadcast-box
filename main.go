package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pion/webrtc/v3"
)

type (
	statusResponse struct {
		Status string `json:"status"`
	}

	configureRequest struct {
		StreamKey string `json:"streamKey"`
	}
)

const (
	statusConfigured   = "configured"
	statusUnconfigured = "unconfigured"
)

var (
	streamKey = ""

	audioTrack, videoTrack = &webrtc.TrackLocalStaticRTP{}, &webrtc.TrackLocalStaticRTP{}
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := statusUnconfigured
	if streamKey != "" {
		status = statusConfigured
	}

	if err := json.NewEncoder(w).Encode(&statusResponse{Status: status}); err != nil {
		log.Fatal(err)
	}
}

func configureHandler(w http.ResponseWriter, r *http.Request) {
	request := configureRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Fatal(err)
	}
	streamKey = request.StreamKey
}

func whipHandler(w http.ResponseWriter, r *http.Request) {
	offer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Fatal(err)
	}

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		var localTrack *webrtc.TrackLocalStaticRTP
		if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
			localTrack = audioTrack
		} else {
			localTrack = videoTrack
		}

		rtpBuf := make([]byte, 1500)
		for {
			i, _, readErr := remoteTrack.Read(rtpBuf)
			switch {
			case errors.Is(readErr, io.EOF):
				return
			case readErr != nil:
				log.Fatal(readErr)
			}

			if _, writeErr := localTrack.Write(rtpBuf[:i]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
				log.Fatal(writeErr)
			}
		}
	})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Fatal(err)
			}
		}
	})

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		log.Fatal(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Fatal(err)
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Fatal(err)
	}
	<-gatherComplete

	fmt.Fprint(w, peerConnection.LocalDescription().SDP)
}

func whepHandler(w http.ResponseWriter, r *http.Request) {
	offer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Fatal(err)
	}

	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		log.Fatal(err)
	}

	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		log.Fatal(err)
	}

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		log.Fatal(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Fatal(err)
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Fatal(err)
	}
	<-gatherComplete

	fmt.Fprint(w, peerConnection.LocalDescription().SDP)
}

func main() {
	var err error
	if videoTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion"); err != nil {
		log.Fatal(err)
	} else if audioTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion"); err != nil {
		log.Fatal(err)
	}

	corsHandler := func(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			next(w, r)
		}
	}

	h := http.NewServeMux()
	h.HandleFunc("/api/status", corsHandler(statusHandler))
	h.HandleFunc("/api/configure", corsHandler(configureHandler))
	h.HandleFunc("/api/whip", corsHandler(whipHandler))
	h.HandleFunc("/api/whep", corsHandler(whepHandler))

	log.Fatal((&http.Server{
		Handler: h,
		Addr:    ":8080",
	}).ListenAndServe())
}
