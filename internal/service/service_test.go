package service_test

import (
	"io"
	"net/http"
	"net/http/httptest"
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
	}{
		"feed (dir of dirs )":           {input: "/", want: feed, WantedContentType: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		"acquisitionFeed(dir of files)": {input: "/mybook", want: acquisitionFeed, WantedContentType: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
		"servingAFile":                  {input: "/mybook/mybook.txt", want: "Fixture", WantedContentType: "text/plain; charset=utf-8"},
		"serving file with spaces":      {input: "/mybook/mybook%20copy.txt", want: "Fixture", WantedContentType: "text/plain; charset=utf-8"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// setup
			s := service.OPDS{"testdata", true}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.input, nil)
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
			assert.Equal(t, 200, resp.StatusCode)
			assert.Equal(t, tc.WantedContentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tc.want, string(body))
		})
	}

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
