package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ListenAddr  string
	LibraryPath string
	DBPath      string
	ScanOnStart bool
	ShowVersion bool
	TLSCert     string
	TLSKey      string
	MaxStreams  int
}

func ParseFlags() *Config {
	cfg := &Config{}

	defaultLib, _ := os.UserHomeDir()
	defaultLib = filepath.Join(defaultLib, "Videos")

	defaultDB := filepath.Join(defaultLib, ".ber", "library.db")

	flag.StringVar(&cfg.ListenAddr, "addr", ":8080", "listen address")
	flag.StringVar(&cfg.LibraryPath, "library", defaultLib, "path to video library directory")
	flag.StringVar(&cfg.DBPath, "db", defaultDB, "path to SQLite database file")
	flag.BoolVar(&cfg.ScanOnStart, "scan", false, "scan library on startup")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "show version and exit")
	flag.StringVar(&cfg.TLSCert, "tls-cert", "", "TLS certificate file path")
	flag.StringVar(&cfg.TLSKey, "tls-key", "", "TLS private key file path")
	flag.IntVar(&cfg.MaxStreams, "max-streams", 10, "maximum concurrent streams")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ber-server [flags]\n\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	return cfg
}
