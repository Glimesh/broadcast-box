package main

import (
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/console"
	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/networktest"
	"github.com/glimesh/broadcast-box/internal/server"
	"github.com/glimesh/broadcast-box/internal/webrtc"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	environment.SetupLogger()
	environment.LoadEnvironmentVariables()
	console.HandleConsoleFlags()

	if shouldProfileApplication := os.Getenv(environment.ENABLE_PROFILING); strings.EqualFold(shouldProfileApplication, "true") {
		go func() {
			runtime.SetBlockProfileRate(1)
			runtime.SetMutexProfileFraction(1)
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	log.Println("Booting up Broadcast", time.Now().Format("2006-01-02 15:04:05"))
	webrtc.Setup()

	if shouldNetworkTest := os.Getenv(environment.NETWORK_TEST_ON_START); strings.EqualFold(shouldNetworkTest, "true") {
		networktest.RunNetworkTest()
	}

	server.StartWebServer()
}
