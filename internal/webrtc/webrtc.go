package webrtc

import (
	"github.com/glimesh/broadcast-box/internal/webrtc/codecs"
	"github.com/glimesh/broadcast-box/internal/webrtc/interceptors"
	"github.com/glimesh/broadcast-box/internal/webrtc/session"

	"github.com/pion/ice/v4"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

func Setup() {
	session.WhipSessions = map[string]*session.WhipSession{}

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
	session.ApiWhip = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(registry),
		webrtc.WithSettingEngine(GetSettingEngine(true, tcpMuxCache, udpMuxCache)),
	)
}

func initializeApiWhep(mediaEngine *webrtc.MediaEngine, udpMuxCache map[int]*ice.MultiUDPMuxDefault, tcpMuxCache map[string]ice.TCPMux, registry *interceptor.Registry) {
	session.ApiWhep = webrtc.NewAPI(
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(registry),
		webrtc.WithSettingEngine(GetSettingEngine(false, tcpMuxCache, udpMuxCache)),
	)
}
