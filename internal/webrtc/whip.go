package webrtc

import (
	"errors"
	"io"
	"log"
	"strings"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

func WHIP(offer, streamKey string) (string, error) {
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", err
	}

	audioTrack, videoTrack, pliChan, err := getTracksForStream(streamKey)
	if err != nil {
		return "", err
	}

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
		var localTrack *webrtc.TrackLocalStaticRTP
		if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
			localTrack = audioTrack
		} else {
			localTrack = videoTrack

			go func() {
				for range pliChan {
					if sendErr := peerConnection.WriteRTCP([]rtcp.Packet{
						&rtcp.PictureLossIndication{
							MediaSSRC: uint32(remoteTrack.SSRC()),
						},
					}); sendErr != nil {
						return
					}
				}
			}()
		}

		rtpBuf := make([]byte, 1500)
		for {
			rtpRead, _, readErr := remoteTrack.Read(rtpBuf)
			switch {
			case errors.Is(readErr, io.EOF):
				return
			case readErr != nil:
				log.Println(readErr)
				return
			}

			if _, writeErr := localTrack.Write(rtpBuf[:rtpRead]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
				log.Println(writeErr)
				return
			}
		}
	})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
				return
			}
		}
	})

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
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
