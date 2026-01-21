package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	chatConnectRequestJSON struct {
		StreamKey string `json:"streamKey"`
	}
	chatConnectResponseJSON struct {
		ChatSessionId string `json:"chatSessionId"`
	}
	chatSendRequestJSON struct {
		Text        string `json:"text"`
		DisplayName string `json:"displayName"`
	}
)

func chatConnectHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		logHTTPError(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var r chatConnectRequestJSON
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	sessionID := ChatManager.Connect(r.StreamKey)

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(chatConnectResponseJSON{ChatSessionId: sessionID}); err != nil {
		log.Println(err)
	}
}

func chatSSEHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")

	vals := strings.Split(req.URL.RequestURI(), "/")
	sessionID := vals[len(vals)-1]

	lastEventIDStr := req.Header.Get("Last-Event-ID")
	var lastEventID uint64
	if lastEventIDStr != "" {
		lastEventID, _ = strconv.ParseUint(lastEventIDStr, 10, 64)
	}

	ch, cleanup, history, err := ChatManager.Subscribe(sessionID, lastEventID)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}
	defer cleanup()

	flusher, ok := res.(http.Flusher)
	if !ok {
		logHTTPError(res, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send history
	if len(history) > 0 {
		data, _ := json.Marshal(history)
		if _, err := fmt.Fprintf(res, "event: history\ndata: %s\n\n", string(data)); err != nil {
			log.Println(err)
			return
		}
		flusher.Flush()
	}

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event := <-ch:
			data, _ := json.Marshal(event.Message)
			if _, err := fmt.Fprintf(res, "id: %d\nevent: message\ndata: %s\n\n", event.ID, string(data)); err != nil {
				log.Println(err)
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := fmt.Fprintf(res, ": ping\n\n"); err != nil {
				log.Println(err)
				return
			}
			flusher.Flush()
		case <-req.Context().Done():
			return
		}
	}
}

func chatSendHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		logHTTPError(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var r chatSendRequestJSON
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if len(r.Text) < 1 || len(r.Text) > 2000 {
		logHTTPError(res, "Invalid message length", http.StatusBadRequest)
		return
	}
	if len(r.DisplayName) < 1 || len(r.DisplayName) > 80 {
		logHTTPError(res, "Invalid display name length", http.StatusBadRequest)
		return
	}

	vals := strings.Split(req.URL.RequestURI(), "/")
	sessionID := vals[len(vals)-1]

	if err := ChatManager.Send(sessionID, r.Text, r.DisplayName); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

func logHTTPError(w http.ResponseWriter, err string, code int) {
	log.Println(err)
	http.Error(w, err, code)
}
