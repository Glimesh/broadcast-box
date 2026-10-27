package whip

import (
	"log"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

// Add a new AudioTrack to the WHIP session
func (w *WHIPSession) addAudioTrack(rid string, streamKey string, codec codecs.TrackCodeType) (*AudioTrack, error) {
	log.Println("WHIPSession.AddAudioTrack:", streamKey, "(", rid, ")")
	w.TracksLock.Lock()
	defer w.TracksLock.Unlock()

	if existingTrack, ok := w.AudioTracks[rid]; ok {
		return existingTrack, nil
	}

	track := &AudioTrack{
		Rid: rid,
		Track: codecs.CreateTrackMultiCodec(
			"audio-"+uuid.New().String(),
			rid,
			streamKey,
			webrtc.RTPCodecTypeAudio,
			codec),
	}
	track.LastReceived.Store(time.Time{})

	w.AudioTracks[track.Rid] = track

	return track, nil
}

// Add a new VideoTrack to the WHIP session
func (w *WHIPSession) addVideoTrack(rid string, streamKey string, codec codecs.TrackCodeType) (*VideoTrack, error) {
	log.Println("WHIPSession.AddVideoTrack:", "(", rid, ")")
	w.TracksLock.Lock()
	defer w.TracksLock.Unlock()

	if existingTrack, ok := w.VideoTracks[rid]; ok {
		return existingTrack, nil
	}

	track := &VideoTrack{
		Rid: rid,
		Track: codecs.CreateTrackMultiCodec(
			"video-"+uuid.New().String(),
			rid,
			streamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastReceived.Store(time.Time{})

	w.VideoTracks[rid] = track

	return track, nil
}

// Remove Audio and Video tracks coming from the whip session id
func (w *WHIPSession) RemoveTracks() {
	log.Println("WHIPSession.RemoveTracks")

	w.TracksLock.Lock()
	w.AudioTracks = make(map[string]*AudioTrack)
	w.VideoTracks = make(map[string]*VideoTrack)
	w.TracksLock.Unlock()
}
