package webrtc

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type (
	whepSession struct {
		streamKey string
		rtpSender *webrtc.RTPSender
	}

	simulcastLayerResponse struct {
		EncodingId string `json:"encodingId"`
	}
)

var (
	whepSessionsLock sync.RWMutex
	whepSessions     map[string]whepSession
)

func WHEPLayers(whepSessionId string) ([]byte, error) {
	whepSessionsLock.Lock()
	defer whepSessionsLock.Unlock()
	whepSession, ok := whepSessions[whepSessionId]
	if !ok {
		return nil, fmt.Errorf("No WHEP Session found with ID:%s", whepSessionId)
	}

	streamMapLock.Lock()
	defer streamMapLock.Unlock()
	whipSession, ok := streamMap[whepSession.streamKey]
	if !ok {
		return nil, fmt.Errorf("No WHIP Session found with streamKey:%s", whepSessionId)
	}

	layers := []simulcastLayerResponse{}
	for i := range whipSession.videoTracks {
		id := whipSession.videoTracks[i].ID()
		if id == videoTrackLabelDefault {
			id = whipSession.defaultVideoTrackLabel
		}

		layers = append(layers, simulcastLayerResponse{EncodingId: id})
	}

	resp := map[string]map[string][]simulcastLayerResponse{
		"1": map[string][]simulcastLayerResponse{
			"layers": layers,
		},
	}

	return json.Marshal(resp)
}

func WHEPChangeLayer(whepSessionId, layer string) error {
	whepSessionsLock.Lock()
	defer whepSessionsLock.Unlock()
	whepSession, ok := whepSessions[whepSessionId]
	if !ok {
		return fmt.Errorf("No WHEP Session found with ID:%s", whepSessionId)
	}

	streamMapLock.Lock()
	defer streamMapLock.Unlock()
	whipSession, ok := streamMap[whepSession.streamKey]
	if !ok {
		return fmt.Errorf("No WHIP Session found with streamKey:%s", whepSessionId)
	}

	if layer == whipSession.defaultVideoTrackLabel {
		layer = videoTrackLabelDefault
	}

	var newTrack *webrtc.TrackLocalStaticRTP
	for i := range whipSession.videoTracks {
		if whipSession.videoTracks[i].ID() == layer {
			newTrack = whipSession.videoTracks[i]
		}
	}

	return whepSession.rtpSender.ReplaceTrack(newTrack)
}

func WHEP(offer, streamKey string) (string, string, error) {
	whepSessionId := uuid.New().String()

	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", "", err
	}

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
			}

			whepSessionsLock.Lock()
			delete(whepSessions, whepSessionId)
			whepSessionsLock.Unlock()
		}
	})

	audioTrack, videoTrack, pliChan, err := getTracksForStream(streamKey)
	if err != nil {
		return "", "", err
	}

	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
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
					case pliChan <- true:
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

	whepSessionsLock.Lock()
	whepSessions[whepSessionId] = whepSession{
		streamKey: streamKey,
		rtpSender: rtpSender,
	}
	whepSessionsLock.Unlock()

	return peerConnection.LocalDescription().SDP, whepSessionId, nil
}
