package whip

import (
	"log"
	"slices"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

func (whipSession *WhipSession) AddAudioTrack(rid string, codec codecs.TrackCodeType) (*AudioTrack, error) {
	log.Println("WhipSession.AddAudioTrack:", whipSession.StreamKey, "(", rid, ")")
	whipSession.TracksLock.Lock()

	for i := range whipSession.AudioTracks {
		if rid == whipSession.AudioTracks[i].Rid {
			whipSession.TracksLock.Unlock()
			return whipSession.AudioTracks[i], nil
		}
	}

	track := &AudioTrack{
		Rid:       rid,
		SessionId: whipSession.SessionId,
		Track: codecs.CreateTrackMultiCodec(
			"audio-"+uuid.New().String(),
			rid,
			whipSession.StreamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastRecieved.Store(time.Time{})

	whipSession.AudioTracks = append(whipSession.AudioTracks, track)
	whipSession.TracksLock.Unlock()

	whipSession.HasHost.Store(true)

	return track, nil
}

func (whipSession *WhipSession) AddVideoTrack(rid string, codec codecs.TrackCodeType) (*VideoTrack, error) {
	log.Println("WhipSession.AddVideoTrack:", whipSession.StreamKey, "(", rid, ")")
	whipSession.TracksLock.Lock()
	for i := range whipSession.VideoTracks {
		if rid == whipSession.VideoTracks[i].Rid {
			whipSession.TracksLock.Unlock()
			return whipSession.VideoTracks[i], nil
		}
	}

	track := &VideoTrack{
		Rid:       rid,
		SessionId: whipSession.SessionId,
		Track: codecs.CreateTrackMultiCodec(
			"video-"+uuid.New().String(),
			rid,
			whipSession.StreamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastRecieved.Store(time.Time{})

	whipSession.VideoTracks = append(whipSession.VideoTracks, track)
	whipSession.TracksLock.Unlock()

	whipSession.HasHost.Store(true)

	return track, nil
}

// Remove Audio and Video tracks coming from the whip session id
func (whipSession *WhipSession) RemoveTracks() {
	log.Println("WhipSession.RemoveTracks:", whipSession.StreamKey)
	whipSession.TracksLock.Lock()

	whipSession.AudioTracks = slices.DeleteFunc(whipSession.AudioTracks, func(track *AudioTrack) bool {
		return track.SessionId == whipSession.SessionId
	})

	whipSession.VideoTracks = slices.DeleteFunc(whipSession.VideoTracks, func(track *VideoTrack) bool {
		return track.SessionId == whipSession.SessionId
	})

	// If no more tracks are available, notify that the stream has no host
	if len(whipSession.AudioTracks) == 0 && len(whipSession.VideoTracks) == 0 {
		whipSession.HasHost.Store(false)
	}

	whipSession.OnTrackChangeChannel <- struct{}{}

	whipSession.TracksLock.Unlock()
}
