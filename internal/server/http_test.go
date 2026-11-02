package server

import (
	"net"
	"testing"

	"github.com/glimesh/broadcast-box/internal/environment"
)

func TestHTTPRedirectAddress(t *testing.T) {
	for _, test := range []struct {
		name  string
		value string
		want  string
	}{
		{"default", "", ":80"},
		{"bare port", "80", ":80"},
		{"colon port", ":80", ":80"},
		{"custom port", "8080", ":8080"},
		{"colon custom port", ":8080", ":8080"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(environment.HTTPSRedirectPort, test.value)
			address := getHTTPRedirectAddress()
			if address != test.want {
				t.Fatalf("got %q, want %q", address, test.want)
			}
			if _, _, err := net.SplitHostPort(address); err != nil {
				t.Fatalf("invalid listen address %q: %v", address, err)
			}
		})
	}
}
