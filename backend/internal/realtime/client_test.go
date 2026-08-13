package realtime

import (
	"testing"

	"github.com/gorilla/websocket"
)

func TestShouldLogReadError(t *testing.T) {
	if shouldLogReadError(&websocket.CloseError{Code: websocket.CloseNormalClosure}) {
		t.Fatal("normal close must not be logged as an error")
	}
	if !shouldLogReadError(&websocket.CloseError{Code: websocket.CloseProtocolError}) {
		t.Fatal("protocol close must be logged")
	}
}
