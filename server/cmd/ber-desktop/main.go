package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
)

func main() {
	startServer()

	// ponytail: Edge --app mode = desktop-app look without a real app
	edge := findEdge()
	if edge == "" {
		log.Fatalf("Microsoft Edge not found")
	}

	cmd := exec.Command(edge, "--app=http://localhost:8080")
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to launch Edge app: %v", err)
	}

	// wait for the window to close or Ctrl+C
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	done := make(chan struct{}, 1)
	go func() {
		cmd.Wait()
		close(done)
	}()
	select {
	case <-sigc:
		cmd.Process.Kill()
	case <-done:
	}
}

func startServer() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(exe)
	serverExe := filepath.Join(dir, "ber-server.exe")
	if _, err := os.Stat(serverExe); os.IsNotExist(err) {
		log.Printf("ber-server.exe not found at %s, assuming already running", serverExe)
		return
	}
	cmd := exec.Command(serverExe)
	cmd.Start()
}

func findEdge() string {
	paths := []string{
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// try PATH
	if p, err := exec.LookPath("msedge"); err == nil {
		return p
	}
	return ""
}
