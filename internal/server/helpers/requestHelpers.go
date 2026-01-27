package helpers

import "net/http"

func GetStreamKey(request *http.Request) (streamKey string) {
	queries := request.URL.Query()
	streamKey = queries.Get("key")

	return streamKey
}
