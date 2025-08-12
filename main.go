package main

import (
	"log"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/networktest"
	"github.com/glimesh/broadcast-box/internal/server"
	"github.com/glimesh/broadcast-box/internal/webrtc"
)

func main() {
	environment.LoadEnvironmentVariables()
	environment.HandleFlags()

	log.Println("Booting up Broadcast")
	webrtc.Setup()

	if shouldNetworkTest := os.Getenv("NETWORK_TEST_ON_START"); strings.EqualFold(shouldNetworkTest, "true") {
		networktest.RunNetworkTest()
	}

	server.StartWebServer()
}
