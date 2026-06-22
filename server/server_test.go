package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/anomalyco/ber/internal/api"
	"github.com/anomalyco/ber/internal/database"
	"github.com/anomalyco/ber/internal/library"
)

func setupTestServer(t *testing.T) (http.Handler, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")
	libPath := filepath.Join(tmpDir, "videos")
	os.MkdirAll(libPath, 0755)

	testFile := filepath.Join(libPath, "test.mp4")
	os.WriteFile(testFile, []byte("fake mp4 content"), 0644)

	store, err := database.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	database.Migrate(store)

	lib := library.New(store, libPath)

	mux := api.NewRouter(lib)
	return mux, testFile
}

func TestStatusEndpoint(t *testing.T) {
	mux, _ := setupTestServer(t)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var env struct {
		OK   bool                   `json:"ok"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if !env.OK {
		t.Fatal("expected ok=true")
	}
}

func TestLibraryEndpoints(t *testing.T) {
	mux, testFile := setupTestServer(t)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("list library (empty)", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/library")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("add to library", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"path": testFile})
		resp, err := http.Post(ts.URL+"/api/library/add",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("list library (with video)", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/library")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var env struct {
			OK   bool              `json:"ok"`
			Data []*library.Video  `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			t.Fatal(err)
		}
		if !env.OK {
			t.Fatal("expected ok=true")
		}
		if len(env.Data) != 1 {
			t.Fatalf("expected 1 video, got %d", len(env.Data))
		}
	})

	t.Run("scan library", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/api/library/scan", "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestStreamEndpoint(t *testing.T) {
	mux, _ := setupTestServer(t)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/stream/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}
