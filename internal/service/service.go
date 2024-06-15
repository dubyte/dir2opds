// package service provides a http handler that reads the path in the request.url and returns
// an xml document that follows the OPDS 1.1 standard
// https://specs.opds.io/opds-1.1.html
package service

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dubyte/dir2opds/opds"
	"golang.org/x/tools/blog/atom"
)

func init() {
	_ = mime.AddExtensionType(".mobi", "application/x-mobipocket-ebook")
	_ = mime.AddExtensionType(".epub", "application/epub+zip")
	_ = mime.AddExtensionType(".cbz", "application/x-cbz")
	_ = mime.AddExtensionType(".cbr", "application/x-cbr")
	_ = mime.AddExtensionType(".fb2", "text/fb2+xml")
	_ = mime.AddExtensionType(".pdf", "application/pdf")
}

const (
	pathTypeFile = iota
	pathTypeDirOfDirs
	pathTypeDirOfFiles
)

const (
	ignoreFile       = true
	includeFile      = false
	currentDirectory = "."
	parentDirectory  = ".."
	hiddenFilePrefix = "."
)

type OPDS struct {
	TrustedRoot      string
	HideCalibreFiles bool
	HideDotFiles     bool
	NoCache          bool
}

type IsDirer interface {
	IsDir() bool
}

func isFile(e IsDirer) bool {
	return !e.IsDir()
}

const navigationType = "application/atom+xml;profile=opds-catalog;kind=navigation"

var TimeNow = timeNowFunc()

