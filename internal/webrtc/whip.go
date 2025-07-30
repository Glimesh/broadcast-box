package webrtc

import (
	"log"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/stream"
)

func WHIP(offer string, profile authorization.Profile) (string, error) {
	log.Println("Incoming stream", profile.StreamKey)

	whipSessionId := uuid.New().String()

	peerConnection, err := CreatePeerConnection(apiWhip)
	if err != nil {
		log.Println("Error creating PeerConnection")
		return "", err
	}

	WhipSessionsLock.Lock()
	defer WhipSessionsLock.Unlock()

	stream, err := stream.GetStream(WhipSessions, profile, whipSessionId)
	if err != nil {
		return "", err
	}

	// PeerConnection Handlers
	peerConnection.OnTrack(OnTrackHandler(stream, peerConnection))
	peerConnection.OnICEConnectionStateChange(
		OnICEConnectionStateChangeHandler(
			peerConnection,
			true,
			profile.StreamKey,
			whipSessionId))

	// Setup PeerConnection RemoteDescription
	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		log.Println("WHIP.PeerConnection.SetRemoteDescription", profile.StreamKey)
		return "", err
	}

	gatheringCompleteResult := webrtc.GatheringCompletePromise(peerConnection)

	peerConnectionAnswer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("WHIP.PeerConnection.CreateAnswer.Error", profile.StreamKey, err)
		return "", err
	}

	if err := peerConnection.SetLocalDescription(peerConnectionAnswer); err != nil {
		log.Println("WHIP.PeerConnection.SetLocalDescription.Error", profile.StreamKey, err)
		return "", err
	}

	<-gatheringCompleteResult
	return appendAnswer(peerConnection.LocalDescription().SDP), nil
}
