package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/glimesh/broadcast-box/internal/server/authorization"
	"github.com/glimesh/broadcast-box/internal/server/helpers"
)

type ICEComponentServer struct {
	Urls       string `json:"urls"`
	Username   string `json:"username"`
	Credential string `json:"credential"`
}

func clientICEHandler(responseWriter http.ResponseWriter, request *http.Request) {
	turnServers := os.Getenv("TURN_SERVERS")
	turnAuthKey := os.Getenv("TURN_SERVER_AUTH_SECRET")
	stunServers := os.Getenv("STUN_SERVERS")
	var servers []ICEComponentServer

	if turnServers == "" && stunServers == "" {
		_, err := responseWriter.Write([]byte("[]"))
		if err != nil {
			log.Println("error writing empty TURN/STUN response", err)
		}

		return
	}

	if turnServers != "" {
		turnServerNames := strings.Split(turnServers, "|")
		for server := range turnServerNames {
			log.Println("Adding TURN server", server)

			if turnAuthKey != "" {
				username, credentials := authorization.GetTURNCredentials()

				servers = append(servers, ICEComponentServer{
					Urls:       "turn:" + turnServerNames[server],
					Username:   username,
					Credential: credentials,
				})
			} else {
				servers = append(servers, ICEComponentServer{
					Urls: "turn:" + turnServerNames[server],
				})
			}
		}

	}

	if stunServers != "" {
		stunServerNames := strings.Split(stunServers, "|")
		for server := range stunServerNames {
			servers = append(servers, ICEComponentServer{
				Urls: "stun:" + stunServerNames[server],
			})
		}
	}

	if len(servers) == 0 {
		_, err := responseWriter.Write([]byte("[]"))
		if err != nil {
			log.Println("error writing empty TURN/STUN response", err)
		}

		return
	}

	if err := json.NewEncoder(responseWriter).Encode(servers); err != nil {
		helpers.LogHttpError(
			responseWriter,
			"Internal Server Error",
			http.StatusInternalServerError)
		log.Println(err.Error())
	}

	responseWriter.Header().Add("Content-Type", "application/json")
}