// Handler serve the content of a book file or
// returns an Acquisition Feed when the entries are documents or
// returns an Navegation Feed when the entries are other folders
func (s OPDS) Handler(w http.ResponseWriter, req *http.Request) error {
	var err error
	urlPath, err := url.PathUnescape(req.URL.Path)
	if err != nil {
		log.Printf("error while serving '%s': %s", req.URL.Path, err)
		return err
	}

	fPath := filepath.Join(s.TrustedRoot, urlPath)

	// verifyPath avoid the http transversal by checking the path is under DirRoot
	_, err = verifyPath(fPath, s.TrustedRoot)
	if err != nil {
		log.Printf("fPath %q err: %s", fPath, err)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	log.Printf("urlPath:'%s'", urlPath)

	if _, err := os.Stat(fPath); err != nil {
		log.Printf("fPath err: %s", err)
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	log.Printf("fPath:'%s'", fPath)

	// it's a file just serve the file
	if getPathType(fPath) == pathTypeFile {
		http.ServeFile(w, req, fPath)
		return nil
	}

	if s.NoCache {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Expires", "0")
	}

	navFeed := s.makeFeed(fPath, req)

	var content []byte
	// it is an acquisition feed
	if getPathType(fPath) == pathTypeDirOfFiles {
		acFeed := &opds.AcquisitionFeed{Feed: &navFeed, Dc: "http://purl.org/dc/terms/", Opds: "http://opds-spec.org/2010/catalog"}
		content, err = xml.MarshalIndent(acFeed, "  ", "    ")
		w.Header().Add("Content-Type", "application/atom+xml;profile=opds-catalog;kind=acquisition")
	} else { // it is a navegation feed
		content, err = xml.MarshalIndent(navFeed, "  ", "    ")
		w.Header().Add("Content-Type", "application/atom+xml;profile=opds-catalog;kind=navigation")
	}
	if err != nil {
		log.Printf("error while serving '%s': %s", fPath, err)
		return err
	}

	content = append([]byte(xml.Header), content...)
	http.ServeContent(w, req, "feed.xml", TimeNow(), bytes.NewReader(content))

	return nil
}

func (s OPDS) makeFeed(fpath string, req *http.Request) atom.Feed {
	feedBuilder := opds.FeedBuilder.
		ID(req.URL.Path).
		Title("Catalog in " + req.URL.Path).
		Updated(TimeNow()).
		AddLink(opds.LinkBuilder.Rel("start").Href("/").Type(navigationType).Build())

	dirEntries, _ := os.ReadDir(fpath)
	for _, entry := range dirEntries {

		if fileShouldBeIgnored(entry.Name(), s.HideCalibreFiles, s.HideDotFiles) {
			continue
		}

		pathType := getPathType(filepath.Join(fpath, entry.Name()))
		feedBuilder = feedBuilder.
			AddEntry(opds.EntryBuilder.
				ID(req.URL.Path + entry.Name()).
				Title(entry.Name()).
				AddLink(opds.LinkBuilder.
					Rel(getRel(entry.Name(), pathType)).
					Title(entry.Name()).
					Href(filepath.Join(req.URL.RequestURI(), url.PathEscape(entry.Name()))).
					Type(getType(entry.Name(), pathType)).
					Build()).
				Build())
	}
	return feedBuilder.Build()
}

func fileShouldBeIgnored(filename string, hideCalibreFiles, hideDotFiles bool) bool {
	// not ignore those directories
	if filename == currentDirectory || filename == parentDirectory {
		return includeFile
	}

	if hideDotFiles && strings.HasPrefix(filename, hiddenFilePrefix) {
		return ignoreFile
	}

	if hideCalibreFiles &&
		(strings.Contains(filename, ".opf") ||
			strings.Contains(filename, "cover.") ||
			strings.Contains(filename, "metadata.db") ||
			strings.Contains(filename, "metadata_db_prefs_backup.json") ||
			strings.Contains(filename, ".caltrash") ||
			strings.Contains(filename, ".calnotes")) {
		return ignoreFile
	}

	return false
}

func getRel(name string, pathType int) string {
	if pathType == pathTypeDirOfFiles || pathType == pathTypeDirOfDirs {
		return "subsection"
	}

	ext := filepath.Ext(name)
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
		return "http://opds-spec.org/image/thumbnail"
	}

	// mobi, epub, etc
	return "http://opds-spec.org/acquisition"
}

func getType(name string, pathType int) string {
	switch pathType {
	case pathTypeFile:
		return mime.TypeByExtension(filepath.Ext(name))
	case pathTypeDirOfFiles:
		return "application/atom+xml;profile=opds-catalog;kind=acquisition"
	case pathTypeDirOfDirs:
		return "application/atom+xml;profile=opds-catalog;kind=navigation"
	default:
		return mime.TypeByExtension("xml")
	}
}

func getPathType(dirpath string) int {
	fi, err := os.Stat(dirpath)
	if err != nil {
		log.Printf("getPathType os.Stat err: %s", err)
	}

	if isFile(fi) {
		return pathTypeFile
	}

	dirEntries, err := os.ReadDir(dirpath)
	if err != nil {
		log.Printf("getPathType: readDir err: %s", err)
	}

	for _, entry := range dirEntries {
		if isFile(entry) {
			return pathTypeDirOfFiles
		}
	}
	// Directory of directories
	return pathTypeDirOfDirs
}

func timeNowFunc() func() time.Time {
	t := time.Now()
	return func() time.Time { return t }
}

// verify path use a trustedRoot to avoid http transversal
// from https://www.stackhawk.com/blog/golang-path-traversal-guide-examples-and-prevention/
func verifyPath(path, trustedRoot string) (string, error) {
	// clean is already used upstream but leaving this
	// to keep the functionality of the function as close as possible to the blog.
	c := filepath.Clean(path)

	// get the canonical path
	r, err := filepath.EvalSymlinks(c)
	if err != nil {
		fmt.Println("Error " + err.Error())
		return c, errors.New("unsafe or invalid path specified")
	}

	if !inTrustedRoot(r, trustedRoot) {
		return r, errors.New("unsafe or invalid path specified")
	}

	return r, nil
}

func inTrustedRoot(path string, trustedRoot string) bool {
	return strings.HasPrefix(path, trustedRoot)
}
