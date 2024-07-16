package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/tls"
	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/networktest"
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
	webhookURL     string
	webhookTimeout int
)

var noBuildDirectoryErr = errors.New("\033[0;31mBuild directory does not exist, run `npm install` and `npm run build` in the web directory.\033[0m")

type (
	whepLayerRequestJSON struct {
		MediaId    string `json:"mediaId"`
		EncodingId string `json:"encodingId"`
	}
)

func logHTTPError(w http.ResponseWriter, err string, code int) {
	log.Println(err)
	http.Error(w, err, code)
}

func validateStreamKey(streamKey string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_\-\.~]+$`).MatchString(streamKey)
}

func extractBearerToken(authHeader string) (string, bool) {
	const bearerPrefix = "Bearer "
	if strings.HasPrefix(authHeader, bearerPrefix) {
		return strings.TrimPrefix(authHeader, bearerPrefix), true
	}
	return "", false
}

func whipHandler(res http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		return
	}

	streamKeyHeader := r.Header.Get("Authorization")
	if streamKeyHeader == "" {
		logHTTPError(res, "Authorization was not set", http.StatusBadRequest)
		return
	}

	streamKey, ok := extractBearerToken(streamKeyHeader)
	if !ok || !validateStreamKey(streamKey) {
		logHTTPError(res, "Invalid stream key format", http.StatusBadRequest)
		return
	}

	// Prepare webhook payload
	payload := prepareWebhookPayload("publish", streamKey, r)
	statusCode, err := handleWebhook(payload)
	if err != nil {
		logHTTPError(res, err.Error(), statusCode)
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
	fmt.Fprint(res, answer)
}

func whepHandler(res http.ResponseWriter, req *http.Request) {
	streamKeyHeader := req.Header.Get("Authorization")
	if streamKeyHeader == "" {
		logHTTPError(res, "Authorization was not set", http.StatusBadRequest)
		return
	}

	streamKey, ok := extractBearerToken(streamKeyHeader)
	if !ok || !validateStreamKey(streamKey) {
		logHTTPError(res, "Invalid stream key format", http.StatusBadRequest)
		return
	}

	// Prepare webhook payload
	payload := prepareWebhookPayload("view", streamKey, req)
	statusCode, err := handleWebhook(payload)
	if err != nil {
		logHTTPError(res, err.Error(), statusCode)
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
	fmt.Fprint(res, answer)
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

	fmt.Fprint(res, "event: layers\n")
	fmt.Fprintf(res, "data: %s\n", string(layers))
	fmt.Fprint(res, "\n\n")
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
	res.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(res).Encode(webrtc.GetStreamStatuses()); err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
	}
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

func main() {
	loadConfigs := func() error {
		if os.Getenv("APP_ENV") == "development" {
			log.Println("Loading `" + envFileDev + "`")
			return godotenv.Load(envFileDev)
		} else {
			log.Println("Loading `" + envFileProd + "`")
			if err := godotenv.Load(envFileProd); err != nil {
				return err
			}

			if _, err := os.Stat("./web/build"); os.IsNotExist(err) && os.Getenv("DISABLE_FRONTEND") == "" {
				return noBuildDirectoryErr
			}

			return nil
		}
	}

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

	webhookURL = os.Getenv("WEBHOOK_URL")
	webhookTimeout, _ := strconv.Atoi(os.Getenv("WEBHOOK_TIMEOUT"))
	if webhookTimeout == 0 {
		webhookTimeout = 5000 // Default to 5 seconds if not set or invalid
	}

	mux := http.NewServeMux()
	if os.Getenv("DISABLE_FRONTEND") == "" {
		mux.Handle("/", indexHTMLWhenNotFound(http.Dir("./web/build")))
	}
	mux.HandleFunc("/api/whip", corsHandler(whipHandler))
	mux.HandleFunc("/api/whep", corsHandler(whepHandler))
	mux.HandleFunc("/api/sse/", corsHandler(whepServerSentEventsHandler))
	mux.HandleFunc("/api/layer/", corsHandler(whepLayerHandler))

	if os.Getenv("DISABLE_STATUS") == "" {
		mux.HandleFunc("/api/status", corsHandler(statusHandler))
	}

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

func prepareWebhookPayload(action, streamKey string, r *http.Request) webrtc.WebhookPayload {
	// Convert url.Values to map[string]string
	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	return webrtc.WebhookPayload{
		Action:      action,
		StreamKey:   streamKey,
		IP:          getIPAddress(r),
		BearerToken: streamKey,
		QueryParams: queryParams,
		UserAgent:   r.UserAgent(),
	}
}

func handleWebhook(payload webrtc.WebhookPayload) (int, error) {
	if webhookURL == "" {
		return http.StatusOK, nil
	}

	statusCode, err := webrtc.CallWebhook(webhookURL, webhookTimeout, payload)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Webhook call failed: %w", err)
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return statusCode, fmt.Errorf("Webhook denied access")
	}
	if statusCode != http.StatusOK {
		return statusCode, fmt.Errorf("Webhook returned unexpected status")
	}
	return http.StatusOK, nil
}

func getIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}
