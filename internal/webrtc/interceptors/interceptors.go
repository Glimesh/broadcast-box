package interceptors

import (
	"log/slog"
	"os"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

func GetRegistry(mediaEngine *webrtc.MediaEngine) interceptor.Registry {
	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		slog.Error("Failed to register default interceptors", "err", err)
		os.Exit(1)
	}

	return *interceptorRegistry
}
