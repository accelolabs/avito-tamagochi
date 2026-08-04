package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", health)

	log.Printf("server started on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
