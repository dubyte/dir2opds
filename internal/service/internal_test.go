package service

import (
	"net/url"
	"os"
	"path/filepath"
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
		title, author := extractEpubMetadata(path)
		// We don't know the exact content of the fixture but it should not crash
		// If the fixture is valid, we could assert more.
		t.Logf("EPUB Title: %q, Author: %q", title, author)
	})

	t.Run("Parse PDF value", func(t *testing.T) {
		line := "/Title (The Great Gatsby) /Author (F. Scott Fitzgerald)"
		assert.Equal(t, "The Great Gatsby", parsePdfValue(line, "/Title"))
		assert.Equal(t, "F. Scott Fitzgerald", parsePdfValue(line, "/Author"))
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

func TestPagination(t *testing.T) {
	s := OPDS{TrustedRoot: "testdata", HideCalibreFiles: true, HideDotFiles: true}

	t.Run("First page", func(t *testing.T) {
		catalog, err := s.Scan("testdata/mybook", "/mybook", 1)
		require.NoError(t, err)
		assert.Equal(t, 1, catalog.Page)
		assert.Equal(t, defaultPageSize, catalog.PageSize)
		assert.Equal(t, 5, catalog.Total)
	})

	t.Run("Page with small page size", func(t *testing.T) {
		s.PageSize = 2
		catalog, err := s.Scan("testdata/mybook", "/mybook", 1)
		require.NoError(t, err)
		assert.Equal(t, 1, catalog.Page)
		assert.Equal(t, 2, catalog.PageSize)
		assert.Equal(t, 5, catalog.Total)
		assert.Len(t, catalog.Entries, 2)
	})

	t.Run("Second page", func(t *testing.T) {
		s.PageSize = 2
		catalog, err := s.Scan("testdata/mybook", "/mybook", 2)
		require.NoError(t, err)
		assert.Equal(t, 2, catalog.Page)
		assert.Equal(t, 5, catalog.Total)
		assert.Len(t, catalog.Entries, 2)
	})

	t.Run("Last page with partial entries", func(t *testing.T) {
		s.PageSize = 2
		catalog, err := s.Scan("testdata/mybook", "/mybook", 3)
		require.NoError(t, err)
		assert.Equal(t, 3, catalog.Page)
		assert.Equal(t, 5, catalog.Total)
		assert.Len(t, catalog.Entries, 1)
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
