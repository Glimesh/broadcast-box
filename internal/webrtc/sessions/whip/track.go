package whip

import (
	"log"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

// Add a new AudioTrack to the Whip session
func (whip *WhipSession) AddAudioTrack(rid string, streamKey string, codec codecs.TrackCodeType) (*AudioTrack, error) {
	log.Println("WhipSession.AddAudioTrack:", streamKey, "(", rid, ")")
	whip.TracksLock.Lock()
	defer whip.TracksLock.Unlock()

	if existingTrack, ok := whip.AudioTracks[rid]; ok {
		return existingTrack, nil
	}

	track := &AudioTrack{
		Rid:       rid,
		SessionId: whip.Id,
		Track: codecs.CreateTrackMultiCodec(
			"audio-"+uuid.New().String(),
			rid,
			streamKey,
			webrtc.RTPCodecTypeAudio,
			codec),
	}
	track.LastReceived.Store(time.Time{})

	whip.AudioTracks[track.Rid] = track

	return track, nil
}

// Add a new VideoTrack to the Whip session
func (whip *WhipSession) AddVideoTrack(rid string, streamKey string, codec codecs.TrackCodeType) (*VideoTrack, error) {
	log.Println("WhipSession.AddVideoTrack:", "(", rid, ")")
	whip.TracksLock.Lock()
	defer whip.TracksLock.Unlock()

	if existingTrack, ok := whip.VideoTracks[rid]; ok {
		return existingTrack, nil
	}

	track := &VideoTrack{
		Rid:       rid,
		SessionId: whip.Id,
		Track: codecs.CreateTrackMultiCodec(
			"video-"+uuid.New().String(),
			rid,
			streamKey,
			webrtc.RTPCodecTypeVideo,
			codec),
	}
	track.LastReceived.Store(time.Time{})

	whip.VideoTracks[rid] = track

	return track, nil
}

// Remove Audio and Video tracks coming from the whip session id
func (whip *WhipSession) RemoveTracks() {
	log.Println("WhipSession.RemoveTracks")

	whip.TracksLock.Lock()
	whip.AudioTracks = make(map[string]*AudioTrack)
	whip.VideoTracks = make(map[string]*VideoTrack)
	whip.TracksLock.Unlock()

	whip.OnTrackChangeChannel <- struct{}{}
}

// Get highest prioritized audio track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (whip *WhipSession) GetHighestPrioritizedAudioTrack() string {
	if len(whip.AudioTracks) == 0 {
		log.Println("No Audio tracks was found for", whip.Id)
		return ""
	}

	whip.TracksLock.RLock()
	var highestPriorityAudioTrack *AudioTrack
	for _, trackPriority := range whip.AudioTracks {
		if highestPriorityAudioTrack == nil {
			highestPriorityAudioTrack = trackPriority
			continue
		}

		if trackPriority.Priority < highestPriorityAudioTrack.Priority {
			highestPriorityAudioTrack = trackPriority
		}
	}
	whip.TracksLock.RUnlock()

	if highestPriorityAudioTrack == nil {
		return ""
	}

	return highestPriorityAudioTrack.Rid

}

// Get highest prioritized video track in the whip session
// This only works if the priority has been set.
// Currently this is only supported when being set through the simulcast
// property in the offer made by the whip connection
func (whip *WhipSession) GetHighestPrioritizedVideoTrack() string {
	if len(whip.VideoTracks) == 0 {
		log.Println("No Video tracks was found for", whip.Id)
	}

	var highestPriorityVideoTrack *VideoTrack

	whip.TracksLock.RLock()
	for _, trackPriority := range whip.VideoTracks {
		if highestPriorityVideoTrack == nil {
			highestPriorityVideoTrack = trackPriority
			continue
		}

		if trackPriority.Priority < highestPriorityVideoTrack.Priority {
			highestPriorityVideoTrack = trackPriority
		}
	}
	whip.TracksLock.RUnlock()

	if highestPriorityVideoTrack == nil {
		return ""
	}

	return highestPriorityVideoTrack.Rid
}
