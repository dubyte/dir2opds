package service

import (
	"archive/zip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	trustedRoot := filepath.Join(wd, "testdata")

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", filepath.Join(trustedRoot, "mybook"), false},
		{"valid path with dots", filepath.Join(trustedRoot, "mybook", ".", "mybook.txt"), false},
		{"traversal attack", filepath.Join(trustedRoot, "..", "..", "etc", "passwd"), true},
		{"path outside root (prefix match)", filepath.Join(wd, "testdata_extra"), true},
		{"path is root", trustedRoot, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := verifyPath(tt.path, trustedRoot)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInTrustedRoot(t *testing.T) {
	root := "/home/books"

	assert.True(t, inTrustedRoot("/home/books", root))
	assert.True(t, inTrustedRoot("/home/books/folder", root))
	assert.False(t, inTrustedRoot("/home/bookkeeping", root))
	assert.False(t, inTrustedRoot("/etc/passwd", root))
}

func TestFileShouldBeIgnored(t *testing.T) {
	assert.False(t, fileShouldBeIgnored("book.epub", true, true))
	assert.True(t, fileShouldBeIgnored(".hidden", true, true))
	assert.False(t, fileShouldBeIgnored(".hidden", true, false))
	assert.True(t, fileShouldBeIgnored("metadata.opf", true, true))
	assert.False(t, fileShouldBeIgnored("metadata.opf", false, true))
	assert.False(t, fileShouldBeIgnored(".", true, true))
	assert.False(t, fileShouldBeIgnored("..", true, true))
}

func TestSortEntries(t *testing.T) {
	now := time.Now()
	entries := []CatalogEntry{
		{Name: "B", Size: 100, ModTime: now.Add(-time.Hour)},
		{Name: "A", Size: 200, ModTime: now},
		{Name: "C", Size: 50, ModTime: now.Add(-2 * time.Hour)},
	}

	s := OPDS{SortBy: "name"}
	s.sortEntries(entries)
	assert.Equal(t, "A", entries[0].Name)

	s.SortBy = "size"
	s.sortEntries(entries)
	assert.Equal(t, "A", entries[0].Name) // Size 200

	s.SortBy = "date"
	s.sortEntries(entries)
	assert.Equal(t, "A", entries[0].Name) // Most recent
}

func TestExtractMetadata(t *testing.T) {
	t.Run("Extract EPUB", func(t *testing.T) {
		path := filepath.Join("testdata", "mybook", "mybook.epub")
		title, author, coverPath, description, series, seriesIndex, subjects := extractEpubMetadata(path)
		t.Logf("EPUB Title: %q, Author: %q, CoverPath: %q, Description: %q, Series: %q, SeriesIndex: %q, Subjects: %v", title, author, coverPath, description, series, seriesIndex, subjects)
		assert.Equal(t, "Unknown Title", title)
		assert.Equal(t, "Unknown Author", author)
	})

	t.Run("Extract FB2", func(t *testing.T) {
		path := filepath.Join("testdata", "mybook", "mybook.fb2")
		title, author, coverPath, description, series, seriesIndex, subjects := extractFb2Metadata(path)
		t.Logf("FB2 Title: %q, Author: %q, CoverPath: %q, Description: %q, Series: %q, SeriesIndex: %q, Subjects: %v", title, author, coverPath, description, series, seriesIndex, subjects)
		assert.Equal(t, "Unknown Title", title)
		// Multiple <author> elements are all captured, in document order;
		// nickname-only authors (e.g. pseudonyms) are kept too.
		assert.Equal(t, "Unknown Author, Jane Q Doe, Penn", author)
		// <annotation> text lives in <p> elements; both paragraphs are joined
		// and inline markup such as <emphasis> contributes its text.
		assert.Equal(t, "First paragraph of the annotation. Second paragraph of the annotation with markup inside.", description)
		assert.Equal(t, []string{"unrecognised"}, subjects)
	})

	t.Run("Extract FB2 with non-UTF-8 encoding", func(t *testing.T) {
		path := filepath.Join("testdata", "mybook", "mybook-win1251.fb2")
		title, author, coverPath, description, series, seriesIndex, subjects := extractFb2Metadata(path)
		t.Logf("FB2 Title: %q, Author: %q, CoverPath: %q, Description: %q, Series: %q, SeriesIndex: %q, Subjects: %v", title, author, coverPath, description, series, seriesIndex, subjects)
		assert.Equal(t, "Война и мир", title)
		assert.Equal(t, "Лев Николаевич Толстой", author)
		assert.Empty(t, description)
	})

	t.Run("Extract FB2 from malformed file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "broken.fb2")
		require.NoError(t, os.WriteFile(path, []byte("<FictionBook><description><title-info><book-title>Broken"), 0o644))
		title, author, coverPath, description, series, seriesIndex, subjects := extractFb2Metadata(path)
		t.Logf("FB2 Title: %q, Author: %q, CoverPath: %q, Description: %q, Series: %q, SeriesIndex: %q, Subjects: %v", title, author, coverPath, description, series, seriesIndex, subjects)
		assert.Empty(t, title)
		assert.Empty(t, author)
		assert.Empty(t, coverPath)
		assert.Empty(t, description)
		assert.Empty(t, series)
		assert.Empty(t, seriesIndex)
		assert.Nil(t, subjects)
	})

	t.Run("Extract PDF", func(t *testing.T) {
		path := filepath.Join("testdata", "mybook", "mybook.pdf")
		title, author, description, subjects := extractPdfMetadata(path)
		t.Logf("PDF Title: %q, Author: %q, Description: %q, Subjects: %v", title, author, description, subjects)
	})
}

func TestParsePage(t *testing.T) {
	assert.Equal(t, 1, parsePage(""))
	assert.Equal(t, 1, parsePage("invalid"))
	assert.Equal(t, 1, parsePage("0"))
	assert.Equal(t, 1, parsePage("-1"))
	assert.Equal(t, 1, parsePage("1"))
	assert.Equal(t, 5, parsePage("5"))
	assert.Equal(t, 100, parsePage("100"))
}

func TestPageSize(t *testing.T) {
	s := OPDS{}
	assert.Equal(t, defaultPageSize, s.pageSize())

	s.PageSize = 10
	assert.Equal(t, 10, s.pageSize())

	s.PageSize = 500
	assert.Equal(t, maxPageSize, s.pageSize())

	s.PageSize = 0
	assert.Equal(t, defaultPageSize, s.pageSize())
}

func TestFb2Text(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     string
	}{
		{"plain text", "First paragraph.", "First paragraph."},
		{"mid-word markup keeps adjacency", "anti<emphasis>hero</emphasis> story", "antihero story"},
		{"markup between words", "one <a>link</a> two", "one link two"},
		{"whitespace collapsed", "  spaced\n\t out  ", "spaced out"},
		{"empty fragment", "", ""},
		{"whitespace only", "   ", ""},
		{"entity unescaping", "AT&amp;T &lt;tag&gt;", "AT&T <tag>"},
		{"malformed fragment is dropped", "<b>unclosed", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fb2Text(tt.fragment))
		})
	}
}

