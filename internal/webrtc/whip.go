package webrtc

import (
	"errors"
	"log"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/peerconnection"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
)

// Initialize WHIP session for incoming stream
func WHIP(offer string, profile authorization.PublicProfile) (sdp string, sessionId string, err error) {
	log.Println("WHIP.Offer.Requested", profile.StreamKey, profile.MOTD)

	if err := utils.ValidateOffer(offer); err != nil {
		return "", "", errors.New("invalid offer: " + err.Error())
	}

	session, err := manager.SessionsManager.GetOrAddSession(profile, true)
	if err != nil {
		return "", "", err
	}

	peerConnection, err := peerconnection.CreateWhipPeerConnection(offer)
	if err != nil || peerConnection == nil {
		log.Println("WHIP.CreateWhipPeerConnection.Failed", err)
		if peerConnection != nil {
			peerConnection.Close()
		}
		return "", "", err
	}

	if err := session.AddHost(peerConnection); err != nil {
		return "", "", err
	}

	host := session.Host.Load()
	if host == nil {
		return "", "", errors.New("host session not available")
	}

	sdp = utils.DebugOutputAnswer(utils.AppendCandidateToAnswer(peerConnection.LocalDescription().SDP))
	sessionId = host.Id
	err = nil
	log.Println("WHIP.Offer.Accepted", profile.StreamKey, profile.MOTD)
	return
}
