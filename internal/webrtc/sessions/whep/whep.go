package whep

import (
	"log"
	"time"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/pion/webrtc/v4"
)

// Create and start a new WHEP session
func CreateNewWHEP(
	whepSessionID string,
	streamKey string,
	audioTrack *codecs.TrackMultiCodec,
	videoTrack *codecs.TrackMultiCodec,
	peerConnection *webrtc.PeerConnection,
	pliSender func(),
) (w *WHEPSession) {
	log.Println("WHEPSession.CreateNewWHEP", whepSessionID)

	w = &WHEPSession{
		SessionID:               whepSessionID,
		StreamKey:               streamKey,
		AudioTrack:              audioTrack,
		VideoTrack:              videoTrack,
		AudioTimestamp:          5000,
		VideoTimestamp:          5000,
		PeerConnection:          peerConnection,
		pliSender:               pliSender,
		videoBitrateWindowStart: time.Now(),
	}

	w.AudioLayerCurrent.Store("")
	w.VideoLayerCurrent.Store("")
	w.IsWaitingForKeyframe.Store(true)
	w.IsSessionClosed.Store(false)
	return w
}

// Closes down the WHEP session completely
func (w *WHEPSession) Close() {
	// Close WHEP channels
	w.SessionClose.Do(func() {
		log.Println("WHEPSession.Close")
		w.IsSessionClosed.Store(true)

		// Close PeerConnection
		log.Println("WHEPSession.Close.PeerConnection.GracefulClose")
		err := w.PeerConnection.Close()
		if err != nil {
			log.Println("WHEPSession.Close.PeerConnection.Error", err)
		}
		log.Println("WHEPSession.Close.PeerConnection.GracefulClose.Completed")

		// Empty tracks
		w.AudioLock.Lock()
		w.VideoLock.Lock()

		w.AudioTrack = nil
		w.VideoTrack = nil

		w.VideoLock.Unlock()
		w.AudioLock.Unlock()

		if w.onClose != nil {
			w.onClose(w.SessionID)
		}
	})
}

func (w *WHEPSession) SetOnClose(onClose func(string)) {
	w.onClose = onClose
}

// Get the current status of the WHEP session
func (w *WHEPSession) GetWHEPSessionStatus() (state SessionState) {
	w.AudioLock.RLock()
	w.VideoLock.Lock()
	w.updateVideoBitrateLocked(time.Now())

	currentAudioLayer := w.AudioLayerCurrent.Load().(string)
	currentVideoLayer := w.VideoLayerCurrent.Load().(string)

	state = SessionState{
		ID: w.SessionID,

		AudioLayerCurrent:   currentAudioLayer,
		AudioTimestamp:      w.AudioTimestamp,
		AudioPacketsWritten: w.AudioPacketsWritten,
		AudioSequenceNumber: uint64(w.AudioSequenceNumber),

		VideoLayerCurrent:   currentVideoLayer,
		VideoTimestamp:      w.VideoTimestamp,
		VideoBitrate:        w.VideoBitrate.Load(),
		VideoPacketsWritten: w.VideoPacketsWritten,
		VideoPacketsDropped: w.VideoPacketsDropped.Load(),
		VideoSequenceNumber: uint64(w.VideoSequenceNumber),
	}

	w.VideoLock.Unlock()
	w.AudioLock.RUnlock()

	return
}

// Sets the requested audio layer for this WHEP session.
func (w *WHEPSession) SetAudioLayer(encodingID string) {
	log.Println("Setting Audio Layer")
	w.AudioLayerCurrent.Store(encodingID)
	w.IsWaitingForKeyframe.Store(true)
	w.SendPLI()
}

// Sets the requested video layer for this WHEP session.
func (w *WHEPSession) SetVideoLayer(encodingID string) {
	log.Println("Setting Video Layer")
	w.VideoLayerCurrent.Store(encodingID)
	w.IsWaitingForKeyframe.Store(true)
	w.SendPLI()
}

func (w *WHEPSession) SendPLI() {
	if w.IsSessionClosed.Load() {
		return
	}

	w.pliSender()
}

func (w *WHEPSession) updateVideoBitrateLocked(now time.Time) {
	if w.videoBitrateWindowStart.IsZero() {
		w.videoBitrateWindowStart = now
		return
	}

	elapsed := now.Sub(w.videoBitrateWindowStart)
	if elapsed < time.Second {
		return
	}

	bytesDiff := w.VideoBytesWritten - w.videoBitrateWindowBytes
	if bytesDiff < 0 {
		bytesDiff = 0
	}

	w.VideoBitrate.Store(uint64(float64(bytesDiff) / elapsed.Seconds()))
	w.videoBitrateWindowStart = now
	w.videoBitrateWindowBytes = w.VideoBytesWritten
}

func (w *WHEPSession) GetAudioLayerOrDefault(defaultLayer string) string {
	w.AudioLock.Lock()
	defer w.AudioLock.Unlock()

	currentLayer, _ := w.AudioLayerCurrent.Load().(string)
	if currentLayer != "" {
		return currentLayer
	}

	w.AudioLayerCurrent.Store(defaultLayer)
	return defaultLayer
}

func (w *WHEPSession) GetVideoLayerOrDefault(defaultLayer string) string {
	w.VideoLock.Lock()
	defer w.VideoLock.Unlock()

	currentLayer, _ := w.VideoLayerCurrent.Load().(string)
	if currentLayer != "" {
		return currentLayer
	}

	w.VideoLayerCurrent.Store(defaultLayer)
	w.IsWaitingForKeyframe.Store(true)
	return defaultLayer
}
