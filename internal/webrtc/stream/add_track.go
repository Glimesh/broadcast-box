package stream

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func AddAudioTrack(session *WhipSession, rid string, whepSessionsLock *sync.RWMutex) (*AudioTrack, error) {
	whepSessionsLock.Lock()
	defer whepSessionsLock.Unlock()

	log.Println("Adding AudioStream to", session.StreamKey, "(", rid, ")")
	for i := range session.AudioTracks {
		if rid == session.AudioTracks[i].Rid {
			return session.AudioTracks[i], nil
		}
	}

	uuid := uuid.New().String()

	track := &AudioTrack{
		Rid:       rid,
		SessionId: session.SessionId,
		Track: NewTrackMultiCodec(
			"audio-"+uuid,
			rid,
			session.StreamKey,
			webrtc.RTPCodecTypeVideo),
	}
	track.LastRecieved.Store(time.Time{})

	session.AudioTracks = append(session.AudioTracks, track)

	return track, nil
}

func AddVideoTrack(session *WhipSession, rid string, whepSessionsLock *sync.RWMutex) (*VideoTrack, error) {
	whepSessionsLock.Lock()
	defer whepSessionsLock.Unlock()

	log.Println("Adding VideoStream to", session.StreamKey, "(", rid, ")")
	for i := range session.VideoTracks {
		if rid == session.VideoTracks[i].Rid {
			log.Println("Found", rid)
			log.Println("Existing tracks", len(session.VideoTracks))

			return session.VideoTracks[i], nil
		}
	}

	uuid := uuid.New().String()

	track := &VideoTrack{
		Rid:       rid,
		SessionId: session.SessionId,
		Track: NewTrackMultiCodec(
			"video-"+uuid,
			rid,
			session.StreamKey,
			webrtc.RTPCodecTypeVideo),
	}
	track.LastRecieved.Store(time.Time{})

	session.VideoTracks = append(session.VideoTracks, track)

	return track, nil
}
