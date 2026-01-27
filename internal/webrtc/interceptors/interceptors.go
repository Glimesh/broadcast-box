package interceptors

import (
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
	"log"
)

func GetRegistry(mediaEngine *webrtc.MediaEngine) interceptor.Registry {
	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		log.Fatal(err)
	}

	return *interceptorRegistry
}
