package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/glimesh/broadcast-box/internal/chat"
	"github.com/glimesh/broadcast-box/internal/networktest"
	"github.com/glimesh/broadcast-box/internal/webhook"
	"github.com/glimesh/broadcast-box/internal/webrtc"
	"github.com/joho/godotenv"
)

const (
	envFileProd = ".env.production"
	envFileDev  = ".env.development"

	networkTestIntroMessage   = "\033[0;33mNETWORK_TEST_ON_START is enabled. If the test fails Broadcast Box will exit.\nSee the README for how to debug or disable NETWORK_TEST_ON_START\033[0m"
	networkTestSuccessMessage = "\033[0;32mNetwork Test passed.\nHave fun using Broadcast Box.\033[0m"
	networkTestFailedMessage  = "\033[0;31mNetwork Test failed.\n%s\nPlease see the README and join Discord for help\033[0m"
)

var (
	errNoBuildDirectoryErr = errors.New("\033[0;31mBuild directory does not exist, run `npm install` and `npm run build` in the web directory.\033[0m")
	errAuthorizationNotSet = errors.New("authorization was not set")
	errInvalidStreamKey    = errors.New("invalid stream key format")

	streamKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.~]+$`)
)

type (
	whepLayerRequestJSON struct {
		MediaId    string `json:"mediaId"`
		EncodingId string `json:"encodingId"`
	}

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

var (
	chatManager *chat.Manager
)

func getStreamKey(action string, r *http.Request) (streamKey string, err error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return "", errAuthorizationNotSet
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authorizationHeader, bearerPrefix) {
		return "", errInvalidStreamKey
	}

	streamKey = strings.TrimPrefix(authorizationHeader, bearerPrefix)
	if webhookUrl := os.Getenv("WEBHOOK_URL"); webhookUrl != "" {
		streamKey, err = webhook.CallWebhook(webhookUrl, action, streamKey, r)
		if err != nil {
			return "", err
		}
	}

	if !streamKeyRegex.MatchString(streamKey) {
		return "", errInvalidStreamKey
	}

	return streamKey, nil
}

func logHTTPError(w http.ResponseWriter, err string, code int) {
	log.Println(err)
	http.Error(w, err, code)
}

func whipHandler(res http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	streamKey, err := getStreamKey("whip-connect", r)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	offer, err := io.ReadAll(r.Body)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	answer, err := webrtc.WHIP(string(offer), streamKey)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Add("Location", "/api/whip")
	res.Header().Add("Content-Type", "application/sdp")
	res.WriteHeader(http.StatusCreated)
	if _, err = fmt.Fprint(res, answer); err != nil {
		log.Println(err)
	}
}

func whepHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		return
	}

	streamKey, err := getStreamKey("whep-connect", req)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	offer, err := io.ReadAll(req.Body)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	answer, whepSessionId, err := webrtc.WHEP(string(offer), streamKey)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	apiPath := req.Host + strings.TrimSuffix(req.URL.RequestURI(), "whep")
	res.Header().Add("Link", `<`+apiPath+"sse/"+whepSessionId+`>; rel="urn:ietf:params:whep:ext:core:server-sent-events"; events="layers"`)
	res.Header().Add("Link", `<`+apiPath+"layer/"+whepSessionId+`>; rel="urn:ietf:params:whep:ext:core:layer"`)
	res.Header().Add("Location", "/api/whep")
	res.Header().Add("Content-Type", "application/sdp")
	res.WriteHeader(http.StatusCreated)
	if _, err = fmt.Fprint(res, answer); err != nil {
		log.Println(err)
	}
}

func whepServerSentEventsHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")

	vals := strings.Split(req.URL.RequestURI(), "/")
	whepSessionId := vals[len(vals)-1]

	layers, err := webrtc.WHEPLayers(whepSessionId)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err = fmt.Fprintf(res, "event: layers\ndata: %s\n\n\n", string(layers)); err != nil {
		log.Println(err)
	}
}

func whepLayerHandler(res http.ResponseWriter, req *http.Request) {
	var r whepLayerRequestJSON
	if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	vals := strings.Split(req.URL.RequestURI(), "/")
	whepSessionId := vals[len(vals)-1]

	if err := webrtc.WHEPChangeLayer(whepSessionId, r.EncodingId); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func statusHandler(res http.ResponseWriter, req *http.Request) {
	if os.Getenv("DISABLE_STATUS") != "" {
		logHTTPError(res, "Status Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	res.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(res).Encode(webrtc.GetStreamStatuses()); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
	}
}

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
	if r.StreamKey == "" || !streamKeyRegex.MatchString(r.StreamKey) {
		logHTTPError(res, errInvalidStreamKey.Error(), http.StatusBadRequest)
		return
	}

	sessionID := chatManager.Connect(r.StreamKey)

	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(res).Encode(chatConnectResponseJSON{ChatSessionId: sessionID}); err != nil {
		log.Println(err)
	}
}

func chatSSEHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		logHTTPError(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	res.Header().Set("Content-Type", "text/event-stream")
	res.Header().Set("Cache-Control", "no-cache")
	res.Header().Set("Connection", "keep-alive")

	vals := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	sessionID := vals[len(vals)-1]

	lastEventIDStr := req.Header.Get("Last-Event-ID")
	var lastEventID uint64
	if lastEventIDStr != "" {
		var err error
		lastEventID, err = strconv.ParseUint(lastEventIDStr, 10, 64)
		if err != nil {
			logHTTPError(res, "Invalid Last-Event-ID", http.StatusBadRequest)
			return
		}
	}

	ch, cleanup, history, err := chatManager.Subscribe(sessionID, lastEventID)
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

	// Send history as normal message events for SSE resume support.
	for _, event := range history {
		data, err := json.Marshal(event.Message)
		if err != nil {
			log.Println(err)
			return
		}
		if _, err := fmt.Fprintf(res, "id: %d\nevent: message\ndata: %s\n\n", event.ID, string(data)); err != nil {
			log.Println(err)
			return
		}
	}
	flusher.Flush()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(event.Message)
			if err != nil {
				log.Println(err)
				return
			}
			if _, err := fmt.Fprintf(res, "id: %d\nevent: message\ndata: %s\n\n", event.ID, string(data)); err != nil {
				log.Println(err)
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if ok := chatManager.TouchSession(sessionID); !ok {
				return
			}
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

	vals := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	sessionID := vals[len(vals)-1]

	if err := chatManager.Send(sessionID, r.Text, r.DisplayName); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

func indexHTMLWhenNotFound(fs http.FileSystem) http.Handler {
	fileServer := http.FileServer(fs)

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_, err := fs.Open(path.Clean(req.URL.Path)) // Do not allow path traversals.
		if errors.Is(err, os.ErrNotExist) {
			http.ServeFile(resp, req, "./web/build/index.html")

			return
		}
		fileServer.ServeHTTP(resp, req)
	})
}

func corsHandler(next func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", "*")
		res.Header().Set("Access-Control-Allow-Methods", "*")
		res.Header().Set("Access-Control-Allow-Headers", "*")
		res.Header().Set("Access-Control-Expose-Headers", "*")

		if req.Method != http.MethodOptions {
			next(res, req)
		}
	}
}

func loadConfigs() error {
	if os.Getenv("APP_ENV") == "development" {
		log.Println("Loading `" + envFileDev + "`")
		return godotenv.Load(envFileDev)
	} else {
		log.Println("Loading `" + envFileProd + "`")
		if err := godotenv.Load(envFileProd); err != nil {
			return err
		}

		if _, err := os.Stat("./web/build"); os.IsNotExist(err) && os.Getenv("DISABLE_FRONTEND") == "" {
			return errNoBuildDirectoryErr
		}

		return nil
	}
}

func main() {
	if err := loadConfigs(); err != nil {
		log.Println("Failed to find config in CWD, changing CWD to executable path")

		exePath, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}

		if err = os.Chdir(filepath.Dir(exePath)); err != nil {
			log.Fatal(err)
		}

		if err = loadConfigs(); err != nil {
			log.Fatal(err)
		}
	}

	webrtc.Configure()
	chatManager = chat.NewManager()

	if os.Getenv("NETWORK_TEST_ON_START") == "true" {
		fmt.Println(networkTestIntroMessage) //nolint

		go func() {
			time.Sleep(time.Second * 5)

			if networkTestErr := networktest.Run(whepHandler); networkTestErr != nil {
				fmt.Printf(networkTestFailedMessage, networkTestErr.Error())
				os.Exit(1)
			} else {
				fmt.Println(networkTestSuccessMessage) //nolint
			}
		}()
	}

	httpsRedirectPort := "80"
	if val := os.Getenv("HTTPS_REDIRECT_PORT"); val != "" {
		httpsRedirectPort = val
	}

	if os.Getenv("HTTPS_REDIRECT_PORT") != "" || os.Getenv("ENABLE_HTTP_REDIRECT") != "" {
		go func() {
			redirectServer := &http.Server{
				Addr: ":" + httpsRedirectPort,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
				}),
			}

			log.Println("Running HTTP->HTTPS redirect Server at :" + httpsRedirectPort)
			log.Fatal(redirectServer.ListenAndServe())
		}()
	}

	mux := http.NewServeMux()
	if os.Getenv("DISABLE_FRONTEND") == "" {
		mux.Handle("/", indexHTMLWhenNotFound(http.Dir("./web/build")))
	}
	mux.HandleFunc("/api/whip", corsHandler(whipHandler))
	mux.HandleFunc("/api/whep", corsHandler(whepHandler))
	mux.HandleFunc("/api/sse/", corsHandler(whepServerSentEventsHandler))
	mux.HandleFunc("/api/layer/", corsHandler(whepLayerHandler))
	mux.HandleFunc("/api/status", corsHandler(statusHandler))
	mux.HandleFunc("/api/chat/connect", corsHandler(chatConnectHandler))
	mux.HandleFunc("/api/chat/sse/", corsHandler(chatSSEHandler))
	mux.HandleFunc("/api/chat/send/", corsHandler(chatSendHandler))

	server := &http.Server{
		Handler: mux,
		Addr:    os.Getenv("HTTP_ADDRESS"),
	}

	tlsKey := os.Getenv("SSL_KEY")
	tlsCert := os.Getenv("SSL_CERT")

	if tlsKey != "" && tlsCert != "" {
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{},
		}

		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			log.Fatal(err)
		}

		server.TLSConfig.Certificates = append(server.TLSConfig.Certificates, cert)

		log.Println("Running HTTPS Server at `" + os.Getenv("HTTP_ADDRESS") + "`")
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		log.Println("Running HTTP Server at `" + os.Getenv("HTTP_ADDRESS") + "`")
		log.Fatal(server.ListenAndServe())
	}
}
