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
		Rid:       rid,
		SessionId: whipSession.SessionId,
		Track: codecs.CreateTrackMultiCodec(
			"audio-"+uuid.New().String(),
			rid,
			whipSession.StreamKey,
			webrtc.RTPCodecTypeAudio,
			codec),
	}
	track.LastReceived.Store(time.Time{})

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
		Rid:       rid,
		SessionId: whipSession.SessionId,
		Track: codecs.CreateTrackMultiCodec(
			"video-"+uuid.New().String(),
			rid,
			whipSession.StreamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastReceived.Store(time.Time{})

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

	whipSession.HasHost.Store(false)
	whipSession.TracksLock.Unlock()

	whipSession.OnTrackChangeChannel <- struct{}{}
}

// Get highest prioritized audio track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (whipSession *WhipSession) GetHighestPrioritizedAudioTrack() string {
	if len(whipSession.AudioTracks) == 0 {
		log.Println("No Audio tracks was found for", whipSession.StreamKey)
		return ""
	}

	whipSession.TracksLock.RLock()
	var highestPriorityAudioTrack *AudioTrack
	for _, trackPriority := range whipSession.AudioTracks {
		if highestPriorityAudioTrack == nil {
			highestPriorityAudioTrack = trackPriority
			continue
		}

		if trackPriority.Priority < highestPriorityAudioTrack.Priority {
			highestPriorityAudioTrack = trackPriority
		}
	}
	whipSession.TracksLock.RUnlock()

	if highestPriorityAudioTrack == nil {
		return ""
	}

	return highestPriorityAudioTrack.Rid

}

// Get highest prioritized video track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (whipSession *WhipSession) GetHighestPrioritizedVideoTrack() string {
	if len(whipSession.VideoTracks) == 0 {
		log.Println("No Video tracks was found for", whipSession.StreamKey)
	}

	var highestPriorityVideoTrack *VideoTrack

	whipSession.TracksLock.RLock()
	for _, trackPriority := range whipSession.VideoTracks {
		if highestPriorityVideoTrack == nil {
			highestPriorityVideoTrack = trackPriority
			continue
		}

		if trackPriority.Priority < highestPriorityVideoTrack.Priority {
			highestPriorityVideoTrack = trackPriority
		}
	}
	whipSession.TracksLock.RUnlock()

	if highestPriorityVideoTrack == nil {
		return ""
	}

	return highestPriorityVideoTrack.Rid
}
