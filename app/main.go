package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/valkey-io/valkey-go"
)

var counter int64
var client valkey.Client

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

	pass := os.Getenv("VALKEY_PASSWORD")
	if pwFile := os.Getenv("VALKEY_PASSWORD_FILE"); pwFile != "" {
		if data, err := os.ReadFile(pwFile); err == nil {
			pass = string(data)
			log.Println("Secret 파일에서 비밀번호 로드 완료")
		}
	}

	addr := os.Getenv("VALKEY_ADDR")
	if addr != "" {
		var err error
		for i := 0; i < 10; i++ {
			client, err = valkey.NewClient(valkey.ClientOption{
				InitAddress: []string{addr},
				Password:    pass,
			})
			if err == nil {
				log.Println("Valkey 연결 성공")
				break
			}
			log.Printf("Valkey 연결 재시도 %d/10: %v", i+1, err)
			time.Sleep(3 * time.Second)
		}
		if err != nil {
			log.Fatalf("Valkey 연결 실패: %v", err)
		}
		defer client.Close()
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
		var id int64
		if client != nil {
			cmd := client.B().Incr().Key("notiflex:id").Build()
			result := client.Do(r.Context(), cmd)
			if err := result.Error(); err != nil {
				log.Printf("Valkey INCR 에러: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			id, _ = result.AsInt64()
		} else {
			id = atomic.AddInt64(&counter, 1)
		}
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
		json.NewEncoder(w).Encode(VersionResponse{Version: "v0.5.0"})
	})

	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