func TestFb2Name(t *testing.T) {
	assert.Equal(t, "Lev Nikolayevich Tolstoy", fb2Name(fb2Author{FirstName: "Lev", MiddleName: "Nikolayevich", LastName: "Tolstoy"}))
	assert.Equal(t, "Lev Tolstoy", fb2Name(fb2Author{FirstName: "Lev", LastName: "Tolstoy", Nickname: "LNT"}), "name parts take precedence over nickname")
	assert.Equal(t, "Penn", fb2Name(fb2Author{Nickname: "Penn"}), "nickname-only author")
	assert.Equal(t, "Penn", fb2Name(fb2Author{Nickname: "  Penn  "}), "nickname is trimmed")
	assert.Equal(t, "", fb2Name(fb2Author{Nickname: "   "}), "whitespace-only nickname")
	assert.Equal(t, "", fb2Name(fb2Author{}), "empty author")
}

func TestPagination(t *testing.T) {
	s := OPDS{TrustedRoot: "testdata", HideCalibreFiles: true, HideDotFiles: true}

	t.Run("First page", func(t *testing.T) {
		catalog, err := s.Scan("testdata/mybook", "/mybook", 1)
		require.NoError(t, err)
		assert.Equal(t, 1, catalog.Page)
		assert.Equal(t, defaultPageSize, catalog.PageSize)
		assert.Equal(t, 7, catalog.Total)
	})

	t.Run("Page with small page size", func(t *testing.T) {
		s.PageSize = 2
		catalog, err := s.Scan("testdata/mybook", "/mybook", 1)
		require.NoError(t, err)
		assert.Equal(t, 1, catalog.Page)
		assert.Equal(t, 2, catalog.PageSize)
		assert.Equal(t, 7, catalog.Total)
		assert.Len(t, catalog.Entries, 2)
	})

	t.Run("Second page", func(t *testing.T) {
		s.PageSize = 2
		catalog, err := s.Scan("testdata/mybook", "/mybook", 2)
		require.NoError(t, err)
		assert.Equal(t, 2, catalog.Page)
		assert.Equal(t, 7, catalog.Total)
		assert.Len(t, catalog.Entries, 2)
	})

	t.Run("Last page with partial entries", func(t *testing.T) {
		s.PageSize = 2
		catalog, err := s.Scan("testdata/mybook", "/mybook", 3)
		require.NoError(t, err)
		assert.Equal(t, 3, catalog.Page)
		assert.Equal(t, 7, catalog.Total)
		assert.Len(t, catalog.Entries, 2)
	})

	t.Run("Page beyond total", func(t *testing.T) {
		s.PageSize = 2
		catalog, err := s.Scan("testdata/mybook", "/mybook", 100)
		require.NoError(t, err)
		assert.Equal(t, 100, catalog.Page)
		assert.Empty(t, catalog.Entries)
	})
}

