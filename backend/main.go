package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config := loadServerConfig()

	source := FileProductSource{
		MetadataPath: filepath.Join(config.DataDir, "metadata.json"),
		DetailsPath:  filepath.Join(config.DataDir, "details.json"),
	}
	popularitySource := FilePopularitySource{
		Path: filepath.Join(config.DataDir, "popularity.json"),
	}
	service := NewProductService(source, config.CacheTTL).WithPopularitySource(popularitySource)

	server := &http.Server{
		Addr:              config.Address(),
		Handler:           buildServerHandler(service, config.CORSAllowOrigin),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on http://%s", config.LogAddress())
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	log.Println("Server stopped")
}

func buildServerHandler(service *ProductService, corsAllowOrigin string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/products", NewProductHandler(service))
	mux.HandleFunc("/health", healthHandler)
	return withCORS(withLogging(mux), corsAllowOrigin)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
