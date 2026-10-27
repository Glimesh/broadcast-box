package console

import (
	"flag"
	"log"
	"os"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
)

func HandleConsoleFlags() {
	createNewProfile := flag.Bool(createNewProfile, false, "Create a new stream profile from the -streamKey flag")
	streamKey := flag.String(createNewProfileStreamKey, "", "The stream key used to identify a streaming session")

	flag.Parse()

	if *createNewProfile {
		if len(*streamKey) == 0 {
			log.Println("No stream key was provided. Use the flags `-createNewProfile -streamKey MyStreamKey` to create a new profile.")
			os.Exit(0)
		}

		token, err := authorization.CreateProfile(*streamKey)
		if err != nil {
			log.Println(err)
			os.Exit(0)
		}

		log.Println("Created", *streamKey, "with bearer token:", token)
		os.Exit(0)
	}
}
