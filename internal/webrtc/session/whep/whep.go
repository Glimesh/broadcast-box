package whep

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

// Create and start a new WHEP session
func CreateNewWhep(whepSessionId string, audioTrack *codecs.TrackMultiCodec, audioLayer string, videoTrack *codecs.TrackMultiCodec, videoLayer string, peerConnection *webrtc.PeerConnection) (whepSession *WhepSession) {
	log.Println("WhepSession.CreateNewWhep", whepSessionId)
	audioChannelSizeStr := os.Getenv(environment.WHEP_SESSION_AUDIOCHANNEL_SIZE)
	videoChannelSizeStr := os.Getenv(environment.WHEP_SESSION_VIDEOCHANNEL_SIZE)

	audioChannelSize, audioOk := strconv.Atoi(audioChannelSizeStr)
	videoChannelSize, videoOk := strconv.Atoi(videoChannelSizeStr)

	if audioOk != nil || videoOk != nil {
		log.Println("WhepSession.CreateNewWhep.AudioVideoChannelSize: Audio/Video channel sizes must be a valid number")
		audioChannelSize = 100
		videoChannelSize = 100
	}

	activeContext, activeContextCancel := context.WithCancel(context.Background())
	whepSession = &WhepSession{
		SessionId:           whepSessionId,
		AudioTrack:          audioTrack,
		VideoTrack:          videoTrack,
		AudioTimestamp:      5000,
		VideoTimestamp:      5000,
		AudioChannel:        make(chan codecs.TrackPacket, audioChannelSize),
		VideoChannel:        make(chan codecs.TrackPacket, videoChannelSize),
		WhipEventsChannel:   make(chan any, 100),
		SseEventsChannel:    make(chan any, 100),
		PeerConnection:      peerConnection,
		ActiveContext:       activeContext,
		ActiveContextCancel: activeContextCancel,
	}

	log.Println("WhepSession.CreateNewWhep.AudioLayer", audioLayer)
	log.Println("WhepSession.CreateNewWhep.VideoLayer", videoLayer)
	whepSession.AudioLayerCurrent.Store(audioLayer)
	whepSession.VideoLayerCurrent.Store(videoLayer)
	whepSession.IsWaitingForKeyframe.Store(true)
	whepSession.IsSessionClosed.Store(false)

	// Start WHEP go routines
	go whepSession.handleEvents()

	go func() {
		<-whepSession.ActiveContext.Done()
		whepSession.Close()
	}()

	return whepSession
}

// Closes down the WHEP session completely
func (whepSession *WhepSession) Close() {
	// Close WHEP channels
	whepSession.SessionClose.Do(func() {
		log.Println("WhepSession.Close")
		whepSession.IsSessionClosed.Store(true)

		// Close PeerConnection
		err := whepSession.PeerConnection.Close()
		if err != nil {
			log.Println("WhepSession.Close.PeerConnection.Error", err)
		}

		// Notify WHIP about closure
		whepSession.ActiveContextCancel()

		// Empty tracks
		whepSession.AudioLock.Lock()
		whepSession.VideoLock.Lock()

		whepSession.AudioTrack = nil
		whepSession.VideoTrack = nil

		whepSession.VideoLock.Unlock()
		whepSession.AudioLock.Unlock()

	})
}

// Get the current status of the WHEP session
func (whepSession *WhepSession) GetWhepSessionStatus() (state WhepSessionState) {
	whepSession.AudioLock.RLock()
	whepSession.VideoLock.RLock()

	currentAudioLayer := whepSession.AudioLayerCurrent.Load().(string)
	currentVideoLayer := whepSession.VideoLayerCurrent.Load().(string)

	state = WhepSessionState{
		Id: whepSession.SessionId,

		AudioLayerCurrent:   currentAudioLayer,
		AudioTimestamp:      whepSession.AudioTimestamp,
		AudioPacketsWritten: whepSession.AudioPacketsWritten,
		AudioSequenceNumber: uint64(whepSession.AudioSequenceNumber),

		VideoLayerCurrent:   currentVideoLayer,
		VideoTimestamp:      whepSession.VideoTimestamp,
		VideoPacketsWritten: whepSession.VideoPacketsWritten,
		VideoSequenceNumber: uint64(whepSession.VideoSequenceNumber),
	}

	whepSession.VideoLock.RUnlock()
	whepSession.AudioLock.RUnlock()

	return
}

// Finds the corresponding Whip session to the Whep session id and sets the requested audio layer
func (whepSession *WhepSession) SetAudioLayer(encodingId string) {
	log.Println("Setting Audio Layer")
	whepSession.AudioLayerCurrent.Store(encodingId)
	whepSession.IsWaitingForKeyframe.Store(true)
}

// Finds the corresponding Whip session to the Whep session id and sets the requested video layer
func (whepSession *WhepSession) SetVideoLayer(encodingId string) {
	log.Println("Setting Video Layer")
	whepSession.VideoLayerCurrent.Store(encodingId)
	whepSession.IsWaitingForKeyframe.Store(true)
}
