package service_test

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dubyte/dir2opds/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func touchTestData(t *testing.T, root string, modTime time.Time) {
	t.Helper()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chtimes(path, modTime, modTime)
	})
	require.NoError(t, err)
}

func TestHandler(t *testing.T) {
	// pre-setup
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	touchTestData(t, "testdata", time.Date(2020, 5, 25, 0, 0, 0, 0, time.UTC))

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
		// mybook has 8 files but mybook.opf should be ignored
		assert.Len(t, catalog.Entries, 7)
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
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

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
  <feed xmlns="http://www.w3.org/2005/Atom" xmlns:opds="http://opds-spec.org/2010/catalog">
      <title>Catalog in /</title>
      <id>/</id>
      <link rel="start" href="/" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
      <link rel="self" href="/" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
      <link rel="http://opds-spec.org/facet" href="/?sort=name" type="application/atom+xml;profile=opds-catalog;kind=navigation" title="Name" xmlns:catalog="http://opds-spec.org/2010/catalog" catalog:facetGroup="Sort By" catalog:activeFacet="true"></link>
      <link rel="http://opds-spec.org/facet" href="/?sort=date" type="application/atom+xml;profile=opds-catalog;kind=navigation" title="Date" xmlns:catalog="http://opds-spec.org/2010/catalog" catalog:facetGroup="Sort By"></link>
      <link rel="http://opds-spec.org/facet" href="/?sort=size" type="application/atom+xml;profile=opds-catalog;kind=navigation" title="Size" xmlns:catalog="http://opds-spec.org/2010/catalog" catalog:facetGroup="Sort By"></link>
      <updated>2020-05-25T00:00:00+00:00</updated>
      <entry>
          <title>emptyFolder</title>
          <id>/emptyFolder</id>
          <link rel="subsection" href="/emptyFolder" type="application/atom+xml;profile=opds-catalog;kind=navigation" title="emptyFolder"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>mybook</title>
          <id>/mybook</id>
          <link rel="subsection" href="/mybook" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="mybook"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>new folder</title>
          <id>/new folder</id>
          <link rel="subsection" href="/new%20folder" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="new folder"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
  </feed>`

var acquisitionFeed = `<?xml version="1.0" encoding="UTF-8"?>
  <feed xmlns="http://www.w3.org/2005/Atom" xmlns:opds="http://opds-spec.org/2010/catalog" xmlns:dc="http://purl.org/dc/terms/">
      <title>Catalog in /mybook</title>
      <id>/mybook</id>
      <link rel="start" href="/" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
      <link rel="self" href="/mybook" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
      <link rel="up" href="/" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
      <link rel="http://opds-spec.org/facet" href="/mybook?sort=name" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="Name" xmlns:catalog="http://opds-spec.org/2010/catalog" catalog:facetGroup="Sort By" catalog:activeFacet="true"></link>
      <link rel="http://opds-spec.org/facet" href="/mybook?sort=date" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="Date" xmlns:catalog="http://opds-spec.org/2010/catalog" catalog:facetGroup="Sort By"></link>
      <link rel="http://opds-spec.org/facet" href="/mybook?sort=size" type="application/atom+xml;profile=opds-catalog;kind=acquisition" title="Size" xmlns:catalog="http://opds-spec.org/2010/catalog" catalog:facetGroup="Sort By"></link>
      <updated>2020-05-25T00:00:00+00:00</updated>
      <entry>
          <title>mybook copy.epub</title>
          <id>/mybookmybook copy.epub</id>
          <link rel="http://opds-spec.org/acquisition/open-access" href="/mybook/mybook%20copy.epub" type="application/epub+zip" title="mybook copy.epub"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>mybook copy.txt</title>
          <id>/mybookmybook copy.txt</id>
          <link rel="http://opds-spec.org/acquisition/open-access" href="/mybook/mybook%20copy.txt" type="text/plain; charset=utf-8" title="mybook copy.txt"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>mybook-win1251.fb2</title>
          <id>/mybookmybook-win1251.fb2</id>
          <link rel="http://opds-spec.org/acquisition/open-access" href="/mybook/mybook-win1251.fb2" type="text/fb2+xml; charset=utf-8" title="mybook-win1251.fb2"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>mybook.epub</title>
          <id>/mybookmybook.epub</id>
          <link rel="http://opds-spec.org/acquisition/open-access" href="/mybook/mybook.epub" type="application/epub+zip" title="mybook.epub"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>mybook.fb2</title>
          <id>/mybookmybook.fb2</id>
          <link rel="http://opds-spec.org/acquisition/open-access" href="/mybook/mybook.fb2" type="text/fb2+xml; charset=utf-8" title="mybook.fb2"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>mybook.pdf</title>
          <id>/mybookmybook.pdf</id>
          <link rel="http://opds-spec.org/acquisition/open-access" href="/mybook/mybook.pdf" type="application/pdf" title="mybook.pdf"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
      </entry>
      <entry>
          <title>mybook.txt</title>
          <id>/mybookmybook.txt</id>
          <link rel="http://opds-spec.org/acquisition/open-access" href="/mybook/mybook.txt" type="text/plain; charset=utf-8" title="mybook.txt"></link>
          <published>2020-05-25T00:00:00+00:00</published>
          <updated>2020-05-25T00:00:00+00:00</updated>
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

func TestCrawlableFeed(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	s := service.OPDS{
		TrustedRoot:      "testdata",
		HideCalibreFiles: true,
		HideDotFiles:     true,
		PageSize:         2,
	}

	t.Run("crawlable link appears in paginated feed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		service.TimeNow = func() time.Time {
			return time.Date(2020, 05, 25, 00, 00, 00, 0, time.UTC)
		}

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.Contains(t, string(body), `rel="http://opds-spec.org/crawlable"`)
		assert.Contains(t, string(body), `complete=true`)
	})

	t.Run("complete=true returns all entries unpaginated", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/?complete=true", nil)
		service.TimeNow = func() time.Time {
			return time.Date(2020, 05, 25, 00, 00, 00, 0, time.UTC)
		}

		err := s.Handler(w, req)
		require.NoError(t, err)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		assert.NotContains(t, string(body), `rel="first"`)
		assert.NotContains(t, string(body), `rel="next"`)
		assert.NotContains(t, string(body), `rel="http://opds-spec.org/crawlable"`)
		assert.Contains(t, string(body), "<title>emptyFolder</title>")
		assert.Contains(t, string(body), "<title>mybook</title>")
		assert.Contains(t, string(body), "<title>new folder</title>")
	})
}

