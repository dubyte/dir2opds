package service_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dubyte/dir2opds/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	// pre-setup
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	tests := map[string]struct {
		input             string
		want              string
		WantedContentType string
		wantedStatusCode  int
	}{
		"feed (dir of dirs )":                 {input: "/", want: feed, WantedContentType: "application/atom+xml;profile=opds-catalog;kind=navigation", wantedStatusCode: 200},
		"acquisitionFeed(dir of files)":       {input: "/mybook", want: acquisitionFeed, WantedContentType: "application/atom+xml;profile=opds-catalog;kind=acquisition", wantedStatusCode: 200},
		"servingAFile":                        {input: "/mybook/mybook.txt", want: "Fixture", WantedContentType: "text/plain; charset=utf-8", wantedStatusCode: 200},
		"serving file with spaces":            {input: "/mybook/mybook%20copy.txt", want: "Fixture", WantedContentType: "text/plain; charset=utf-8", wantedStatusCode: 200},
		"http trasversal vulnerability check": {input: "/../../../../mybook", want: feed, WantedContentType: "application/atom+xml;profile=opds-catalog;kind=navigation", wantedStatusCode: 404},
		"browser request (HTML)":              {input: "/", want: "dir2opds", WantedContentType: "text/html; charset=utf-8", wantedStatusCode: 200},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// setup
			s := service.OPDS{
				TrustedRoot:      "testdata",
				HideCalibreFiles: true,
				HideDotFiles:     true,
				NoCache:          true,
				EnableHTML:       strings.Contains(name, "browser request"),
			}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.input, nil)
			if strings.Contains(name, "browser request") {
				req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
			}
			service.TimeNow = func() time.Time {
				return time.Date(2020, 05, 25, 00, 00, 00, 0, time.UTC)
			}

			// act
			err := s.Handler(w, req)
			require.NoError(t, err)

			// post act
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// verify
			require.Equal(t, tc.wantedStatusCode, resp.StatusCode)
			if tc.wantedStatusCode != http.StatusOK {
				return
			}
			assert.Equal(t, tc.WantedContentType, resp.Header.Get("Content-Type"))
			if name == "browser request (HTML)" {
				assert.Contains(t, string(body), tc.want)
			} else {
				assert.Equal(t, tc.want, string(body))
			}
		})
	}

}

func TestScan(t *testing.T) {
	s := service.OPDS{TrustedRoot: "testdata", HideCalibreFiles: true, HideDotFiles: true}

	t.Run("Scan root (dir of dirs)", func(t *testing.T) {
		catalog, err := s.Scan("testdata", "/", 1)
		require.NoError(t, err)
		assert.Equal(t, "/", catalog.ID)
		// testdata has 3 folders: emptyFolder, mybook, new folder
		assert.Len(t, catalog.Entries, 3)
	})

	t.Run("Scan mybook (dir of files)", func(t *testing.T) {
		catalog, err := s.Scan("testdata/mybook", "/mybook", 1)
		require.NoError(t, err)
		assert.Equal(t, "/mybook", catalog.ID)
		// mybook has 6 files but mybook.opf should be ignored
		assert.Len(t, catalog.Entries, 5)
		for _, entry := range catalog.Entries {
			assert.NotContains(t, entry.Name, ".opf")
		}
	})

	t.Run("Scan empty folder", func(t *testing.T) {
		catalog, err := s.Scan("testdata/emptyFolder", "/emptyFolder", 1)
		require.NoError(t, err)
		assert.Empty(t, catalog.Entries)
	})
}

func TestBaseURL(t *testing.T) {
	s := service.OPDS{
		TrustedRoot: "testdata",
		BaseURL:     "https://opds.example.com",
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := s.Handler(w, req)
	require.NoError(t, err)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Contains(t, string(body), `href="https://opds.example.com/"`)
	assert.Contains(t, string(body), `href="https://opds.example.com/mybook"`)

	t.Run("Search with BaseURL", func(t *testing.T) {
		s.EnableSearch = true
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/search?q=mybook", nil)

		err := s.SearchHandler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Contains(t, string(body), `href="https://opds.example.com/mybook/mybook.epub"`)
	})

	t.Run("Search browser support", func(t *testing.T) {
		s.EnableSearch = true
		s.EnableHTML = true
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/search?q=mybook", nil)
		req.Header.Set("Accept", "text/html")

		err := s.SearchHandler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
		assert.Contains(t, string(body), "Search results for: mybook")
	})

	t.Run("OpenSearch with BaseURL", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/opensearch.xml", nil)

		s.OpenSearchHandler(w, req)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Contains(t, string(body), `template="https://opds.example.com/search?q={searchTerms}"`)
	})
}

