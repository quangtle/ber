package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/anomalyco/ber/internal/library"
)

type envelope struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

type Handler struct {
	lib *library.Library
}

// NewRouter builds the API mux with all routes. Middleware is applied separately.
func NewRouter(lib *library.Library) *http.ServeMux {
	h := &Handler{lib: lib}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/status", h.status)
	mux.HandleFunc("GET /api/library", h.listLibrary)
	mux.HandleFunc("GET /api/library/{id}", h.getVideo)
	mux.HandleFunc("POST /api/library/scan", h.scanLibrary)
	mux.HandleFunc("POST /api/library/add", h.addToLibrary)
	mux.HandleFunc("DELETE /api/library/{id}", h.removeFromLibrary)
	mux.HandleFunc("GET /api/stream/{id}", h.streamVideo)
	mux.HandleFunc("GET /api/stream/{id}/thumbnail", h.thumbnail)
	mux.HandleFunc("GET /api/browse", h.browse)
	mux.HandleFunc("POST /api/library/scan-dir", h.scanDir)

	return mux
}

// --- middleware -----------------------------------------------------------

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORS(next http.Handler) http.Handler {
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

// --- helpers --------------------------------------------------------------

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

// --- handlers -------------------------------------------------------------

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
		videos = []*library.Video{}
	}
	writeOK(w, videos)
}

func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
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
	id := r.PathValue("id")
	if err := h.lib.Remove(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeOK(w, "removed")
}

func (h *Handler) browse(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
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
			p := string(d) + ":\\"
			if _, err := os.Stat(p); err == nil {
				list = append(list, map[string]interface{}{
					"name":   string(d) + ":",
					"is_dir": true,
					"path":   p,
				})
			}
		}
		slices.SortFunc(list, func(a, b map[string]interface{}) int {
			return strings.Compare(a["name"].(string), b["name"].(string))
		})
	} else {
		list = append(list, map[string]interface{}{
			"name": "/", "is_dir": true, "path": "/",
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
	if err := h.lib.ScanDir(req.Path); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeOK(w, "scan completed")
}

func (h *Handler) streamVideo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	v, err := h.lib.Get(id)
	if err != nil || v == nil {
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
	id := r.PathValue("id")
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
