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
func WHIP(offer string, profile authorization.PublicProfile) (sdp string, sessionID string, err error) {
	log.Println("WHIP.Offer.Requested", profile.StreamKey, profile.MOTD)

	if err := utils.ValidateOffer(offer); err != nil {
		return "", "", errors.New("invalid offer: " + err.Error())
	}

	session, err := manager.SessionsManager.GetOrAddSession(profile, true)
	if err != nil {
		return "", "", err
	}

	peerConnection, err := peerconnection.CreateWHIPPeerConnection(offer)
	if err != nil || peerConnection == nil {
		log.Println("WHIP.CreateWHIPPeerConnection.Failed", err)
		if peerConnection != nil {
			if closeErr := peerConnection.Close(); closeErr != nil {
				log.Println("WHIP.CreateWHIPPeerConnection.Close.Failed", closeErr)
			}
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
	sessionID = host.ID
	err = nil
	log.Println("WHIP.Offer.Accepted", profile.StreamKey, profile.MOTD)
	return
}
