package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	fileSize := stat.Size()

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Type", detectContentType(v.FilePath))
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		w.Header().Set("Accept-Ranges", "bytes")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
		return
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.SplitN(rangeStr, "-", 2)
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "invalid range")
		return
	}

	start, _ := strconv.ParseInt(parts[0], 10, 64)
	var end int64
	if parts[1] == "" {
		end = fileSize - 1
	} else {
		end, _ = strconv.ParseInt(parts[1], 10, 64)
	}

	if start > end || start < 0 || end >= fileSize {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		writeError(w, http.StatusRequestedRangeNotSatisfiable, "invalid range")
		return
	}

	chunkSize := end - start + 1
	w.Header().Set("Content-Type", detectContentType(v.FilePath))
	w.Header().Set("Content-Length", strconv.FormatInt(chunkSize, 10))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)

	file.Seek(start, io.SeekStart)
	io.CopyN(w, file, chunkSize)
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

func detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".flv":
		return "video/x-flv"
	case ".mpeg", ".mpg":
		return "video/mpeg"
	case ".ts", ".mts":
		return "video/mp2t"
	case ".ogv":
		return "video/ogg"
	default:
		return "application/octet-stream"
	}
}

func thumbnailPath(filePath string) string {
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	return filepath.Join(dir, ".ber", base+".png")
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
