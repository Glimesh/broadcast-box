package webrtc

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type (
	whepSession struct {
		videoTrack         *trackMultiCodec
		currentLayer       atomic.Value
		waitingForKeyframe atomic.Bool
		sequenceNumber     uint16
		timestamp          uint32
		packetsWritten     uint64
	}

	simulcastLayerResponse struct {
		EncodingId string `json:"encodingId"`
	}
)

func WHEPLayers(whepSessionId string) ([]byte, error) {
	streamMapLock.Lock()
	defer streamMapLock.Unlock()

	layers := []simulcastLayerResponse{}
	for streamKey := range streamMap {
		streamMap[streamKey].whepSessionsLock.Lock()
		defer streamMap[streamKey].whepSessionsLock.Unlock()

		if _, ok := streamMap[streamKey].whepSessions[whepSessionId]; ok {
			for i := range streamMap[streamKey].videoTracks {
				layers = append(layers, simulcastLayerResponse{EncodingId: streamMap[streamKey].videoTracks[i].rid})
			}

			break
		}
	}

	resp := map[string]map[string][]simulcastLayerResponse{
		"1": map[string][]simulcastLayerResponse{
			"layers": layers,
		},
	}

	return json.Marshal(resp)
}

func WHEPChangeLayer(whepSessionId, layer string) error {
	streamMapLock.Lock()
	defer streamMapLock.Unlock()

	for streamKey := range streamMap {
		streamMap[streamKey].whepSessionsLock.Lock()
		defer streamMap[streamKey].whepSessionsLock.Unlock()

		if _, ok := streamMap[streamKey].whepSessions[whepSessionId]; ok {
			streamMap[streamKey].whepSessions[whepSessionId].currentLayer.Store(layer)
			streamMap[streamKey].whepSessions[whepSessionId].waitingForKeyframe.Store(true)
			streamMap[streamKey].pliChan <- true
		}
	}

	return nil
}

func WHEP(offer, streamKey string) (string, string, error) {
	maybePrintOfferAnswer(offer, true)

	streamMapLock.Lock()
	defer streamMapLock.Unlock()
	stream, err := getStream(streamKey, "")
	if err != nil {
		return "", "", err
	}

	whepSessionId := uuid.New().String()

	videoTrack := &trackMultiCodec{id: "video", streamID: "pion"}

	peerConnection, err := newPeerConnection(apiWhep)
	if err != nil {
		return "", "", err
	}

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed || i == webrtc.ICEConnectionStateClosed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
			}

			peerConnectionDisconnected(false, streamKey, whepSessionId)
		}
	})

	if _, err = peerConnection.AddTrack(stream.audioTrack); err != nil {
		return "", "", err
	}

	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		return "", "", err
	}

	go func() {
		for {
			rtcpPackets, _, rtcpErr := rtpSender.ReadRTCP()
			if rtcpErr != nil {
				return
			}

			for _, r := range rtcpPackets {
				if _, isPLI := r.(*rtcp.PictureLossIndication); isPLI {
					select {
					case stream.pliChan <- true:
					default:
					}
				}
			}
		}
	}()

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  offer,
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		return "", "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		return "", "", err
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", "", err
	}

	<-gatherComplete

	stream.whepSessionsLock.Lock()
	defer stream.whepSessionsLock.Unlock()

	stream.whepSessions[whepSessionId] = &whepSession{
		videoTrack: videoTrack,
		timestamp:  50000,
	}
	stream.whepSessions[whepSessionId].currentLayer.Store("")
	stream.whepSessions[whepSessionId].waitingForKeyframe.Store(false)

	return maybePrintOfferAnswer(appendAnswer(peerConnection.LocalDescription().SDP), false), whepSessionId, nil
}

func (w *whepSession) sendVideoPacket(rtpPkt *rtp.Packet, layer string, timeDiff int64, sequenceDiff int, codec videoTrackCodec, isKeyframe bool) {
	if w.currentLayer.Load() == "" {
		w.currentLayer.Store(layer)
	} else if layer != w.currentLayer.Load() {
		return
	} else if w.waitingForKeyframe.Load() {
		if !isKeyframe {
			return
		}

		w.waitingForKeyframe.Store(false)
	}

	w.packetsWritten += 1
	w.sequenceNumber = uint16(int(w.sequenceNumber) + sequenceDiff)
	w.timestamp = uint32(int64(w.timestamp) + timeDiff)

	rtpPkt.SequenceNumber = w.sequenceNumber
	rtpPkt.Timestamp = w.timestamp

	if err := w.videoTrack.WriteRTP(rtpPkt, codec); err != nil && !errors.Is(err, io.ErrClosedPipe) {
		log.Println(err)
	}
}
