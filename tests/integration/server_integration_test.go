//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/anomalyco/ber/internal/api"
	"github.com/anomalyco/ber/internal/config"
	"github.com/anomalyco/ber/internal/database"
	"github.com/anomalyco/ber/internal/library"
)

func TestLibraryScanWithProbe(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found, skipping integration test")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	libPath := filepath.Join(tmpDir, "videos")
	os.MkdirAll(libPath, 0755)

	fixturePath := filepath.Join("..", "tests", "fixtures", "sample_h264.mp4")
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("test fixture not found, skipping")
	}

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(libPath, "sample.mp4"), data, 0644)

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.Migrate()

	lib := library.New(db, libPath)
	ffmpegProbe := func(path string) (*library.Video, error) {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		return &library.Video{
			Title:    filepath.Base(path),
			FilePath: path,
			FileSize: info.Size(),
			Duration: 1.0,
			Width:    640,
			Height:   480,
			Codec:    "h264",
			Container: "mp4",
		}, nil
	}

	if err := lib.ScanWithProbe(ffmpegProbe); err != nil {
		t.Fatal(err)
	}

	videos, err := lib.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 1 {
		t.Fatalf("expected 1 video, got %d", len(videos))
	}
	if videos[0].Codec != "h264" {
		t.Fatalf("expected h264 codec, got %s", videos[0].Codec)
	}
}

func TestStreamRangeRequestIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	libPath := filepath.Join(tmpDir, "videos")
	os.MkdirAll(libPath, 0755)

	testContent := []byte("fake video content for range request testing")
	testFile := filepath.Join(libPath, "test.mp4")
	os.WriteFile(testFile, testContent, 0644)

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.Migrate()

	lib := library.New(db, libPath)
	cfg := &config.Config{LibraryPath: libPath, DBPath: dbPath}

	v, err := lib.Add(testFile)
	if err != nil {
		t.Fatal(err)
	}

	mux := api.NewRouter(lib, cfg)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("full file request", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/stream/" + v.ID)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("range request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", ts.URL+"/api/stream/"+v.ID, nil)
		req.Header.Set("Range", "bytes=0-4")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusPartialContent {
			t.Fatalf("expected 206, got %d", resp.StatusCode)
		}
	})
}
