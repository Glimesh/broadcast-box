package webrtc

import (
	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/peerconnection"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/glimesh/broadcast-box/internal/webrtc/utils"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func WHEP(offer string, streamKey string) (string, string, error) {
	utils.DebugOutputOffer(offer)

	profile := authorization.PublicProfile{
		StreamKey: streamKey,
	}

	stream, err := session.GetStream(profile, "")

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

	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		return "", "", err
	}

	peerconnection.RegisterHandlers(peerConnection, stream, false, whepSessionId)

	go session.RtcpPacketReaderLoop(stream, rtpSender)

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

	session.WhipSessionsLock.Lock()
	stream.WhepSessionsLock.Lock()

	stream.WhepSessions[whepSessionId] = &session.WhepSession{
		AudioTrack:     audioTrack,
		VideoTrack:     videoTrack,
		AudioTimestamp: 5000,
		VideoTimestamp: 5000,
		SSEChannel:     make(chan any, 100),
	}

	var defaultAudioTrack = stream.GetHighestPrioritizedAudioTrack()
	var defaultVideoTrack = stream.GetHighestPrioritizedVideoTrack()

	whepSession := stream.WhepSessions[whepSessionId]
	whepSession.VideoLayerCurrent.Store(defaultVideoTrack)
	whepSession.AudioLayerCurrent.Store(defaultAudioTrack)
	whepSession.IsWaitingForKeyframe.Store(false)

	session.WhipSessionsLock.Unlock()
	stream.WhepSessionsLock.Unlock()

	// When WHEP is established, send initial messages to client
	go func() {
		whepSession.SSEChannel <- session.GetSessionStatsJsonString(stream)
		whepSession.SSEChannel <- session.GetAvailableLayersJsonString(stream)
		whepSession.SSEChannel <- session.GetWhepSessionStatus(whepSession)
	}()

	return utils.DebugOutputAnswer(utils.AppendAnswer(peerConnection.LocalDescription().SDP)), whepSessionId, nil
}
