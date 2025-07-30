package networktest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/pion/ice/v4"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v4"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
)

func Run(whepHandler func(res http.ResponseWriter, req *http.Request)) error {
	m := &webrtc.MediaEngine{}
	codecs.RegisterCodecs(m)

	s := webrtc.SettingEngine{}
	s.SetNetworkTypes([]webrtc.NetworkType{
		webrtc.NetworkTypeUDP4,
		webrtc.NetworkTypeUDP6,
		webrtc.NetworkTypeTCP4,
		webrtc.NetworkTypeTCP6,
	})

	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithSettingEngine(s)).NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		return err
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return err
	}

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return err
	}

	iceConnected, iceConnectedCancel := context.WithCancel(context.TODO())
	iceFailed, iceFailedCancel := context.WithCancel(context.TODO())

	peerConnection.OnICEConnectionStateChange(func(s webrtc.ICEConnectionState) {
		switch s {
		case webrtc.ICEConnectionStateFailed:
			iceFailedCancel()
		case webrtc.ICEConnectionStateConnected:
			iceConnectedCancel()
		}
	})

	req := httptest.NewRequest("POST", "/api/whip", strings.NewReader(offer.SDP))
	req.Header["Authorization"] = []string{"Bearer networktest"}
	recorder := httptest.NewRecorder()

	whepHandler(recorder, req)
	res := recorder.Result()

	if res.StatusCode != 201 {
		return fmt.Errorf("unexpected HTTP StatusCode %d", res.StatusCode)
	}

	if contentType := res.Header.Get("Content-Type"); contentType != "application/sdp" {
		return fmt.Errorf("unexpected HTTP Content-Type %s", contentType)
	}

	respBody, _ := io.ReadAll(res.Body)

	answerParsed := sdp.SessionDescription{}
	if err = answerParsed.Unmarshal(respBody); err != nil {
		return err
	}

	firstMediaSection := answerParsed.MediaDescriptions[0]
	filteredAttributes := []sdp.Attribute{}
	for i := range firstMediaSection.Attributes {
		a := firstMediaSection.Attributes[i]

		if a.Key == "candidate" {
			c, err := ice.UnmarshalCandidate(a.Value)
			if err != nil {
				return err
			}

			ip := net.ParseIP(c.Address())
			if ip == nil {
				return fmt.Errorf("candidate with invalid IP %s", c.Address())
			}

			if !ip.IsPrivate() {
				filteredAttributes = append(filteredAttributes, a)
			}
		} else {
			filteredAttributes = append(filteredAttributes, a)
		}
	}
	firstMediaSection.Attributes = filteredAttributes

	answer, err := answerParsed.Marshal()
	if err != nil {
		return err
	}

	if err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  string(answer),
	}); err != nil {
		return err
	}

	select {
	case <-iceConnected.Done():
		_ = peerConnection.Close()
		return nil
	case <-iceFailed.Done():
		_ = peerConnection.Close()

		return errors.New("network Test client failed to connect to Broadcast Box")
	case <-time.After(time.Second * 30):
		_ = peerConnection.Close()

		return errors.New("network Test client reported nothing in 30 seconds")
	}
}
