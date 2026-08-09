package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type event struct {
	Event string `json:"event"`
}

func main() {
	url := flag.String("url", "ws://localhost:8080/ws", "WebSocket URL")
	cookie := flag.String("cookie", "", "session_id cookie value")
	want := flag.String("event", "", "comma-separated events to wait for; empty means any event")
	timeout := flag.Duration("timeout", 10*time.Second, "maximum wait time")
	flag.Parse()

	if *cookie == "" {
		log.Fatal("-cookie is required")
	}
	expected := map[string]bool{}
	for _, name := range strings.Split(*want, ",") {
		if name != "" {
			expected[name] = true
		}
	}

	header := http.Header{}
	header.Set("Cookie", "session_id="+*cookie)
	conn, _, err := websocket.DefaultDialer.Dial(*url, header)
	if err != nil {
		log.Fatalf("connect WebSocket: %v", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(*timeout)); err != nil {
		log.Fatalf("set read deadline: %v", err)
	}
	for {
		var value event
		if err := conn.ReadJSON(&value); err != nil {
			log.Fatalf("read event: %v", err)
		}
		fmt.Printf("event=%s\n", value.Event)
		if len(expected) == 0 {
			return
		}
		if expected[value.Event] {
			delete(expected, value.Event)
		}
		if len(expected) == 0 {
			return
		}
	}

}