func TestETagAndLastModified(t *testing.T) {
	s := service.OPDS{
		TrustedRoot:      "testdata",
		HideCalibreFiles: true,
		HideDotFiles:     true,
		EnableCache:      true,
	}

	t.Run("ETag header is set", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		service.TimeNow = func() time.Time {
			return time.Date(2020, 05, 25, 00, 00, 00, 0, time.UTC)
		}

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		assert.NotEmpty(t, resp.Header.Get("ETag"), "ETag header should be set")
		assert.Contains(t, resp.Header.Get("ETag"), `"`, "ETag should be quoted")
	})

	t.Run("Last-Modified header is set", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		assert.NotEmpty(t, resp.Header.Get("Last-Modified"), "Last-Modified header should be set")
	})

	t.Run("304 Not Modified with If-None-Match", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		service.TimeNow = func() time.Time {
			return time.Date(2020, 05, 25, 00, 00, 00, 0, time.UTC)
		}

		err := s.Handler(w, req)
		require.NoError(t, err)

		etag := w.Result().Header.Get("ETag")

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.Header.Set("If-None-Match", etag)

		err = s.Handler(w2, req2)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotModified, w2.Result().StatusCode)
	})

	t.Run("304 Not Modified with If-Modified-Since", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		err := s.Handler(w, req)
		require.NoError(t, err)

		lastModified := w.Result().Header.Get("Last-Modified")

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.Header.Set("If-Modified-Since", lastModified)

		err = s.Handler(w2, req2)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotModified, w2.Result().StatusCode)
	})
}

var feed = `<?xml version="1.0" encoding="UTF-8"?>
  <feed xmlns="http://www.w3.org/2005/Atom">
      <title>Catalog in /</title>
      <id>/</id>
      <link rel="start" href="/" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
      <updated>2020-05-25T00:00:00+00:00</updated>
      <entry>
          <title>emptyFolder</title>
          <id>/emptyFolder</id>
          <link rel="subsection" href="/emptyFolder" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="emptyFolder"></link>
          <published></published>
          <updated></updated>
      </entry>
      <entry>
          <title>mybook</title>
          <id>/mybook</id>
          <link rel="subsection" href="/mybook" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="mybook"></link>
          <published></published>
          <updated></updated>
      </entry>
      <entry>
          <title>new folder</title>
          <id>/new folder</id>
          <link rel="subsection" href="/new%20folder" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="new folder"></link>
          <published></published>
          <updated></updated>
      </entry>
  </feed>`

var acquisitionFeed = `<?xml version="1.0" encoding="UTF-8"?>
  <feed xmlns="http://www.w3.org/2005/Atom" xmlns:dc="http://purl.org/dc/terms/" xmlns:opds="http://opds-spec.org/2010/catalog">
      <title>Catalog in /mybook</title>
      <id>/mybook</id>
      <link rel="start" href="/" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
      <updated>2020-05-25T00:00:00+00:00</updated>
      <entry>
          <title>mybook copy.epub</title>
          <id>/mybookmybook copy.epub</id>
          <link rel="http://opds-spec.org/acquisition" href="/mybook/mybook%20copy.epub" type="application/epub+zip" title="mybook copy.epub"></link>
          <published></published>
          <updated></updated>
      </entry>
      <entry>
          <title>mybook copy.txt</title>
          <id>/mybookmybook copy.txt</id>
          <link rel="http://opds-spec.org/acquisition" href="/mybook/mybook%20copy.txt" type="text/plain; charset=utf-8" title="mybook copy.txt"></link>
          <published></published>
          <updated></updated>
      </entry>
      <entry>
          <title>mybook.epub</title>
          <id>/mybookmybook.epub</id>
          <link rel="http://opds-spec.org/acquisition" href="/mybook/mybook.epub" type="application/epub+zip" title="mybook.epub"></link>
          <published></published>
          <updated></updated>
      </entry>
      <entry>
          <title>mybook.pdf</title>
          <id>/mybookmybook.pdf</id>
          <link rel="http://opds-spec.org/acquisition" href="/mybook/mybook.pdf" type="application/pdf" title="mybook.pdf"></link>
          <published></published>
          <updated></updated>
      </entry>
      <entry>
          <title>mybook.txt</title>
          <id>/mybookmybook.txt</id>
          <link rel="http://opds-spec.org/acquisition" href="/mybook/mybook.txt" type="text/plain; charset=utf-8" title="mybook.txt"></link>
          <published></published>
          <updated></updated>
      </entry>
  </feed>`

func TestContentRange(t *testing.T) {
	s := service.OPDS{
		TrustedRoot:      "testdata",
		HideCalibreFiles: true,
		HideDotFiles:     true,
	}

	t.Run("Range request returns 206 Partial Content", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/mybook/mybook.txt", nil)
		req.Header.Set("Range", "bytes=0-3")

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusPartialContent, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Range"), "bytes 0-3/")
		assert.Equal(t, "Fixt", string(body))
	})

	t.Run("Range request with offset", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/mybook/mybook.txt", nil)
		req.Header.Set("Range", "bytes=4-")

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Equal(t, http.StatusPartialContent, resp.StatusCode)
		assert.Contains(t, resp.Header.Get("Content-Range"), "bytes 4-")
		assert.Equal(t, "ure", string(body))
	})

	t.Run("Accept-Ranges header is set for files", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/mybook/mybook.txt", nil)

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		assert.Equal(t, "bytes", resp.Header.Get("Accept-Ranges"))
	})

	t.Run("Invalid range returns 416 Requested Range Not Satisfiable", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/mybook/mybook.txt", nil)
		req.Header.Set("Range", "bytes=invalid")

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, resp.StatusCode)
	})
}