func TestLastModifiedTracksChanges(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	processStart := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	service.TimeNow = func() time.Time {
		return processStart
	}

	for name, enableCache := range map[string]bool{
		"default mode": false,
		"enable-cache": true,
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			book1 := filepath.Join(root, "book1.epub")
			require.NoError(t, os.WriteFile(book1, []byte("epub1"), 0o644))
			require.NoError(t, os.Chtimes(book1, t1, t1))
			require.NoError(t, os.Chtimes(root, t1, t1))

			s := service.OPDS{
				TrustedRoot:      root,
				HideCalibreFiles: true,
				HideDotFiles:     true,
				EnableCache:      enableCache,
			}

			w1 := httptest.NewRecorder()
			req1 := httptest.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, s.Handler(w1, req1))

			resp1 := w1.Result()
			assert.Equal(t, http.StatusOK, resp1.StatusCode)
			lm1 := resp1.Header.Get("Last-Modified")
			etag1 := resp1.Header.Get("ETag")
			assert.Equal(t, t1.Format(http.TimeFormat), lm1)

			// Add a newer book with mtime after both processStart and t1
			book2 := filepath.Join(root, "book2.epub")
			require.NoError(t, os.WriteFile(book2, []byte("epub2"), 0o644))
			require.NoError(t, os.Chtimes(book2, t2, t2))
			require.NoError(t, os.Chtimes(root, t2, t2))

			w2 := httptest.NewRecorder()
			req2 := httptest.NewRequest(http.MethodGet, "/", nil)
			req2.Header.Set("If-Modified-Since", lm1)
			require.NoError(t, s.Handler(w2, req2))

			resp2 := w2.Result()
			body2, err := io.ReadAll(resp2.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp2.StatusCode)
			assert.Equal(t, t2.Format(http.TimeFormat), resp2.Header.Get("Last-Modified"))
			if enableCache {
				assert.NotEqual(t, etag1, resp2.Header.Get("ETag"))
			}
			assert.Contains(t, string(body2), "book1.epub")
			assert.Contains(t, string(body2), "book2.epub")
			assert.Contains(t, string(body2), "<updated>2026-01-01T12:00:00+00:00</updated>")
		})
	}
}

