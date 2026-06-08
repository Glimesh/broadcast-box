package webrtc

import (
	"errors"
	"strings"

	"github.com/glimesh/broadcast-box/internal/chat"
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/interceptors"
	"github.com/glimesh/broadcast-box/internal/webrtc/sessions/manager"
	"github.com/pion/ice/v4"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

func Setup(chatManager *chat.Manager) {
	manager.SessionsManager = &manager.SessionManager{
		ChatManager: chatManager,
	}
	manager.SessionsManager.Setup()

	// Initialize media engine
	mediaEngine := &webrtc.MediaEngine{}
	codecs.RegisterCodecs(mediaEngine)

	interceptorRegistry := interceptors.GetRegistry(mediaEngine)
	udpMuxCache := map[int]*ice.MultiUDPMuxDefault{}
	tcpMuxCache := map[string]ice.TCPMux{}

	initializeAPIWHIP(mediaEngine, udpMuxCache, tcpMuxCache, &interceptorRegistry)
	initializeAPIWHEP(mediaEngine, udpMuxCache, tcpMuxCache, &interceptorRegistry)
}

func initializeAPIWHIP(mediaEngine *webrtc.MediaEngine, udpMuxCache map[int]*ice.MultiUDPMuxDefault, tcpMuxCache map[string]ice.TCPMux, registry *interceptor.Registry) {
	manager.APIWHIP = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(registry),
		webrtc.WithSettingEngine(getSettingEngine(true, tcpMuxCache, udpMuxCache)),
	)
}

func initializeAPIWHEP(mediaEngine *webrtc.MediaEngine, udpMuxCache map[int]*ice.MultiUDPMuxDefault, tcpMuxCache map[string]ice.TCPMux, registry *interceptor.Registry) {
	manager.APIWHEP = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(registry),
		webrtc.WithSettingEngine(getSettingEngine(false, tcpMuxCache, udpMuxCache)),
	)
}

func HandleWHEPPatch(sessionID, body string) error {
	session, isFound := manager.SessionsManager.GetWHEPSessionByID(sessionID)

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

func HandleWHIPPatch(sessionID, body string) error {
	session, isFound := manager.SessionsManager.GetSessionByID(sessionID)

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

func HandleWHIPDelete(sessionID string) error {
	session, isFound := manager.SessionsManager.GetSessionByHostSessionID(sessionID)

	if !isFound {
		return errors.New("no session found")
	}

	session.RemoveHost()
	if session.GetStreamStatus().ViewerCount == 0 {
		session.Close()
	}

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
