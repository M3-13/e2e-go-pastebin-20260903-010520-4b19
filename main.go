package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"example.com/pastebin/handlers"
	"example.com/pastebin/store"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func newMux(api *handlers.API) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("POST /pastes", api.CreatePaste)
	mux.HandleFunc("GET /pastes/{id}", api.GetPaste)
	mux.HandleFunc("GET /pastes", api.ListPastes)
	mux.HandleFunc("DELETE /pastes/{id}", api.DeletePaste)
	return mux
}

func main() {
	addr := getenv("PASTEBIN_ADDR", ":8080")
	maxBodyBytesStr := getenv("PASTEBIN_MAX_BODY_BYTES", "1048576")
	maxBodyBytes, err := strconv.ParseInt(maxBodyBytesStr, 10, 64)
	if err != nil || maxBodyBytes <= 0 {
		log.Fatalf("invalid PASTEBIN_MAX_BODY_BYTES %q", maxBodyBytesStr)
	}

	st := store.NewStore()
	api := handlers.NewAPI(st, maxBodyBytes)

	srv := &http.Server{
		Addr:              addr,
		Handler:           newMux(api),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("pastebin listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
