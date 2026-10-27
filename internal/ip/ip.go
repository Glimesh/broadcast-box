package ip

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func GetPublicIP() string {
	req, err := http.Get("http://ip-api.com/json/")

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if closeErr := req.Body.Close(); closeErr != nil {
			log.Fatal(err)
		}
	}()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	ip := struct {
		Query string
	}{}

	if err = json.Unmarshal(body, &ip); err != nil {
		log.Fatal(err)
	}

	if ip.Query == "" {
		log.Fatal("Query entry was not populated")
	}

	return ip.Query
}
