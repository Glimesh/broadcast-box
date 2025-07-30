package webrtc

import (
	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/stream"

	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

func WHEP(offer string, streamKey string) (string, string, error) {
	WhipSessionsLock.Lock()
	defer WhipSessionsLock.Unlock()

	profile := authorization.Profile{
		StreamKey: streamKey,
	}

	session, err := stream.GetStream(WhipSessions, profile, "")
	if err != nil {
		return "", "", err
	}

	whepSessionId := uuid.New().String()

	whepPeerConnection, err := CreatePeerConnection(apiWhep)
	if err != nil {
		return "", "", err
	}

	audioTrack := stream.NewTrackMultiCodec(
		"audio",
		"pion",
		streamKey,
		webrtc.RTPCodecTypeAudio)

	videoTrack := stream.NewTrackMultiCodec(
		"video",
		"pion",
		streamKey,
		webrtc.RTPCodecTypeVideo)

	_, err = whepPeerConnection.AddTrack(audioTrack)
	if err != nil {
		return "", "", err
	}

	rtpSender, err := whepPeerConnection.AddTrack(videoTrack)
	if err != nil {
		return "", "", err
	}

	whepPeerConnection.OnICEConnectionStateChange(
		OnICEConnectionStateChangeHandler(
			whepPeerConnection,
			false,
			streamKey,
			whepSessionId))

	go func() {
		for {
			rtcpPackets, _, rtcpErr := rtpSender.ReadRTCP()
			if rtcpErr != nil {
				return
			}

			for _, r := range rtcpPackets {
				if _, isPli := r.(*rtcp.PictureLossIndication); isPli {
					select {
					case session.PliChan <- true:
					default:
					}
				}
			}
		}
	}()

	if err := whepPeerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  offer,
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		return "", "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(whepPeerConnection)
	answer, err := whepPeerConnection.CreateAnswer(nil)

	if err != nil {
		return "", "", err
	} else if err = whepPeerConnection.SetLocalDescription(answer); err != nil {
		return "", "", err
	}

	<-gatherComplete

	session.WhepSessionsLock.Lock()
	defer session.WhepSessionsLock.Unlock()

	session.WhepSessions[whepSessionId] = &stream.WhepSession{
		AudioTrack:     audioTrack,
		VideoTrack:     videoTrack,
		AudioTimestamp: 5000,
		VideoTimestamp: 5000,
	}

	session.WhepSessions[whepSessionId].VideoLayerCurrent.Store("")
	session.WhepSessions[whepSessionId].AudioLayerCurrent.Store("")
	session.WhepSessions[whepSessionId].IsWaitingForKeyframe.Store(false)

	return appendAnswer(whepPeerConnection.LocalDescription().SDP), whepSessionId, nil

}
