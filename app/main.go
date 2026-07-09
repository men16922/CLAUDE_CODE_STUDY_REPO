package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

var counter int64

type HealthResponse struct {
	Status string `json:"status"`
}

type IdResponse struct {
	ID          string `json:"id"`
	GeneratedBy string `json:"generated_by"`
}

type VersionResponse struct {
	Version string `json:"version"`
}

func main() {
	podName := os.Getenv("HOSTNAME")
	if podName == "" {
		podName = "localhost"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
	})

	http.HandleFunc("/id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := atomic.AddInt64(&counter, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(IdResponse{
			ID:          fmt.Sprintf("%d", id),
			GeneratedBy: podName,
		})
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(VersionResponse{Version: "v0.1.1"})
	})

	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
