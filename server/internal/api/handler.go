package api

import (
	"encoding/json"
	"runtime"
	"sort"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/anomalyco/ber/internal/config"
	"github.com/anomalyco/ber/internal/library"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type envelope struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeOK(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, envelope{OK: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, envelope{OK: false, Error: msg})
}

type Handler struct {
	lib *library.Library
	cfg *config.Config
}

func NewRouter(lib *library.Library, cfg *config.Config) *chi.Mux {
	h := &Handler{lib: lib, cfg: cfg}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(corsMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Get("/status", h.status)
		r.Get("/library", h.listLibrary)
		r.Get("/library/{id}", h.getVideo)
		r.Post("/library/scan", h.scanLibrary)
		r.Post("/library/add", h.addToLibrary)
		r.Delete("/library/{id}", h.removeFromLibrary)
		r.Get("/stream/{id}", h.streamVideo)
		r.Get("/stream/{id}/thumbnail", h.thumbnail)
		r.Get("/browse", h.browse)
		r.Post("/library/scan-dir", h.scanDir)
	})
	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) status(w http.ResponseWriter, _ *http.Request) {
	videos, err := h.lib.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeOK(w, map[string]interface{}{
		"version":      "dev",
		"library_path": h.lib.LibraryPath(),
		"video_count":  len(videos),
		"uptime":       "N/A",
	})
}

func (h *Handler) listLibrary(w http.ResponseWriter, _ *http.Request) {
	videos, err := h.lib.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if videos == nil {
		videos = []library.Video{}
	}
	writeOK(w, videos)
}

func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.lib.Get(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if v == nil {
		writeError(w, http.StatusNotFound, "video not found")
		return
	}
	writeOK(w, v)
}

func (h *Handler) scanLibrary(w http.ResponseWriter, _ *http.Request) {
	if err := h.lib.Scan(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, "scan completed")
}

func (h *Handler) addToLibrary(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	v, err := h.lib.Add(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeOK(w, v)
}

func (h *Handler) removeFromLibrary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.lib.Remove(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeOK(w, "removed")
}
func (h *Handler) browse(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		// list all drives on Windows
		drives := listDrives()
		writeOK(w, map[string]interface{}{
			"current": "",
			"parent":  "",
			"entries": drives,
		})
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read directory: "+err.Error())
		return
	}
	type entry struct {
		Name  string `json:"name"`
		IsDir bool   `json:"is_dir"`
		Path  string `json:"path"`
	}
	list := make([]entry, 0)
	for _, e := range entries {
		list = append(list, entry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Path:  filepath.Join(path, e.Name()),
		})
	}
	parent := filepath.Dir(path)
	if parent == path {
		parent = ""
	}
	writeOK(w, map[string]interface{}{
		"current": path,
		"parent":  parent,
		"entries": list,
	})
}

func listDrives() []map[string]interface{} {
	var list []map[string]interface{}
	if runtime.GOOS == "windows" {
		for _, d := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
			path := string(d) + ":\\"
			if _, err := os.Stat(path); err == nil {
				list = append(list, map[string]interface{}{
					"name": string(d) + ":",
					"is_dir": true,
					"path": path,
				})
			}
		}
		sort.Slice(list, func(i, j int) bool { return list[i]["name"].(string) < list[j]["name"].(string) })
	} else {
		list = append(list, map[string]interface{}{
			"name": "/",
			"is_dir": true,
			"path": "/",
		})
	}
	return list
}
func (h *Handler) scanDir(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}
	err := h.lib.ScanDir(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeOK(w, "scan completed")
}


func (h *Handler) streamVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.lib.Get(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if v == nil {
		writeError(w, http.StatusNotFound, "video not found")
		return
	}

	file, err := os.Open(v.FilePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot open file")
		return
	}
	defer file.Close()

	stat, _ := file.Stat()
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

func (h *Handler) thumbnail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.lib.Get(id)
	if err != nil || v == nil {
		writeError(w, http.StatusNotFound, "video not found")
		return
	}

	thumbPath := thumbnailPath(v.FilePath)
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, thumbPath)
}

func thumbnailPath(filePath string) string {
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	return filepath.Join(dir, ".ber", base+".png")
}

