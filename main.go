package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	status := statusUnconfigured
	if streamKey != "" {
		status = statusConfigured
	}

	if err := json.NewEncoder(w).Encode(&statusResponse{Status: status}); err != nil {
		log.Fatal(err)
	}
}

func configureHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	request := configureRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Fatal(err)
	}
	streamKey = request.StreamKey
}

func whipHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Fatal(err)
	}

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		fmt.Println("OnTrack")
	})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		fmt.Println(i)
	})

	offer, err := ioutil.ReadAll(r.Body)
	if err != nil {
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

func whepHandler(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)
}

func main() {
	h := http.NewServeMux()
	h.HandleFunc("/api/status", statusHandler)
	h.HandleFunc("/api/configure", configureHandler)
	h.HandleFunc("/api/whip", whipHandler)
	h.HandleFunc("/api/whep", whipHandler)

	s := &http.Server{
		Handler: h,
		Addr:    ":8080",
	}

	log.Fatal(s.ListenAndServe())
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}
