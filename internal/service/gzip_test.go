package service

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcceptsGzip(t *testing.T) {
	tests := []struct {
		name     string
		encoding string
		wantGzip bool
	}{
		{"no encoding", "", false},
		{"gzip", "gzip", true},
		{"gzip with spaces", "gzip, deflate", true},
		{"gzip after other", "deflate, gzip", true},
		{"not gzip", "deflate", false},
		{"gzip with quality", "gzip;q=1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.encoding != "" {
				req.Header.Set("Accept-Encoding", tt.encoding)
			}
			assert.Equal(t, tt.wantGzip, acceptsGzip(req))
		})
	}
}

func TestGzipMiddleware(t *testing.T) {
	t.Run("without gzip acceptance", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello world"))
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		GzipMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "hello world", rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("with gzip acceptance", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello world"))
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		GzipMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Header().Get("Vary"), "Accept-Encoding")

		gr, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)
		defer gr.Close()

		body, err := io.ReadAll(gr)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(body))
	})

	t.Run("large content compression", func(t *testing.T) {
		content := strings.Repeat("a", 10000)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		GzipMiddleware(handler).ServeHTTP(rec, req)

		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Less(t, rec.Body.Len(), len(content), "Compressed size should be smaller")

		gr, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)
		defer gr.Close()

		body, err := io.ReadAll(gr)
		require.NoError(t, err)
		assert.Equal(t, content, string(body))
	})
}
