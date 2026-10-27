package webrtc

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/peerconnection"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func WHEP(offer string, streamKey string) (string, string, error) {
	utils.DebugOutputOffer(offer)

	profile := authorization.PublicProfile{
		StreamKey: streamKey,
	}

	session, err := manager.SessionsManager.GetOrAddSession(profile, false)
	if err != nil {
		return "", "", err
	}

	whepSessionID := uuid.New().String()

	peerConnection, err := peerconnection.CreateWHEPPeerConnection()
	if err != nil {
		return "", "", err
	}

	audioTrack, videoTrack := codecs.GetDefaultTracks(streamKey)

	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		return "", "", err
	}

	videoRTCPSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		return "", "", err
	}

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  offer,
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		return "", "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		return "", "", err
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", "", err
	}

	// TODO: Should this be before gatherComplete to assure registered events are triggered at correct time?
	if err := session.AddWHEP(
		whepSessionID,
		peerConnection,
		audioTrack,
		videoTrack,
		videoRTCPSender,
		func() {
			manager.SessionsManager.SendPLIByWHEPSessionID(whepSessionID)
		},
	); err != nil {
		return "", "", err
	}

	<-gatherComplete
	log.Println("WHEPSession.GatheringCompletePromise: Completed Gathering for", streamKey)

	return utils.DebugOutputAnswer(utils.AppendCandidateToAnswer(peerConnection.LocalDescription().SDP)),
		whepSessionID,
		nil
}
