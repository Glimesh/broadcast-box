package webrtc

import (
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

func WHEP(offer, streamKey string) (string, error) {
	//nolint:all
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", err
	}

	audioTrack, videoTrack, pliChan, err := getTracksForStream(streamKey)
	if err != nil {
		return "", err
	}

	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		return "", err
	}

	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		return "", err
	}

	go func() {
		for {
			rtcpPackets, _, rtcpErr := rtpSender.ReadRTCP()
			if rtcpErr != nil {
				return
			}

			for _, r := range rtcpPackets {
				if _, isPLI := r.(*rtcp.PictureLossIndication); isPLI {
					select {
					case pliChan <- true:
					default:
					}
				}
			}
		}
	}()

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  offer,
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		return "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		return "", err
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}

	<-gatherComplete
	return peerConnection.LocalDescription().SDP, nil
}