func TestLastModifiedFloorsAtProcessStart(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	oldFileTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	processStart := time.Date(2026, 2, 1, 8, 30, 0, 0, time.UTC)

	service.TimeNow = func() time.Time {
		return processStart
	}

	root := t.TempDir()
	book := filepath.Join(root, "book.epub")
	require.NoError(t, os.WriteFile(book, []byte("epub"), 0o644))
	require.NoError(t, os.Chtimes(book, oldFileTime, oldFileTime))
	require.NoError(t, os.Chtimes(root, oldFileTime, oldFileTime))

	s := service.OPDS{
		TrustedRoot:      root,
		HideCalibreFiles: true,
		HideDotFiles:     true,
		EnableCache:      true,
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, s.Handler(w, req))

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, processStart.Format(http.TimeFormat), resp.Header.Get("Last-Modified"))
	assert.Contains(t, string(body), "<updated>2025-01-01T00:00:00+00:00</updated>")
}

func TestSearchFeedHasNoLastModified(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	fixedStart := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	service.TimeNow = func() time.Time {
		return fixedStart
	}

	s := service.OPDS{
		TrustedRoot:      "testdata",
		HideCalibreFiles: true,
		HideDotFiles:     true,
		EnableSearch:     true,
		EnableCache:      true,
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search?q=mybook", nil)
	require.NoError(t, s.SearchHandler(w, req))

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Last-Modified"))
	assert.Empty(t, resp.Header.Get("ETag"))
	assert.Contains(t, string(body), "<updated>2026-03-01T09:00:00+00:00</updated>")
}

func TestETagVariesBySort(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	service.TimeNow = func() time.Time {
		return time.Date(2020, 5, 25, 0, 0, 0, 0, time.UTC)
	}

	s := service.OPDS{
		TrustedRoot:      "testdata",
		HideCalibreFiles: true,
		HideDotFiles:     true,
		EnableCache:      true,
	}

	wName := httptest.NewRecorder()
	reqName := httptest.NewRequest(http.MethodGet, "/?sort=name", nil)
	require.NoError(t, s.Handler(wName, reqName))
	etagName := wName.Result().Header.Get("ETag")
	require.NotEmpty(t, etagName)

	wDate := httptest.NewRecorder()
	reqDate := httptest.NewRequest(http.MethodGet, "/?sort=date", nil)
	require.NoError(t, s.Handler(wDate, reqDate))
	etagDate := wDate.Result().Header.Get("ETag")
	require.NotEmpty(t, etagDate)

	assert.NotEqual(t, etagName, etagDate, "sort variants must have distinct ETags")

	// Cross-variant If-None-Match must return 200, not 304
	wCross := httptest.NewRecorder()
	reqCross := httptest.NewRequest(http.MethodGet, "/?sort=date", nil)
	reqCross.Header.Set("If-None-Match", etagName)
	require.NoError(t, s.Handler(wCross, reqCross))
	assert.Equal(t, http.StatusOK, wCross.Result().StatusCode)
}

func TestETagVariesByVariant(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	service.TimeNow = func() time.Time {
		return time.Date(2020, 5, 25, 0, 0, 0, 0, time.UTC)
	}

	s := service.OPDS{
		TrustedRoot:      "testdata",
		HideCalibreFiles: true,
		HideDotFiles:     true,
		EnableCache:      true,
		EnableHTML:       true,
	}

	wXML := httptest.NewRecorder()
	reqXML := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, s.Handler(wXML, reqXML))
	etagXML := wXML.Result().Header.Get("ETag")
	require.NotEmpty(t, etagXML)

	wHTML := httptest.NewRecorder()
	reqHTML := httptest.NewRequest(http.MethodGet, "/", nil)
	reqHTML.Header.Set("Accept", "text/html")
	require.NoError(t, s.Handler(wHTML, reqHTML))
	etagHTML := wHTML.Result().Header.Get("ETag")
	require.NotEmpty(t, etagHTML)

	assert.NotEqual(t, etagXML, etagHTML, "HTML and XML responses must have distinct ETags")

	// Cross-variant If-None-Match must return 200, not 304
	wCross := httptest.NewRecorder()
	reqCross := httptest.NewRequest(http.MethodGet, "/", nil)
	reqCross.Header.Set("Accept", "text/html")
	reqCross.Header.Set("If-None-Match", etagXML)
	require.NoError(t, s.Handler(wCross, reqCross))
	assert.Equal(t, http.StatusOK, wCross.Result().StatusCode)
}

