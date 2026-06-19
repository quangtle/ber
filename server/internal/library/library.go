package library

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var supportedExts = map[string]bool{
	".mp4":  true,
	".m4v":  true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
	".mpeg": true,
	".mpg":  true,
	".ts":   true,
	".mts":  true,
	".3gp":  true,
	".ogv":  true,
}

type Video struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	FilePath  string  `json:"file_path"`
	FileSize  int64   `json:"file_size"`
	Duration  float64 `json:"duration"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Codec     string  `json:"codec"`
	Bitrate   int     `json:"bitrate"`
	Container string  `json:"container"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type Library struct {
	db          *sql.DB
	libraryPath string
}

func New(db *sql.DB, libraryPath string) *Library {
	return &Library{db: db, libraryPath: libraryPath}
}

func (l *Library) List() ([]Video, error) {
	rows, err := l.db.Query(`
		SELECT id, title, file_path, file_size, duration,
		       width, height, codec, bitrate, container,
		       created_at, updated_at
		FROM videos ORDER BY title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.Title, &v.FilePath, &v.FileSize,
			&v.Duration, &v.Width, &v.Height, &v.Codec,
			&v.Bitrate, &v.Container, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, rows.Err()
}

func (l *Library) Get(id string) (*Video, error) {
	var v Video
	err := l.db.QueryRow(`
		SELECT id, title, file_path, file_size, duration,
		       width, height, codec, bitrate, container,
		       created_at, updated_at
		FROM videos WHERE id = ?
	`, id).Scan(&v.ID, &v.Title, &v.FilePath, &v.FileSize,
		&v.Duration, &v.Width, &v.Height, &v.Codec,
		&v.Bitrate, &v.Container, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (l *Library) Add(path string) (*Video, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access path: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, use AddDir")
	}

	ext := strings.ToLower(filepath.Ext(path))
	if !supportedExts[ext] {
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	v := &Video{
		ID:       uuid.New().String(),
		Title:    strings.TrimSuffix(filepath.Base(path), ext),
		FilePath: path,
		FileSize: info.Size(),
		Duration: 0,
		Width:    0,
		Height:   0,
		Codec:    "",
		Bitrate:  0,
		Container: ext[1:],
	}

	_, err = l.db.Exec(`
		INSERT INTO videos (id, title, file_path, file_size, duration,
		                    width, height, codec, bitrate, container)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			title=excluded.title, file_size=excluded.file_size,
			updated_at=datetime('now')
	`, v.ID, v.Title, v.FilePath, v.FileSize, v.Duration,
		v.Width, v.Height, v.Codec, v.Bitrate, v.Container)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (l *Library) Remove(id string) error {
	res, err := l.db.Exec("DELETE FROM videos WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
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
		return fmt.Errorf("scan completed with %d errors:\n%s",
			len(errs), strings.Join(errs, "\n"))
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
