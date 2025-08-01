package webrtc

import (
	"log"
	"slices"
	"strings"

	"github.com/glimesh/broadcast-box/internal/webrtc/stream"
	"github.com/glimesh/broadcast-box/internal/webrtc/whip"
	"github.com/pion/webrtc/v4"
)

func OnICEConnectionStateChangeHandler(
	peerConnection *webrtc.PeerConnection,
	isWhip bool,
	streamKey string,
	sessionId string,
) func(webrtc.ICEConnectionState) {

	return func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateFailed || state == webrtc.ICEConnectionStateClosed {
			if err := peerConnection.Close(); err != nil {
				log.Println("PeerConnection.OnICEConnectionStateChange.Error", err)
			}

			peerDisconnected(isWhip, streamKey, sessionId)
		}
	}
}

func OnTrackHandler(stream *stream.WhipSession, peerConnection *webrtc.PeerConnection) func(*webrtc.TrackRemote, *webrtc.RTPReceiver) {
	return func(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {

		if strings.HasPrefix(remoteTrack.Codec().MimeType, "audio") {
			// Handle audio stream
			whip.AudioWriter(remoteTrack, stream, peerConnection)
		} else {
			// Handle video stream
			whip.VideoWriter(remoteTrack, stream, peerConnection)
		}
	}
}

func peerDisconnected(isWhip bool, streamKey string, sessionId string) {
	WhipSessionsLock.Lock()
	defer WhipSessionsLock.Unlock()

	session, ok := WhipSessions[streamKey]
	if !ok {
		return
	}

	session.WhepSessionsLock.Lock()
	defer session.WhepSessionsLock.Unlock()

	if !isWhip {
		log.Println("Disconnected WHEP", streamKey, sessionId)
		delete(session.WhepSessions, sessionId)
	} else {

		session.AudioTracks = slices.DeleteFunc(session.AudioTracks, func(track *stream.AudioTrack) bool {
			return track.SessionId == session.SessionId
		})

		session.VideoTracks = slices.DeleteFunc(session.VideoTracks, func(track *stream.VideoTrack) bool {
			return track.SessionId == session.SessionId
		})

		if session.SessionId != sessionId {
			return
		}
		session.HasHost.Store(false)
	}

	if len(session.WhepSessions) != 0 || session.HasHost.Load() {
		return
	}

	session.ActiveContextCancel()
	delete(WhipSessions, streamKey)
}
