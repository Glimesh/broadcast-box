package webrtc

import (
	"errors"
	"strings"

	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/interceptors"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
	"github.com/pion/ice/v4"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

func Setup() {
	manager.SessionsManager = &manager.SessionManager{}
	manager.SessionsManager.Setup()

	// Initialize media engine
	mediaEngine := &webrtc.MediaEngine{}
	codecs.RegisterCodecs(mediaEngine)

	interceptorRegistry := interceptors.GetRegistry(mediaEngine)
	udpMuxCache := map[int]*ice.MultiUDPMuxDefault{}
	tcpMuxCache := map[string]ice.TCPMux{}

	initializeApiWhip(mediaEngine, udpMuxCache, tcpMuxCache, &interceptorRegistry)
	initializeApiWhep(mediaEngine, udpMuxCache, tcpMuxCache, &interceptorRegistry)
}

func initializeApiWhip(mediaEngine *webrtc.MediaEngine, udpMuxCache map[int]*ice.MultiUDPMuxDefault, tcpMuxCache map[string]ice.TCPMux, registry *interceptor.Registry) {
	manager.ApiWhip = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(registry),
		webrtc.WithSettingEngine(GetSettingEngine(true, tcpMuxCache, udpMuxCache)),
	)
}

func initializeApiWhep(mediaEngine *webrtc.MediaEngine, udpMuxCache map[int]*ice.MultiUDPMuxDefault, tcpMuxCache map[string]ice.TCPMux, registry *interceptor.Registry) {
	manager.ApiWhep = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(registry),
		webrtc.WithSettingEngine(GetSettingEngine(false, tcpMuxCache, udpMuxCache)),
	)
}

func HandleWhepPatch(sessionId, body string) error {
	session, isFound := manager.SessionsManager.GetWhepSessionById(sessionId)

	if !isFound {
		return errors.New("no session found")
	}

	session.PeerConnectionLock.Lock()
	if err := patchPeerConnection(session.PeerConnection, body); err != nil {
		session.PeerConnectionLock.Unlock()
		return err
	}
	session.PeerConnectionLock.Unlock()

	return nil
}

func HandleWhipPatch(sessionId, body string) error {
	session, isFound := manager.SessionsManager.GetSessionById(sessionId)

	if !isFound {
		return errors.New("no session found")
	}

	host := session.Host.Load()
	if host == nil {
		return errors.New("no host found")
	}

	host.PeerConnectionLock.Lock()
	if err := patchPeerConnection(host.PeerConnection, body); err != nil {
		host.PeerConnectionLock.Unlock()
		return err
	}
	host.PeerConnectionLock.Unlock()

	return nil
}

func HandleWhipDelete(sessionId string) error {
	session, isFound := manager.SessionsManager.GetSessionByHostSessionId(sessionId)

	if !isFound {
		return errors.New("no session found")
	}

	session.Close()
	return nil
}

func patchPeerConnection(peerConnection *webrtc.PeerConnection, body string) error {
	oldUfrag := getSdpKeyValue(peerConnection.CurrentRemoteDescription().SDP, "ice-ufrag")
	oldPwd := getSdpKeyValue(peerConnection.CurrentRemoteDescription().SDP, "ice-pwd")
	newUfrag, newPwd := getSdpKeyValue(body, "ice-ufrag"), getSdpKeyValue(body, "ice-pwd")

	isICERestart := oldUfrag != newUfrag || oldPwd != newPwd

	if isICERestart {
		return errors.New("ice restart not supported")
	}

	for line := range strings.SplitSeq(body, "\n") {
		expectedPrefix := "a=candidate:"

		if strings.HasPrefix(line, expectedPrefix) {
			if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{
				Candidate: strings.TrimSpace(strings.TrimPrefix(line, "a=")),
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// Retrieve value by SDP key from SDP body
func getSdpKeyValue(sdp string, key string) string {
	for l := range strings.SplitSeq(sdp, "\n") {
		expectedPrefix := "a=" + key + ":"
		if after, ok := strings.CutPrefix(l, expectedPrefix); ok {
			return after
		}
	}

	return ""
}
