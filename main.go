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
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/dubyte/dir2opds/internal/service"
)

var (
	port    = flag.String("port", "8080", "The server will listen in this port")
	host    = flag.String("host", "0.0.0.0", "The server will listen in this host")
	dirRoot = flag.String("dir", "./books", "A directory with books")
	debug   = flag.Bool("debug", false, "If it is set it will log the requests")
	calibre = flag.Bool("calibre", false, "Hide files stored by calibre")
)

func main() {

	flag.Parse()

	if !*debug {
		log.SetOutput(io.Discard)
	}

	fmt.Println(startValues())
	var err error

	// Use the absoluteCannonical path of the dir parm as the trustedRoot.
	// helpfull avoid http trasversal. https://github.com/dubyte/dir2opds/issues/17
	*dirRoot, err = absoluteCannnonicalPath(*dirRoot)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%q will be used as your trusted root", *dirRoot)

	s := service.OPDS{TrustedRoot: *dirRoot, IsCalibreLibrary: *calibre}

	http.HandleFunc("/", errorHandler(s.Handler))

	log.Fatal(http.ListenAndServe(*host+":"+*port, nil))
}

func startValues() string {
	result := fmt.Sprintf("listening in: %s:%s", *host, *port)
	return result
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

// absoluteCannnonicalPath returns the cannonical path of the absolute path that was passed
func absoluteCannnonicalPath(aPath string) (string, error) {
	// get absolute path
	aPath, err := filepath.Abs(aPath)
	if err != nil {
		return "", fmt.Errorf("get absolute path %s: %w", aPath, err)
	}

	// get cannonical path
	aPath, err = filepath.EvalSymlinks(aPath)
	if err != nil {
		return "", fmt.Errorf("get connonical path from absolute path %s: %w", aPath, err)
	}

	return aPath, nil
}
