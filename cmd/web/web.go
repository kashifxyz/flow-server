package web

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed default.html
var defaultHTML []byte

func Register(mux *http.ServeMux) {
	dir := staticDir()
	files := http.FileServer(http.Dir(dir))

	mux.HandleFunc("GET /{$}", serveIndex(dir))
	mux.Handle("GET /assets/", files)
	mux.Handle("GET /favicon.svg", files)
	mux.HandleFunc("GET /{path...}", spa(dir, files))
}

func serveIndex(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeIndex(w, dir)
	}
}

func spa(dir string, files http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if _, err := os.Stat(filepath.Join(dir, filepath.Clean(name))); err == nil {
				files.ServeHTTP(w, r)
				return
			}
		}
		writeIndex(w, dir)
	}
}

func writeIndex(w http.ResponseWriter, dir string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	built := filepath.Join(dir, "index.html")
	if b, err := os.ReadFile(built); err == nil && len(b) > 0 {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(defaultHTML)
}

func staticDir() string {
	if dir := strings.TrimSpace(os.Getenv("FLOW_WEB_DIR")); dir != "" {
		return dir
	}
	candidates := []string{
		filepath.Join("cmd", "web", "static"),
		filepath.Join("web", "static"),
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "web", "static"))
	}
	for _, dir := range candidates {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	return filepath.Join("cmd", "web", "static")
}
