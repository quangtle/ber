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
	"path/filepath"
	"syscall"
	"time"

	"github.com/anomalyco/ber/internal/api"
	"github.com/anomalyco/ber/internal/config"
	"github.com/anomalyco/ber/internal/database"
	"github.com/anomalyco/ber/internal/library"
	systray "github.com/getlantern/systray"
)

//go:embed all:web
var webFS embed.FS

//go:embed icon.png
var iconBytes []byte

var Version = "dev"

func main() {
	logFile := openLog()
	if logFile != nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	// ponytail: systray runs the Windows message loop on the main thread
	systray.Run(onReady, onExit)
}

func onReady() {
	cfg := config.ParseFlags()

	systray.SetIcon(iconBytes)
	systray.SetTooltip("ber-server")

	mStatus := systray.AddMenuItem(fmt.Sprintf("ber-server — %s", cfg.ListenAddr), "Server status")
	mStatus.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Shut down the server")

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

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

	go func() {
		log.Printf("ber-server %s listening on %s", Version, cfg.ListenAddr)
		mStatus.SetTitle(fmt.Sprintf("ber-server %s — listening on %s", Version, cfg.ListenAddr))
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

	// Graceful shutdown on SIGINT/SIGTERM (for when running with console)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("shutting down...")
		shutdown(srv, db)
		systray.Quit()
	}()

	// Quit via tray menu
	<-mQuit.ClickedCh
	log.Println("shutting down via tray...")
	shutdown(srv, db)
	systray.Quit()
}

func onExit() {
	// cleanup done in shutdown
}

func shutdown(srv *http.Server, db *database.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	db.Close()
}

func openLog() *os.File {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	logDir := filepath.Join(dir, ".ber")
	os.MkdirAll(logDir, 0755)
	f, err := os.OpenFile(filepath.Join(logDir, "server.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}
	return f
}
