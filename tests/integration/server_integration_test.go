//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/ber/internal/api"
	"github.com/anomalyco/ber/internal/database"
	"github.com/anomalyco/ber/internal/library"
)

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
	database.Migrate(db)

	lib := library.New(db, libPath)

	v, err := lib.Add(testFile)
	if err != nil {
		t.Fatal(err)
	}

	mux := api.NewRouter(lib)
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
