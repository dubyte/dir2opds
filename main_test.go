package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartValues(t *testing.T) {
	// pre-setup
	oldHost, oldPort := *host, *port
	defer func() {
		*host = oldHost
		*port = oldPort
	}()

	// setup
	*host = "wow.com"
	*port = "42"

	// act
	res := startValues()

	// assert
	assert.Equal(t, "listening in: wow.com:42", res)

}

func TestErrorHandler(t *testing.T) {
	// pre-setup
	stdOutput := log.Writer()
	defer func() {
		log.SetOutput(stdOutput)
	}()

	// setup
	var buf bytes.Buffer
	log.SetOutput(&buf) // to record what is logged

	f := func(http.ResponseWriter, *http.Request) error {
		return errors.New("scary error")
	}
	h := errorHandler(f)
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// act
	h(res, req)

	// assert
	assert.Contains(t, buf.String(), `handling "/": scary error`)
}

func Test_absoluteCannnonicalPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("not possible to get current dir")
	}
	type args struct {
		aPath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "dir relative path", args: args{aPath: "./opds"}, want: path.Join(wd, "opds"), wantErr: false},
		{name: "dir not exists", args: args{aPath: "books"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := absoluteCannnonicalPath(tt.args.aPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("absoluteCannnonicalPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("absoluteCannnonicalPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
