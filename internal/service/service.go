//package service provides a http handler that reads the path in the request.url and returns
// an xml document that follows the OPDS 1.1 standard
// https://specs.opds.io/opds-1.1.html
package service

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
}

type OPDS struct {
	DirRoot     string
	Author      string
	AuthorEmail string
	AuthorURI   string
}

var TimeNowFunc = timeNow

func (s OPDS) Handler(w http.ResponseWriter, req *http.Request) error {
	fPath := filepath.Join(s.DirRoot, req.URL.Path)

	log.Printf("fPath:'%s'", fPath)

	fi, err := os.Stat(fPath)
	if err != nil {
		return err
	}

	if isFile(fi) {
		http.ServeFile(w, req, fPath)
		return nil
	}

	content, err := s.getContent(req, fPath)
	if err != nil {
		return err
	}

	content = append([]byte(xml.Header), content...)
	http.ServeContent(w, req, "feed.xml", TimeNowFunc(), bytes.NewReader(content))
	return nil
}

func (s OPDS) getContent(req *http.Request, dirpath string) (result []byte, err error) {
	feed := s.makeFeed(dirpath, req)
	if getPathType(dirpath) == pathTypeDirOfFiles {
		acFeed := &opds.AcquisitionFeed{&feed, "http://purl.org/dc/terms/", "http://opds-spec.org/2010/catalog"}
		result, err = xml.MarshalIndent(acFeed, "  ", "    ")
	} else {
		result, err = xml.MarshalIndent(feed, "  ", "    ")
	}
	return
}

const navigationType = "application/atom+xml;profile=opds-catalog;kind=navigation"

func (s OPDS) makeFeed(dirpath string, req *http.Request) atom.Feed {
	feedBuilder := opds.FeedBuilder.
		ID(req.URL.Path).
		Title("Catalog in " + req.URL.Path).
		Author(opds.AuthorBuilder.Name(s.Author).Email(s.AuthorEmail).URI(s.AuthorURI).Build()).
		Updated(TimeNowFunc()).
		AddLink(opds.LinkBuilder.Rel("start").Href("/").Type(navigationType).Build())

	fis, _ := ioutil.ReadDir(dirpath)
	for _, fi := range fis {
		pathType := getPathType(filepath.Join(dirpath, fi.Name()))
		feedBuilder = feedBuilder.
			AddEntry(opds.EntryBuilder.
				ID(req.URL.Path + fi.Name()).
				Title(fi.Name()).
				Updated(TimeNowFunc()).
				Published(TimeNowFunc()).
				AddLink(opds.LinkBuilder.
					Rel(getRel(fi.Name(), pathType)).
					Title(fi.Name()).
					Href(getHref(req, fi.Name())).
					Type(getType(fi.Name(), pathType)).
					Build()).
				Build())
	}
	return feedBuilder.Build()
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
	if pathType == pathTypeFile {
		return mime.TypeByExtension(filepath.Ext(name))
	}
	return "application/atom+xml;profile=opds-catalog;kind=acquisition"
}

func getHref(req *http.Request, name string) string {
	return filepath.Join(req.URL.RequestURI(), url.PathEscape(name))
}

const (
	pathTypeFile = iota
	pathTypeDirOfDirs
	pathTypeDirOfFiles
)

func getPathType(dirpath string) int {
	fi, _ := os.Stat(dirpath)
	if isFile(fi) {
		return pathTypeFile
	}

	fis, _ := ioutil.ReadDir(dirpath)
	for _, fi := range fis {
		if isFile(fi) {
			return pathTypeDirOfFiles
		}
	}
	// Directory of directories
	return pathTypeDirOfDirs
}

func isFile(fi os.FileInfo) bool {
	return !fi.IsDir()
}

func timeNow() time.Time {
	return time.Now()
}
