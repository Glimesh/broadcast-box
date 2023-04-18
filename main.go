package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"log"
	"net/http"

	"github.com/glimesh/broadcast-box/internal/webrtc"
	"github.com/joho/godotenv"
)

const (
	envFileProd = ".env.production"
	envFileDev  = ".env.development"
)

func logHTTPError(w http.ResponseWriter, err string, code int) {
	log.Println(err)
	http.Error(w, err, code)
}

func whipHandler(res http.ResponseWriter, r *http.Request) {
	streamKey := r.Header.Get("Authorization")
	if streamKey == "" {
		logHTTPError(res, "Authorization was not set", http.StatusBadRequest)
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

	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, answer)
}

func whepHandler(res http.ResponseWriter, req *http.Request) {
	streamKey := req.Header.Get("Authorization")
	if streamKey == "" {
		logHTTPError(res, "Authorization was not set", http.StatusBadRequest)
		return
	}

	offer, err := io.ReadAll(req.Body)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	answer, err := webrtc.WHEP(string(offer), streamKey)
	if err != nil {
		logHTTPError(res, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(res, answer)
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

		if req.Method != http.MethodOptions {
			next(res, req)
		}
	}
}

func main() {
	if os.Getenv("APP_ENV") == "production" {
		log.Println("Loading `" + envFileProd + "`")

		if err := godotenv.Load(envFileProd); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Loading `" + envFileDev + "`")

		if err := godotenv.Load(envFileDev); err != nil {
			log.Fatal(err)
		}
	}

	webrtc.Configure()

	mux := http.NewServeMux()
	mux.Handle("/", indexHTMLWhenNotFound(http.Dir("./web/build")))
	mux.HandleFunc("/api/whip", corsHandler(whipHandler))
	mux.HandleFunc("/api/whep", corsHandler(whepHandler))

	log.Println("Running HTTP Server at `" + os.Getenv("HTTP_ADDRESS") + "`")

	log.Fatal((&http.Server{
		Handler: mux,
		Addr:    os.Getenv("HTTP_ADDRESS"),
	}).ListenAndServe())
}
