package ip

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func GetPublicIP() string {
	req, err := http.Get("http://ip-api.com/json/")

	if err != nil {
		slog.Error("Failed to get Public IP", "err", err)
		os.Exit(1)
	}

	defer func() {
		if closeErr := req.Body.Close(); closeErr != nil {
			slog.Error("Failed to get Public IP", "err", err)
			os.Exit(1)
		}
	}()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("Failed to get Public IP", "err", err)
		os.Exit(1)
	}

	ip := struct {
		Query string
	}{}

	if err = json.Unmarshal(body, &ip); err != nil {
		slog.Error("Failed to get Public IP", "err", err)
		os.Exit(1)
	}

	if ip.Query == "" {
		slog.Error("Failed to get Public IP", "err", "Query entry was not populated")
		os.Exit(1)
	}

	return ip.Query
}
