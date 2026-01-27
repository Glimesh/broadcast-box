package webrtc

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/peerconnection"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

// Initialize WHIP session for incoming stream
func WHIP(offer string, profile authorization.PublicProfile) (sdp string, sessionId string, err error) {
	log.Println("WHIP.Offer.Requested", profile.StreamKey, profile.MOTD)

	session, err := manager.SessionsManager.GetOrAddSession(profile, true)
	if err != nil {
		return "", "", err
	}

	peerConnection, err := peerconnection.CreateWhipPeerConnection(offer)
	if err != nil {
		log.Println("WHIP.CreateWhipPeerConnection.Failed", err)
		peerConnection.Close()
		return "", "", err
	}

	if err := session.AddHost(peerConnection); err != nil {
		return "", "", err
	}

	sdp = utils.DebugOutputAnswer(utils.AppendCandidateToAnswer(peerConnection.LocalDescription().SDP))
	sessionId = session.Host.Id
	err = nil
	log.Println("WHIP.Offer.Accepted", profile.StreamKey, profile.MOTD)
	return
}
