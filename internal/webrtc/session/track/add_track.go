package track

import (
	"log"
	"sync"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func AddAudioTrack(stream *session.WhipSession, rid string, codec int, whepSessionsLock *sync.RWMutex) (*session.AudioTrack, error) {
	whepSessionsLock.Lock()
	defer whepSessionsLock.Unlock()

	for i := range stream.AudioTracks {
		if rid == stream.AudioTracks[i].Rid {
			return stream.AudioTracks[i], nil
		}
	}

	uuid := uuid.New().String()

	track := &session.AudioTrack{
		Rid:       rid,
		SessionId: stream.SessionId,
		Track: codecs.CreateTrackMultiCodec(
			"audio-"+uuid,
			rid,
			stream.StreamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastRecieved.Store(time.Time{})

	stream.TracksLock.Lock()
	stream.AudioTracks = append(stream.AudioTracks, track)
	stream.TracksLock.Unlock()

	return track, nil
}

func AddVideoTrack(stream *session.WhipSession, rid string, codec int, whepSessionsLock *sync.RWMutex) (*session.VideoTrack, error) {
	whepSessionsLock.Lock()
	defer whepSessionsLock.Unlock()

	log.Println("Adding VideoStream to", stream.StreamKey, "(", rid, ")")
	for i := range stream.VideoTracks {
		if rid == stream.VideoTracks[i].Rid {
			return stream.VideoTracks[i], nil
		}
	}

	uuid := uuid.New().String()

	track := &session.VideoTrack{
		Rid:       rid,
		SessionId: stream.SessionId,
		Track: codecs.CreateTrackMultiCodec(
			"video-"+uuid,
			rid,
			stream.StreamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastRecieved.Store(time.Time{})

	stream.TracksLock.Lock()
	stream.VideoTracks = append(stream.VideoTracks, track)
	stream.TracksLock.Unlock()

	return track, nil
}
