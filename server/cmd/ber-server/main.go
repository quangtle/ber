package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anomalyco/ber/internal/api"
	"github.com/anomalyco/ber/internal/config"
	"github.com/anomalyco/ber/internal/database"
	"github.com/anomalyco/ber/internal/library"
)

//go:embed all:web
var webFS embed.FS

var Version = "dev"

func main() {
	cfg := config.ParseFlags()

	if cfg.ShowVersion {
		fmt.Printf("ber-server %s\n", Version)
		os.Exit(0)
	}

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	lib := library.New(db, cfg.LibraryPath)
	if cfg.ScanOnStart {
		if err := lib.Scan(); err != nil {
			log.Printf("warning: initial scan failed: %v", err)
		}
	}

	webSub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to get web subdirectory: %v", err)
	}

	mux := api.NewRouter(lib, cfg)
	mux.Handle("/*", http.FileServer(http.FS(webSub)))

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("ber-server %s listening on %s", Version, cfg.ListenAddr)
		if cfg.TLSCert != "" && cfg.TLSKey != "" {
			if err := srv.ListenAndServeTLS(cfg.TLSCert, cfg.TLSKey); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server error: %v", err)
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server error: %v", err)
			}
		}
	}()

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}
