package main

import (
	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/server"
	"github.com/glimesh/broadcast-box/internal/webrtc"
	"log"
)

func main() {
	environment.LoadEnvironmentVariables()
	environment.HandleFlags()

	log.Println("Booting up Broadcast")
	webrtc.Setup()

	server.StartWebServer()
}
