package main

//nolint:all
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"

	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
)

const (
	envFileProd = ".env.production"
	envFileDev  = ".env.development"
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
	api           *webrtc.API
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
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
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
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
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

func indexHTMLWhenNotFound(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_, err := fs.Open(path.Clean(req.URL.Path)) // Do not allow path traversals.
		if errors.Is(err, os.ErrNotExist) {
			http.ServeFile(resp, req, "./web/build/index.html")

			return
		}
		fileServer.ServeHTTP(resp, req)
	})
}

func corsHandler(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Methods", "*")
		res.Header().Set("Access-Control-Allow-Headers", "*")

		if req.Method != http.MethodOptions {
			next(res, req)
		}
	}
}

func getPublicIP() string {
	//nolint:all
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		//nolint:all
		log.Fatal(err)
	}

	//nolint:all
	ip := struct {
		Query string
	}{}
	if err = json.Unmarshal(body, &ip); err != nil {
		log.Fatal(err)
	}

	if ip.Query == "" {
		log.Fatal("Query entry was not populated")
	}

	return ip.Query
}

//nolint:all
func populateSettingEngine(settingEngine *webrtc.SettingEngine) {
	NAT1To1IPs := []string{}

	if os.Getenv("INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP") != "" {
		NAT1To1IPs = append(NAT1To1IPs, getPublicIP())
	}

	if os.Getenv("NAT_1_TO_1_IP") != "" {
		NAT1To1IPs = append(NAT1To1IPs, os.Getenv("NAT_1_TO_1_IP"))
	}

	if len(NAT1To1IPs) != 0 {
		settingEngine.SetNAT1To1IPs(NAT1To1IPs, webrtc.ICECandidateTypeHost)
	}

	if os.Getenv("INTERFACE_FILTER") != "" {
		settingEngine.SetInterfaceFilter(func(i string) bool {
			return i == os.Getenv("INTERFACE_FILTER")
		})
	}

	if os.Getenv("UDP_MUX_PORT") != "" {
		udpPort, err := strconv.Atoi(os.Getenv("UDP_MUX_PORT"))
		if err != nil {
			log.Fatal(err)
		}

		udpMux, err := ice.NewMultiUDPMuxFromPort(udpPort)
		if err != nil {
			log.Fatal(err)
		}

		settingEngine.SetICEUDPMux(udpMux)
	}

	if os.Getenv("TCP_MUX_ADDRESS") != "" {
		tcpAddr, err := net.ResolveTCPAddr("udp", os.Getenv("TCP_MUX_ADDRESS"))
		if err != nil {
			log.Fatal(err)
		}

		tcpListener, err := net.ListenTCP("tcp", tcpAddr)
		if err != nil {
			log.Fatal(err)
		}

		//nolint:all
		settingEngine.SetICETCPMux(webrtc.NewICETCPMux(nil, tcpListener, 8))
	}
}

func main() {
	if os.Getenv("APP_ENV") == "production" {
		log.Println("Loading `" + envFileProd + "`")

		if err := godotenv.Load(envFileProd); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Loading `" + envFileDev + "`")

		if err := godotenv.Load(envFileDev); err != nil {
			log.Fatal(err)
		}
	}

	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		log.Fatal(err)
	}

	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		log.Fatal(err)
	}

	//nolint:all
	settingEngine := webrtc.SettingEngine{}
	populateSettingEngine(&settingEngine)

	api = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
		webrtc.WithSettingEngine(settingEngine),
	)

	streamMap = map[string]stream{}
	mux := http.NewServeMux()
	mux.Handle("/", indexHTMLWhenNotFound(http.Dir("./web/build")))
	mux.HandleFunc("/api/whip", corsHandler(whipHandler))
	mux.HandleFunc("/api/whep", corsHandler(whepHandler))

	log.Println("Running HTTP Server at `" + os.Getenv("HTTP_ADDRESS") + "`")

	//nolint:all
	log.Fatal((&http.Server{
		Handler: mux,
		Addr:    os.Getenv("HTTP_ADDRESS"),
	}).ListenAndServe())
}
