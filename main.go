package main

//nolint:all
import (
	"errors"
	"fmt"
	"io"

	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/pion/webrtc/v3"
)

type (
	stream struct {
		audioTrack, videoTrack *webrtc.TrackLocalStaticRTP
	}
)

//nolint:all
var (
	streamMap     map[string]stream
	streamMapLock sync.Mutex
)

func logHTTPError(w http.ResponseWriter, err string, code int) {
	log.Println(err)
	http.Error(w, err, code)
}

func getTracksForStream(streamName string) (
	*webrtc.TrackLocalStaticRTP,
	*webrtc.TrackLocalStaticRTP,
	error,
) {
	streamMapLock.Lock()
	defer streamMapLock.Unlock()

	foundStream, ok := streamMap[streamName]
	if !ok {
		//nolint:all
		videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
		if err != nil {
			//nolint:all
			return nil, nil, err
		}

		//nolint:all
		audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
		if err != nil {
			//nolint:all
			return nil, nil, err
		}

		foundStream = stream{
			audioTrack: audioTrack,
			videoTrack: videoTrack,
		}
		streamMap[streamName] = foundStream
	}

	return foundStream.audioTrack, foundStream.videoTrack, nil
}

//nolint:all
func whipHandler(w http.ResponseWriter, r *http.Request) {
	streamKey := r.Header.Get("Authorization")
	if streamKey == "" {
		logHTTPError(w, "Authorization was not set", http.StatusBadRequest)
		return
	}

	offer, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	}

	//nolint:all
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		logHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	}

	audioTrack, videoTrack, err := getTracksForStream(streamKey)
	if err != nil {
		logHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	}

	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		var localTrack *webrtc.TrackLocalStaticRTP
		if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
			localTrack = audioTrack
		} else {
			localTrack = videoTrack
		}

		//nolint:all
		rtpBuf := make([]byte, 1500)
		for {
			rtpRead, _, readErr := remoteTrack.Read(rtpBuf)
			switch {
			case errors.Is(readErr, io.EOF):
				return
			case readErr != nil:
				log.Println(readErr)
				return
			}

			if _, writeErr := localTrack.Write(rtpBuf[:rtpRead]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
				log.Println(writeErr)
				return
			}
		}
	})

	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
		if i == webrtc.ICEConnectionStateFailed {
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
				return
			}
		}
	})

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		logHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		logHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		logHTTPError(w, err.Error(), http.StatusBadRequest)
		return
	}

	<-gatherComplete

	fmt.Fprint(w, peerConnection.LocalDescription().SDP)
}

//nolint:all
func whepHandler(res http.ResponseWriter, req *http.Request) {
	streamKey := req.Header.Get("Authorization")
	if streamKey == "" {
		logHTTPError(res, "Authorization was not set", http.StatusBadRequest)
		return
	}

	offer, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	//nolint:all
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	audioTrack, videoTrack, err := getTracksForStream(streamKey)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = peerConnection.AddTrack(audioTrack); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
		SDP:  string(offer),
		Type: webrtc.SDPTypeOffer,
	}); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)

	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	<-gatherComplete

	fmt.Fprint(res, peerConnection.LocalDescription().SDP)
}

func main() {
	corsHandler := func(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
		return func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Access-Control-Allow-Origin", "*")
			res.Header().Set("Access-Control-Allow-Methods", "*")
			res.Header().Set("Access-Control-Allow-Headers", "*")

			if req.Method != http.MethodOptions {
				next(res, req)
			}
		}
	}

	streamMap = map[string]stream{}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./web/build")))
	mux.HandleFunc("/api/whip", corsHandler(whipHandler))
	mux.HandleFunc("/api/whep", corsHandler(whepHandler))

	//nolint:all
	log.Fatal((&http.Server{
		Handler: mux,
		Addr:    ":8080",
	}).ListenAndServe())
}