func TestSubSecondEditInvalidatesETag(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	processStart := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 1, 1, 11, 0, 0, 100_000_000, time.UTC)
	t2 := time.Date(2026, 1, 1, 11, 0, 0, 500_000_000, time.UTC)

	service.TimeNow = func() time.Time {
		return processStart
	}

	root := t.TempDir()
	book1 := filepath.Join(root, "book1.epub")
	require.NoError(t, os.WriteFile(book1, []byte("epub1"), 0o644))
	require.NoError(t, os.Chtimes(book1, t1, t1))
	require.NoError(t, os.Chtimes(root, t1, t1))

	s := service.OPDS{
		TrustedRoot:      root,
		HideCalibreFiles: true,
		HideDotFiles:     true,
		EnableCache:      true,
	}

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, s.Handler(w1, req1))
	etag1 := w1.Result().Header.Get("ETag")
	lm1 := w1.Result().Header.Get("Last-Modified")

	// Sub-second update within the same second
	require.NoError(t, os.Chtimes(book1, t2, t2))
	require.NoError(t, os.Chtimes(root, t2, t2))

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("If-None-Match", etag1)
	req2.Header.Set("If-Modified-Since", lm1)
	require.NoError(t, s.Handler(w2, req2))

	resp2 := w2.Result()
	assert.Equal(t, http.StatusOK, resp2.StatusCode, "new ETag must win over same-second If-Modified-Since")
	assert.NotEqual(t, etag1, resp2.Header.Get("ETag"))
}

func TestNoCacheSuppresses304(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	fixedStart := time.Date(2020, 5, 25, 0, 0, 0, 0, time.UTC)
	service.TimeNow = func() time.Time {
		return fixedStart
	}

	s := service.OPDS{
		TrustedRoot:      "testdata",
		HideCalibreFiles: true,
		HideDotFiles:     true,
		NoCache:          true,
		EnableCache:      true,
		EnableSearch:     true,
	}

	t.Run("feed with NoCache and EnableCache emits no validators and never returns 304", func(t *testing.T) {
		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, s.Handler(w1, req1))

		resp1 := w1.Result()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
		assert.Equal(t, "no-cache, no-store, must-revalidate", resp1.Header.Get("Cache-Control"))
		assert.Equal(t, "0", resp1.Header.Get("Expires"))
		assert.Empty(t, resp1.Header.Get("ETag"))
		assert.Empty(t, resp1.Header.Get("Last-Modified"))

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.Header.Set("If-Modified-Since", fixedStart.Format(http.TimeFormat))
		req2.Header.Set("If-None-Match", `W/"deadbeefdeadbeef"`)
		require.NoError(t, s.Handler(w2, req2))

		assert.Equal(t, http.StatusOK, w2.Result().StatusCode)
	})

	t.Run("search with NoCache sets Cache-Control headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/search?q=mybook", nil)
		require.NoError(t, s.SearchHandler(w, req))

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
		assert.Equal(t, "0", resp.Header.Get("Expires"))
		assert.Empty(t, resp.Header.Get("Last-Modified"))
	})
}

