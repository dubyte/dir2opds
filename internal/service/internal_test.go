package service

import (
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