func TestBuildPageURL(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		query    map[string]string
		page     int
		want     string
	}{
		{
			name:     "simple path",
			basePath: "/",
			query:    map[string]string{},
			page:     1,
			want:     "/?page=1",
		},
		{
			name:     "path with existing query",
			basePath: "/mybook",
			query:    map[string]string{"q": "test"},
			page:     2,
			want:     "/mybook?page=2&q=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := make(url.Values)
			for k, v := range tt.query {
				values.Set(k, v)
			}
			result := buildPageURL(tt.basePath, values, tt.page)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestEtag(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	etag1 := etag("/path", t1, 1)
	etag2 := etag("/path", t1, 1)
	assert.Equal(t, etag1, etag2, "same inputs should produce same etag")

	etag3 := etag("/path", t2, 1)
	assert.NotEqual(t, etag1, etag3, "different time should produce different etag")

	etag4 := etag("/path", t1, 2)
	assert.NotEqual(t, etag1, etag4, "different page should produce different etag")

	etag5 := etag("/other", t1, 1)
	assert.NotEqual(t, etag1, etag5, "different path should produce different etag")

	assert.True(t, strings.HasPrefix(etag1, `"`), "etag should be quoted")
	assert.True(t, strings.HasSuffix(etag1, `"`), "etag should be quoted")
}

func TestCatalogModTime(t *testing.T) {
	s := OPDS{TrustedRoot: "testdata", HideCalibreFiles: true, HideDotFiles: true}

	catalog, err := s.Scan("testdata/mybook", "/mybook", 1)
	require.NoError(t, err)
	assert.False(t, catalog.ModTime.IsZero(), "ModTime should be set")
}

func TestExtractEpubCover(t *testing.T) {
	t.Run("Valid EPUB", func(t *testing.T) {
		path := filepath.Join("testdata", "mybook", "mybook.epub")
		data, contentType, err := extractEpubCover(path)
		require.NoError(t, err)
		t.Logf("Cover content-type: %s, size: %d bytes", contentType, len(data))
	})

	t.Run("Non-existent file", func(t *testing.T) {
		_, _, err := extractEpubCover(filepath.Join("testdata", "nonexistent.epub"))
		assert.Error(t, err)
	})
}

func TestFindEpubCover(t *testing.T) {
	t.Run("EPUB with cover", func(t *testing.T) {
		path := filepath.Join("testdata", "mybook", "mybook.epub")
		r, err := zip.OpenReader(path)
		require.NoError(t, err)
		defer r.Close()

		var opfPath string
		for _, f := range r.File {
			if strings.HasSuffix(f.Name, ".opf") {
				opfPath = f.Name
				break
			}
		}
		require.NotEmpty(t, opfPath, "should find OPF file")

		coverPath := findEpubCover(r, nil, opfPath)
		t.Logf("Found cover path: %q", coverPath)
	})
}