func writeTestEpubWithCover(t *testing.T, epubPath string) {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	opf := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
  </metadata>
  <manifest>
    <item id="cover-image" href="cover.jpg" media-type="image/jpeg"/>
  </manifest>
</package>`
	fw, err := zw.Create("content.opf")
	require.NoError(t, err)
	_, err = fw.Write([]byte(opf))
	require.NoError(t, err)

	cw, err := zw.Create("cover.jpg")
	require.NoError(t, err)
	_, err = cw.Write([]byte("\xff\xd8\xff\xe0fakejpeg"))
	require.NoError(t, err)

	require.NoError(t, zw.Close())
	require.NoError(t, os.WriteFile(epubPath, buf.Bytes(), 0o644))
}

func TestCoverHandlerCaching(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	processStart := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	fileTime := time.Date(2026, 1, 2, 15, 30, 45, 789_000_000, time.UTC)
	expectedLM := time.Date(2026, 1, 2, 15, 30, 45, 0, time.UTC).Format(http.TimeFormat)

	service.TimeNow = func() time.Time {
		return processStart
	}

	root := t.TempDir()
	epubPath := filepath.Join(root, "book.epub")
	writeTestEpubWithCover(t, epubPath)
	require.NoError(t, os.Chtimes(epubPath, fileTime, fileTime))

	t.Run("default mode sets truthful Last-Modified and answers 304", func(t *testing.T) {
		s := service.OPDS{TrustedRoot: root}

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/cover?file=/book.epub", nil)
		require.NoError(t, s.CoverHandler(w1, req1))

		resp1 := w1.Result()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
		assert.Equal(t, "max-age=86400", resp1.Header.Get("Cache-Control"))
		assert.Equal(t, expectedLM, resp1.Header.Get("Last-Modified"))

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/cover?file=/book.epub", nil)
		req2.Header.Set("If-Modified-Since", expectedLM)
		require.NoError(t, s.CoverHandler(w2, req2))

		assert.Equal(t, http.StatusNotModified, w2.Result().StatusCode)
	})

	t.Run("NoCache suppresses max-age, Last-Modified and 304", func(t *testing.T) {
		s := service.OPDS{TrustedRoot: root, NoCache: true}

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/cover?file=/book.epub", nil)
		require.NoError(t, s.CoverHandler(w1, req1))

		resp1 := w1.Result()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
		assert.Equal(t, "no-cache, no-store, must-revalidate", resp1.Header.Get("Cache-Control"))
		assert.Empty(t, resp1.Header.Get("Last-Modified"))

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/cover?file=/book.epub", nil)
		req2.Header.Set("If-Modified-Since", expectedLM)
		require.NoError(t, s.CoverHandler(w2, req2))

		assert.Equal(t, http.StatusOK, w2.Result().StatusCode)
	})
}

func TestHTMLCaching(t *testing.T) {
	nowFn := service.TimeNow
	defer func() {
		service.TimeNow = nowFn
	}()

	processStart := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	fileTime := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	expectedLM := fileTime.Format(http.TimeFormat)

	service.TimeNow = func() time.Time {
		return processStart
	}

	root := t.TempDir()
	book := filepath.Join(root, "book.epub")
	require.NoError(t, os.WriteFile(book, []byte("epub"), 0o644))
	require.NoError(t, os.Chtimes(book, fileTime, fileTime))
	require.NoError(t, os.Chtimes(root, fileTime, fileTime))

	t.Run("default mode HTML emits Last-Modified and answers 304", func(t *testing.T) {
		s := service.OPDS{
			TrustedRoot:      root,
			HideCalibreFiles: true,
			HideDotFiles:     true,
			EnableHTML:       true,
		}

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		req1.Header.Set("Accept", "text/html")
		require.NoError(t, s.Handler(w1, req1))

		resp1 := w1.Result()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
		assert.Equal(t, expectedLM, resp1.Header.Get("Last-Modified"))

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.Header.Set("Accept", "text/html")
		req2.Header.Set("If-Modified-Since", expectedLM)
		require.NoError(t, s.Handler(w2, req2))

		assert.Equal(t, http.StatusNotModified, w2.Result().StatusCode)
	})

	t.Run("NoCache HTML suppresses Last-Modified and 304", func(t *testing.T) {
		s := service.OPDS{
			TrustedRoot:      root,
			HideCalibreFiles: true,
			HideDotFiles:     true,
			EnableHTML:       true,
			NoCache:          true,
		}

		w1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		req1.Header.Set("Accept", "text/html")
		require.NoError(t, s.Handler(w1, req1))

		resp1 := w1.Result()
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
		assert.Empty(t, resp1.Header.Get("Last-Modified"))

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2.Header.Set("Accept", "text/html")
		req2.Header.Set("If-Modified-Since", expectedLM)
		require.NoError(t, s.Handler(w2, req2))

		assert.Equal(t, http.StatusOK, w2.Result().StatusCode)
	})
}
