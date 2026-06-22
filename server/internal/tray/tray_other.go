//go:build !windows

package tray

import (
	"os"
	"os/signal"
	"syscall"
)

// Run creates a no-op tray that blocks until Quit is called or SIGINT/SIGTERM.
func Run(icon []byte, tooltip string, onReady func(*Tray)) {
	t := &Tray{icon: icon, tooltip: tooltip, quit: make(chan struct{}, 1)}
	onReady(t)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-t.quit:
	case <-sig:
	}
}

// Quit signals the tray to close.
func (t *Tray) Quit() {
	select {
	case t.quit <- struct{}{}:
	default:
	}
}
