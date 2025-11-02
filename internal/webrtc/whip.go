package webrtc

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/peerconnection"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

func WHIP(offer string, profile authorization.PublicProfile) (sdp string, sessionId string, err error) {
	log.Println("Incoming stream", profile.StreamKey, profile.MOTD)

	whipSession, err := session.SessionManager.GetOrAddStream(profile, true)
	if err != nil {
		return "", "", err
	}

	peerConnection, err := peerconnection.CreateWhipPeerConnection(offer)
	if err != nil {
		log.Println("WHIP.CreateWhipPeerConnection.Failed", err)
		whipSession.ActiveContextCancel()
		return "", "", err
	}

	whipSession.AddPeerConnection(peerConnection)
	peerconnection.RegisterWhipHandlers(whipSession, peerConnection, whipSession.SessionId)

	go whipSession.StartWhipSessionStatusLoop()
	go whipSession.SnapShot()

	sdp = utils.DebugOutputAnswer(utils.AppendCandidateToAnswer(peerConnection.LocalDescription().SDP))
	sessionId = whipSession.SessionId
	err = nil
	return
}
