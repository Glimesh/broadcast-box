package whip

import (
	"log"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

// Add a new AudioTrack to the Whip session
func (whipSession *WhipSession) AddAudioTrack(rid string, codec codecs.TrackCodeType) (*AudioTrack, error) {
	log.Println("WhipSession.AddAudioTrack:", whipSession.StreamKey, "(", rid, ")")
	whipSession.TracksLock.Lock()
	defer whipSession.TracksLock.Unlock()

	if existingTrack, ok := whipSession.AudioTracks[rid]; ok {
		return existingTrack, nil
	}

	track := &AudioTrack{
		Rid:                rid,
		SessionId:          whipSession.SessionId,
		TrackStreamChannel: make(chan codecs.TrackPacket, 1000),
		Track: codecs.CreateTrackMultiCodec(
			"audio-"+uuid.New().String(),
			rid,
			whipSession.StreamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastRecieved.Store(time.Time{})

	whipSession.AudioTracks[track.Rid] = track
	whipSession.HasHost.Store(true)

	return track, nil
}

// Add a new VideoTrack to the Whip session
func (whipSession *WhipSession) AddVideoTrack(rid string, codec codecs.TrackCodeType) (*VideoTrack, error) {
	log.Println("WhipSession.AddVideoTrack:", whipSession.StreamKey, "(", rid, ")")
	whipSession.TracksLock.Lock()
	defer whipSession.TracksLock.Unlock()

	if existingTrack, ok := whipSession.VideoTracks[rid]; ok {
		return existingTrack, nil
	}

	track := &VideoTrack{
		Rid:                rid,
		SessionId:          whipSession.SessionId,
		TrackStreamChannel: make(chan codecs.TrackPacket, 1000),
		Track: codecs.CreateTrackMultiCodec(
			"video-"+uuid.New().String(),
			rid,
			whipSession.StreamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastRecieved.Store(time.Time{})

	whipSession.VideoTracks[rid] = track
	whipSession.HasHost.Store(true)

	return track, nil
}

// Remove Audio and Video tracks coming from the whip session id
func (whipSession *WhipSession) RemoveTracks() {
	log.Println("WhipSession.RemoveTracks:", whipSession.StreamKey)
	whipSession.TracksLock.Lock()

	whipSession.AudioTracks = make(map[string]*AudioTrack)
	whipSession.VideoTracks = make(map[string]*VideoTrack)

	// If no more tracks are available, notify that the stream has no host
	if len(whipSession.AudioTracks) == 0 && len(whipSession.VideoTracks) == 0 {
		whipSession.HasHost.Store(false)
	}

	whipSession.OnTrackChangeChannel <- struct{}{}
	whipSession.TracksLock.Unlock()
}
