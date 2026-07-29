package whip

import (
	"sync"
	"sync/atomic"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

type (
	WHIPSession struct {
		ID                 string
		PeerConnection     *webrtc.PeerConnection
		closeOnce          sync.Once
		onClosed           func()
		PeerConnectionLock sync.RWMutex

		// Protects AudioTrack, VideoTracks
		TracksLock  sync.RWMutex
		VideoTracks map[string]*VideoTrack
		AudioTracks map[string]*AudioTrack

		// TODO: WHEPSessionsSnapshot should contain serializable state, not runtime references.
		WHEPSessionsSnapshot atomic.Value
	}

	VideoTrack struct {
		Rid             string
		Priority        int
		Bitrate         atomic.Uint64
		PacketsReceived atomic.Uint64
		PacketsDropped  atomic.Uint64
		LastReceived    atomic.Value
		LastKeyFrame    atomic.Value
		MediaSSRC       atomic.Uint32
		Track           *codecs.TrackMultiCodec
	}
	AudioTrack struct {
		Rid             string
		Priority        int
		PacketsReceived atomic.Uint64
		PacketsDropped  atomic.Uint64
		LastReceived    atomic.Value
		Track           *codecs.TrackMultiCodec
	}
)
