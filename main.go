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
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dubyte/dir2opds/internal/service"
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

func main() {

	flag.Parse()

	if !*debug {
		log.SetOutput(ioutil.Discard)
	}

	fmt.Println(startValues())

	s := service.OPDS{DirRoot: *dirRoot, Author: *author, AuthorEmail: *authorEmail, AuthorURI: *authorURI}

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
