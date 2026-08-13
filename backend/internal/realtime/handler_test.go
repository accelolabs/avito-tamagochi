package realtime

import (
	"net/http/httptest"
	"testing"
)

func TestSameOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin []string
		want   bool
	}{
		{name: "non-browser request", want: true},
		{name: "same host", origin: []string{"https://game.example"}, want: true},
		{name: "same host case insensitive", origin: []string{"https://GAME.EXAMPLE"}, want: true},
		{name: "foreign host", origin: []string{"https://evil.example"}, want: false},
		{name: "multiple origins", origin: []string{"https://game.example", "https://evil.example"}, want: false},
		{name: "malformed origin", origin: []string{"://bad"}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "http://game.example/ws", nil)
			for _, origin := range test.origin {
				request.Header.Add("Origin", origin)
			}
			if got := sameOrigin(request); got != test.want {
				t.Fatalf("sameOrigin() = %t, want %t", got, test.want)
			}
		})
	}
}
