package webrtc

import (
	"context"
	"testing"
	"time"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/pion/webrtc/v4"
	"github.com/stretchr/testify/require"
)

const testStreamKey = "test"

var (
	testProfile = authorization.Profile{
		StreamKey: "test",
	}
)

func doesWHIPSessionExist() (ok bool) {
	session.WhipSessionsLock.Lock()
	defer session.WhipSessionsLock.Unlock()

	_, ok = session.WhipSessions[testStreamKey]
	return
}

// Asserts that a old PeerConnection doesn't destroy the new one
// when it disconnects
func TestReconnect(t *testing.T) {
	Setup()

	localTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion",
	)
	require.NoError(t, err)

	// Create the first WHIP Session
	firstPublisherConnected, firstPublisherConnectedDone := context.WithCancel(context.TODO())

	firstPublisher, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	require.NoError(t, err)

	firstPublisher.OnConnectionStateChange(func(c webrtc.PeerConnectionState) {
		if c == webrtc.PeerConnectionStateConnected {
			firstPublisherConnectedDone()

		}
	})

	_, err = firstPublisher.AddTrack(localTrack)
	require.NoError(t, err)

	offer, err := firstPublisher.CreateOffer(nil)
	require.NoError(t, err)
	require.NoError(t, firstPublisher.SetLocalDescription(offer))

	answer, _, err := WHIP(offer.SDP, testProfile)

	require.NoError(t, err)

	require.NoError(t, firstPublisher.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answer,
	}))

	require.True(t, doesWHIPSessionExist())
	<-firstPublisherConnected.Done()

	// Create the second WHIP Session
	secondPublisherConnected, secondPublisherConnectedDone := context.WithCancel(context.TODO())

	secondPublisher, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	require.NoError(t, err)

	secondPublisher.OnConnectionStateChange(func(c webrtc.PeerConnectionState) {
		if c == webrtc.PeerConnectionStateConnected {
			secondPublisherConnectedDone()

		}
	})

	_, err = secondPublisher.AddTrack(localTrack)
	require.NoError(t, err)

	offer, err = secondPublisher.CreateOffer(nil)
	require.NoError(t, err)
	require.NoError(t, secondPublisher.SetLocalDescription(offer))

	answer, _, err = WHIP(offer.SDP, testProfile)
	require.NoError(t, err)

	require.NoError(t, secondPublisher.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answer,
	}))

	require.True(t, doesWHIPSessionExist())
	<-secondPublisherConnected.Done()

	// Close the first WHIP Session, the session must still exist
	require.NoError(t, firstPublisher.Close())
	time.Sleep(time.Second)
	require.True(t, doesWHIPSessionExist())

	// Close the second WHIP Session, the session must be gone
	require.NoError(t, secondPublisher.Close())
	time.Sleep(time.Second)
	require.False(t, doesWHIPSessionExist())
}
