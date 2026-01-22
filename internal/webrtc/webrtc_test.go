package webrtc

import (
	"context"
	"strings"
	"testing"

	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/require"
)

func TestICETrickle(t *testing.T) {
	Configure()
	localTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion",
	)
	require.NoError(t, err)

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	require.NoError(t, err)

	connectedCtx, connectedDone := context.WithCancel(context.TODO())
	peerConnection.OnConnectionStateChange(func(c webrtc.PeerConnectionState) {
		if c == webrtc.PeerConnectionStateConnected {
			connectedDone()
		}
	})

	gatheredCtx, gatheredDone := context.WithCancel(context.TODO())
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			gatheredDone()
		}
	})

	_, err = peerConnection.AddTrack(localTrack)
	require.NoError(t, err)

	offer, err := peerConnection.CreateOffer(nil)
	require.NoError(t, err)
	require.NoError(t, peerConnection.SetLocalDescription(offer))

	answer, err := WHIP(offer.SDP, testStreamKey)
	require.NoError(t, err)

	noCandidateAnswer := ""
	for _, l := range strings.Split(answer, "\n") {
		if !strings.HasPrefix(l, "a=candidate:") {
			noCandidateAnswer += l + "\n"
		}
	}

	require.NoError(t, peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  noCandidateAnswer,
	}))

	<-gatheredCtx.Done()
	require.NoError(t, HandlePatch(testStreamKey, peerConnection.LocalDescription().SDP, true))

	<-connectedCtx.Done()
}
