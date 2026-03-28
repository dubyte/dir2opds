package service

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.gz.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	return w.gz.Close()
}

func (w *gzipResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
	w.gz.Flush()
}

func acceptsGzip(r *http.Request) bool {
	encodings := r.Header.Get("Accept-Encoding")
	for _, enc := range strings.Split(encodings, ",") {
		if strings.TrimSpace(enc) == "gzip" {
			return true
		}
	}
	return false
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !acceptsGzip(r) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			gz:             gz,
		}

		next.ServeHTTP(gzw, r)
	})
}
