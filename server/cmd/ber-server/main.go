package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/anomalyco/ber/internal/api"
	"github.com/anomalyco/ber/internal/config"
	"github.com/anomalyco/ber/internal/database"
	"github.com/anomalyco/ber/internal/library"
	"github.com/anomalyco/ber/internal/tray"
)

//go:embed all:web
var webFS embed.FS

//go:embed icon.ico
var iconBytes []byte

var Version = "dev"

func main() {
	logFile := openLog()
	if logFile != nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	cfg := config.ParseFlags()

	store, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	database.Migrate(store)

	lib := library.New(store, cfg.LibraryPath)
	if cfg.ScanOnStart {
		if err := lib.Scan(); err != nil {
			log.Printf("warning: initial scan failed: %v", err)
		}
	}

	webSub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("failed to get web subdirectory: %v", err)
	}

	apiMux := api.NewRouter(lib)
	mux := http.NewServeMux()
	mux.Handle("/api/", apiMux)
	mux.Handle("/", http.FileServer(http.FS(webSub)))

	// ponytail: middleware stack wraps everything
	var handler http.Handler = mux
	handler = api.CORS(handler)
	handler = api.Recoverer(handler)
	handler = api.Logger(handler)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

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

	// UDP beacon for client auto-discovery
	go beacon(cfg.ListenAddr)

	// Use a single quit channel for both signal and tray-quit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// ponytail: tray.Run blocks on the main goroutine,
	// running the Windows message loop.
	tray.Run(iconBytes, "ber-server", func(t *tray.Tray) {
		mDesktop := t.AddMenuItem("Open Desktop", "Launch the desktop app")
		mBrowser := t.AddMenuItem("Open in Browser", "Open web UI in your browser")
		t.AddSeparator()
		mQuit := t.AddMenuItem("Quit", "Shut down the server")

		go func() {
			for {
				select {
				case <-mDesktop.ClickedCh:
					launchDesktop()
				case <-mBrowser.ClickedCh:
					openBrowser("http://" + cfg.ListenAddr)
				case <-mQuit.ClickedCh:
					shutdown(srv, store)
					t.Quit()
					return
				case <-quit:
					shutdown(srv, store)
					t.Quit()
					return
				}
			}
		}()
	})
}

func beacon(listenAddr string) {
	addr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:10001")
	if err != nil {
		log.Printf("beacon: %v", err)
		return
	}
	conn, err := net.DialUDP("udp4", nil, addr)
	if err != nil {
		log.Printf("beacon: %v", err)
		return
	}
	defer conn.Close()
	msg := []byte("ber-server:" + listenAddr)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		conn.Write(msg)
	}
}

func launchDesktop() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("can't find server exe: %v", err)
		return
	}
	desktop := filepath.Join(filepath.Dir(exe), "ber-desktop.exe")
	if _, err := os.Stat(desktop); os.IsNotExist(err) {
		log.Printf("ber-desktop.exe not found at %s", desktop)
		return
	}
	cmd := exec.Command(desktop)
	cmd.Start()
}

func openBrowser(url string) {
	exec.Command("cmd", "/c", "start", url).Start()
}

func shutdown(srv *http.Server, store *database.Store) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	store.Close()
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
