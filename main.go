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
	"encoding/xml"
	"encoding/base64"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/tools/blog/atom"
)

var dirRoot string

func init() {
	mime.AddExtensionType(".mobi", "application/x-mobipocket-ebook")
	mime.AddExtensionType(".epub", "application/epub+zip")
	mime.AddExtensionType(".fb2", "txt/xml")
	http.HandleFunc("/", errorHandler(catalogRoot))
}

func main() {
	portPtr := flag.String("port", "8080", "The server will listen in this port")
	dirPtr := flag.String("dir", "./books", "A directory with books")
	flag.Parse()

	dirRoot = *dirPtr
	log.Fatal(http.ListenAndServe(":"+*portPtr, nil))
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

func catalogRoot(w http.ResponseWriter, req *http.Request) error {
	dirPath := filepath.Join(dirRoot, req.URL.Path)
	fi, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return catalogFeed(w, req, dirPath)
	}
	return writeFileTo(w, dirPath)
}

func catalogFeed(w io.Writer, r *http.Request, dirPath string) error {
	fis, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	feed := &atom.Feed{Title: "OPDS Catalog: " + r.URL.Path}
	if len(fis) < 1 {
		return writeFeedTo(w, feed)
	}

	err = FeedEntries(feed, fis, r)
	if err != nil {
		return err
	}

	return writeFeedTo(w, feed)
}

func FeedEntries(f *atom.Feed, fis []os.FileInfo, r *http.Request) error {
	for _, fi := range fis {
		e := &atom.Entry{Title: fi.Name()}
		encoded := base64.StdEncoding.EncodeToString([]byte(path.Join(r.URL.Path, fi.Name())))
		e.ID = encoded
		l := atom.Link{Title: fi.Name(), Href: path.Join(r.URL.EscapedPath(), fi.Name())}
		if !fi.IsDir() {
			l.Rel = "http://opds-spec.org/acquisition"
		}
		lType, err := getLinkType(path.Join(dirRoot, r.URL.Path, fi.Name()))
		if err != nil {
			return err
		}
		l.Type = lType
		e.Link = append(e.Link, l)
		f.Entry = append(f.Entry, e)
	}
	return nil
}

func writeFeedTo(w io.Writer, feed *atom.Feed) error {
	io.WriteString(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	if err := enc.Encode(feed); err != nil {
		return err
	}
	return nil
}

func writeFileTo(w io.Writer, filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	if err != nil {
		f.Close()
		return err
	}
	f.Close()
	return nil
}

func getLinkType(lPath string) (string, error) {
	fi, err := os.Stat(lPath)
	if err != nil {
		return "", err
	}
	if !fi.IsDir() {
		return mime.TypeByExtension(filepath.Ext(lPath)), nil
	}
	files, err := ioutil.ReadDir(lPath)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if !file.IsDir() {
			return "application/atom+xml;profile=opds-catalog;kind=acquisition", nil
		}
	}
	return "application/atom+xml;profile=opds-catalog;kind=navigation", nil
}
