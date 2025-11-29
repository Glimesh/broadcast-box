package webrtc

import (
	"log"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/peerconnection"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

// TODO: Move peer connection inside of Whep struct, like it is done in Whip session
func WHEP(offer string, streamKey string) (string, string, error) {
	utils.DebugOutputOffer(offer)

	profile := authorization.PublicProfile{
		StreamKey: streamKey,
	}

	whipSession, err := session.SessionManager.GetOrAddStream(profile, false)
	if err != nil {
		return "", "", err
	}

	whepSessionId := uuid.New().String()

	peerConnection, err := peerconnection.CreateWhepPeerConnection()
	if err != nil {
		return "", "", err
	}

	audioTrack, videoTrack := codecs.GetDefaultTracks(streamKey)

	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		return "", "", err
	}

	videoRtcpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		return "", "", err
	}

	peerconnection.RegisterWhepHandlers(whipSession, peerConnection, whepSessionId)

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

	<-gatherComplete
	log.Println("WhepSession.GatheringCompletePromise: Completed Gathering for", streamKey)

	session.SessionManager.AddWhepSession(whepSessionId, whipSession, peerConnection, audioTrack, videoTrack, videoRtcpSender)

	return utils.DebugOutputAnswer(utils.AppendCandidateToAnswer(peerConnection.LocalDescription().SDP)), whepSessionId, nil
}
