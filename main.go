package main

import (
	"log"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/console"
	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server"
	"github.com/glimesh/broadcast-box/internal/test"
	"github.com/glimesh/broadcast-box/internal/webrtc"
)

func main() {
	environment.SetupLogger()
	environment.LoadEnvironmentVariables()
	console.HandleConsoleFlags()

	log.Println("Booting up Broadcast", time.Now().Format("2006-01-02 15:04:05"))
	webrtc.Setup()

	if shouldNetworkTest := os.Getenv(environment.NETWORK_TEST_ON_START); strings.EqualFold(shouldNetworkTest, "true") {
		networktest.RunNetworkTest()
	}

	server.StartWebServer()
}
