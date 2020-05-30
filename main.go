/*
  Copyright (C) 2017 Sinuhé Téllez Rivera

  dir2opds is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  dir2opds is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with dir2opds.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dubyte/dir2opds/opds"
	"golang.org/x/tools/blog/atom"
)

var (
	port        = flag.String("port", "8080", "The server will listen in this port")
	host        = flag.String("host", "0.0.0.0", "The server will listen in this host")
	dirRoot     = flag.String("dir", "./books", "A directory with books")
	author      = flag.String("author", "", "The server Feed author")
	authorURI   = flag.String("uri", "", "The feed's author uri")
	authorEmail = flag.String("email", "", "The feed's author email")
	debug       = flag.Bool("debug", false, "If it is set it will log the requests")
)

type acquisitionFeed struct {
	*atom.Feed
	Dc   string `xml:"xmlns:dc,attr"`
	Opds string `xml:"xmlns:opds,attr"`
}

func init() {
	_ = mime.AddExtensionType(".mobi", "application/x-mobipocket-ebook")
	_ = mime.AddExtensionType(".epub", "application/epub+zip")
	_ = mime.AddExtensionType(".cbz", "application/x-cbz")
	_ = mime.AddExtensionType(".cbr", "application/x-cbr")
	_ = mime.AddExtensionType(".fb2", "text/fb2+xml")
}

func main() {

	flag.Parse()

	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	fmt.Println(startValues())

	http.HandleFunc("/", errorHandler(handler))

	log.Fatal(http.ListenAndServe(*host+":"+*port, nil))
}

func startValues() string {
	result := fmt.Sprintf("listening in: %s:%s", *host, *port)
	return result
}

func handler(w http.ResponseWriter, req *http.Request) error {
	fPath := path.Join(*dirRoot, req.URL.Path)

	log.Printf("fPath:'%s'", fPath)

	fi, err := os.Stat(fPath)
	if err != nil {
		return err
	}

	if isFile(fi) {
		http.ServeFile(w, req, fPath)
		return nil
	}

	content, err := getContent(req, fPath)
	if err != nil {
		return err
	}

	content = append([]byte(xml.Header), content...)
	http.ServeContent(w, req, "feed.xml", time.Now(), bytes.NewReader(content))
	return nil
}

func getContent(req *http.Request, dirpath string) (result []byte, err error) {
	feed := makeFeed(dirpath, req)
	if getPathType(dirpath) == pathTypeDirOfFiles {
		acFeed := &acquisitionFeed{&feed, "http://purl.org/dc/terms/", "http://opds-spec.org/2010/catalog"}
		result, err = xml.MarshalIndent(acFeed, "  ", "    ")
	} else {
		result, err = xml.MarshalIndent(feed, "  ", "    ")
	}
	return
}

const navigationType = "application/atom+xml;profile=opds-catalog;kind=navigation"

func makeFeed(dirpath string, req *http.Request) atom.Feed {
	feedBuilder := opds.FeedBuilder.
		ID(req.URL.Path).
		Title("Catalog in " + req.URL.Path).
		Author(opds.AuthorBuilder.Name(*author).Email(*authorEmail).URI(*authorURI).Build()).
		Updated(time.Now()).
		AddLink(opds.LinkBuilder.Rel("start").Href("/").Type(navigationType).Build())

	fis, _ := ioutil.ReadDir(dirpath)
	for _, fi := range fis {
		pathType := getPathType(path.Join(dirpath, fi.Name()))
		feedBuilder = feedBuilder.
			AddEntry(opds.EntryBuilder.
				ID(req.URL.Path + fi.Name()).
				Title(fi.Name()).
				Updated(time.Now()).
				Published(time.Now()).
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
	return path.Join(req.URL.RequestURI(), name)
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

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("handling %q: %v", r.RequestURI, err)
		}
	}
}
