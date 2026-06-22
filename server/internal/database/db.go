package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Video is a persisted video record.
type Video struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	Duration  float64   `json:"duration"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Codec     string    `json:"codec"`
	Bitrate   int       `json:"bitrate"`
	Container string    `json:"container"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Store is a thread-safe, JSON-backed video database.
type Store struct {
	mu     sync.RWMutex
	byID   map[string]*Video
	byPath map[string]string // file_path → id
	path   string
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	s := &Store{
		byID:   make(map[string]*Video),
		byPath: make(map[string]string),
		path:   path,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return s, nil // fresh store
	}
	var list []*Video
	if err := json.Unmarshal(data, &list); err != nil {
		return s, nil // corrupt file, start fresh
	}
	for _, v := range list {
		s.byID[v.ID] = v
		s.byPath[v.FilePath] = v.ID
	}
	return s, nil
}

func (s *Store) Close() error { return s.save() }

func (s *Store) save() error {
	list := make([]*Video, 0, len(s.byID))
	for _, v := range s.byID {
		list = append(list, v)
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) List() []*Video {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Video, 0, len(s.byID))
	for _, v := range s.byID {
		out = append(out, v)
	}
	return out
}

func (s *Store) Get(id string) *Video {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.byID[id]
}

// Add upserts a video by file path, preserving the existing ID on conflict.
func (s *Store) Add(v *Video) (*Video, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingID, ok := s.byPath[v.FilePath]; ok {
		existing := s.byID[existingID]
		existing.Title = v.Title
		existing.FileSize = v.FileSize
		existing.Duration = v.Duration
		existing.Width = v.Width
		existing.Height = v.Height
		existing.Codec = v.Codec
		existing.Bitrate = v.Bitrate
		existing.Container = v.Container
		existing.UpdatedAt = time.Now()
		return existing, s.save()
	}

	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	s.byID[v.ID] = v
	s.byPath[v.FilePath] = v.ID
	return v, s.save()
}

func (s *Store) Remove(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.byID[id]
	if !ok {
		return false
	}
	delete(s.byPath, v.FilePath)
	delete(s.byID, id)
	s.save() // best-effort
	return true
}

// Migrate is a no-op for the JSON store — kept for API compatibility.
func Migrate(*Store) error { return nil }
