package library

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/anomalyco/ber/internal/database"
)

// Video is re-exported from the database package for convenience.
type Video = database.Video

var supportedExts = map[string]bool{
	".mp4": true, ".m4v": true, ".mkv": true, ".avi": true,
	".mov": true, ".wmv": true, ".flv": true, ".webm": true,
	".mpeg": true, ".mpg": true, ".ts": true, ".mts": true,
	".3gp": true, ".ogv": true,
}

type Library struct {
	store       *database.Store
	libraryPath string
}

func New(store *database.Store, libraryPath string) *Library {
	return &Library{store: store, libraryPath: libraryPath}
}

func (l *Library) List() ([]*Video, error) {
	videos := l.store.List()
	slices.SortFunc(videos, func(a, b *Video) int {
		return strings.Compare(a.Title, b.Title)
	})
	return videos, nil
}

func (l *Library) Get(id string) (*Video, error) {
	return l.store.Get(id), nil
}

func (l *Library) Add(path string) (*Video, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a video file")
	}

	ext := strings.ToLower(filepath.Ext(path))
	if !supportedExts[ext] {
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	id, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	v := &database.Video{
		ID:        id,
		Title:     strings.TrimSuffix(filepath.Base(path), ext),
		FilePath:  path,
		FileSize:  info.Size(),
		Container: ext[1:],
	}

	return l.store.Add(v)
}

func (l *Library) Remove(id string) error {
	if !l.store.Remove(id) {
		return fmt.Errorf("video not found: %s", id)
	}
	return nil
}

func (l *Library) Scan() error {
	entries, err := os.ReadDir(l.libraryPath)
	if err != nil {
		return err
	}

	var errs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fullPath := filepath.Join(l.libraryPath, entry.Name())
		if _, err := l.Add(fullPath); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", entry.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("scan completed with %d errors:\n%s", len(errs), strings.Join(errs, "\n"))
	}
	return nil
}

func (l *Library) ScanDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}
		l.Add(path) // best-effort per file
		return nil
	})
}

func (l *Library) LibraryPath() string {
	return l.libraryPath
}

// newUUID generates a v4 UUID using crypto/rand.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
