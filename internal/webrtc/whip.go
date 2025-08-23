package webrtc

import (
	"log"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/peerconnection"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"

	"github.com/glimesh/broadcast-box/internal/webrtc/session"
)

func WHIP(offer string, profile authorization.PublicProfile) (sdp string, sessionId string, err error) {
	log.Println("Incoming stream", profile.StreamKey, profile.MOTD)

	whipSessionId := uuid.New().String()

	peerConnection, err := peerconnection.CreateWhipPeerConnection()
	if err != nil {
		return "", "", err
	}

	stream, err := session.GetStream(profile, whipSessionId)
	if err != nil {
		return "", "", err
	}

	// PeerConnection Handlers
	peerconnection.RegisterHandlers(peerConnection, stream, true, whipSessionId)

	// Setup PeerConnection RemoteDescription
	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		log.Println("WHIP.PeerConnection.SetRemoteDescription", profile.StreamKey)
		return "", "", err
	}

	// TODO: Should this be replace with trickle from the peerConnection.OnICECandidate() ?
	gatheringCompleteResult := webrtc.GatheringCompletePromise(peerConnection)

	peerConnectionAnswer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("WHIP.PeerConnection.CreateAnswer.Error", profile.StreamKey, err)
		return "", "", err
	}

	if err := peerConnection.SetLocalDescription(peerConnectionAnswer); err != nil {
		log.Println("WHIP.PeerConnection.SetLocalDescription.Error", profile.StreamKey, err)
		return "", "", err
	}

	<-gatheringCompleteResult

	// Start sending out stream status at interval
	go session.StartWhipSessionLoop(stream)

	return utils.DebugOutputAnswer(utils.AppendAnswer(peerConnection.LocalDescription().SDP)), whipSessionId, nil
}
