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
	"bufio"
	"encoding/xml"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/tools/blog/atom"
)

type AcquisitionFeed struct {
	*atom.Feed
	Dc   string `xml:"xmlns:dc,attr"`
	Opds string `xml:"xmlns:opds,attr"`
}

type CatalogFeed atom.Feed

const acquisitionType = "application/atom+xml;profile=opds-catalog;kind=acquisition"
const navegationType = "application/atom+xml;profile=opds-catalog;kind=navigation"

var (
	port,
	dirRoot,
	author,
	authorUri,
	authorEmail string
	updated atom.TimeStr
)

func init() {
	mime.AddExtensionType(".mobi", "application/x-mobipocket-ebook")
	mime.AddExtensionType(".epub", "application/epub+zip")
	mime.AddExtensionType(".fb2", "txt/xml")

	flag.StringVar(&port, "port", "8080", "The server will listen in this port")
	flag.StringVar(&dirRoot, "dir", "./books", "A directory with books")
	flag.StringVar(&author, "author", "", "The author of the feed")
	flag.StringVar(&authorUri, "uri", "", "The author uri")
	flag.StringVar(&authorEmail, "email", "", "The author email")
	flag.Parse()
	updated = atom.Time(time.Now())
}

func main() {
	http.HandleFunc("/", errorHandler(func(w http.ResponseWriter, req *http.Request) error {
		dirPath := filepath.Join(dirRoot, req.URL.Path)
		fi, err := os.Stat(dirPath)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			w.Write([]byte(xml.Header))
			return writeFeedTo(w, req.URL)
		}

		return writeFileTo(w, dirPath)
	}))

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func writeFeedTo(w io.Writer, u *url.URL) error {
	isAcquisition, err := isAcquisitionFeed(filepath.Join(dirRoot, u.Path))
	if err != nil {
		return err
	}
	if isAcquisition {
		return writeAcquisitionFeed(w, u)
	}
	return writeCatalogFeed(w, u)
}

func isAcquisitionFeed(p string) (bool, error) {
	fis, err := ioutil.ReadDir(p)
	if err != nil {
		return false, err
	}
	for _, fi := range fis {
		if !fi.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

func writeCatalogFeed(w io.Writer, u *url.URL) error {
	feed := &CatalogFeed{ID: u.Path, Title: "Catalog feed in " + u.Path}
	feed.Author = &atom.Person{Name: author, Email: authorEmail, URI: authorUri}
	feed.Updated = updated
	feed.Link = []atom.Link{{
		Rel:  "start",
		Href: "/",
		Type: navegationType,
	}}

	abs_path := filepath.Join(dirRoot, u.Path)
	fis, err := ioutil.ReadDir(abs_path)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		link := atom.Link{
			Rel:   "subsection",
			Title: fi.Name(),
			Href:  filepath.Join(u.EscapedPath(), url.PathEscape(fi.Name())),
			Type:  acquisitionType,
		}
		entry := &atom.Entry{
			ID:        filepath.Join(u.Path, fi.Name()),
			Title:     fi.Name(),
			Updated:   updated,
			Published: updated,
			Link:      []atom.Link{link},
		}
		feed.Entry = append(feed.Entry, entry)

	}

	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	enc.Encode(feed)
	return nil
}

func writeAcquisitionFeed(w io.Writer, u *url.URL) error {
	f := &atom.Feed{}
	feed := &AcquisitionFeed{f, "http://purl.org/dc/terms/", "http://opds-spec.org/2010/catalog"}
	feed.ID = u.Path
	feed.Updated = updated
	feed.Title = filepath.Base(u.Path)
	feed.Author = &atom.Person{Name: author, Email: authorEmail, URI: authorUri}
	feed.Link = []atom.Link{{
		Rel:  "start",
		Href: "/",
		Type: navegationType,
	}}

	abs_path := filepath.Join(dirRoot, u.Path)
	fis, err := ioutil.ReadDir(abs_path)
	if err != nil {
		return err
	}
	entry := &atom.Entry{
		ID:        u.Path,
		Title:     filepath.Base(u.Path),
		Updated:   updated,
		Published: updated,
	}
	for _, fi := range fis {
		ext := filepath.Ext(fi.Name())
		mime_type := mime.TypeByExtension(ext)
		var rel string
		if rel = "http://opds-spec.org/acquisition"; ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gift" {
			rel = "http://opds-spec.org/image/thumbnail"
		}
		link := atom.Link{
			Title: fi.Name(),
			Rel:  rel,
			Type: mime_type,
			Href: filepath.Join(u.EscapedPath(), url.PathEscape(fi.Name())),
		}
		entry.Link = append(entry.Link, link)
	}
	feed.Entry = append(feed.Entry, entry)

	enc := xml.NewEncoder(w)
	enc.Indent("  ", "    ")
	enc.Encode(feed)
	return nil
}

func writeFileTo(w io.Writer, p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	r.WriteTo(w)
	return nil
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
