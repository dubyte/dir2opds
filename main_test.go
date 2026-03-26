package main

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	// setup
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	oldLogger := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(oldLogger)

	f := func(http.ResponseWriter, *http.Request) error {
		return errors.New("scary error")
	}
	h := errorHandler(f)
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// act
	h(res, req)

	// assert
	assert.Contains(t, buf.String(), `"level":"ERROR"`)
	assert.Contains(t, buf.String(), `"msg":"request error"`)
	assert.Contains(t, buf.String(), `"uri":"/"`)
	assert.Contains(t, buf.String(), `"error":"scary error"`)
}

func Test_absoluteCanonicalPath(t *testing.T) {
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
		{name: "dir relative path", args: args{aPath: "./opds"}, want: filepath.Join(wd, "opds"), wantErr: false},
		{name: "dir not exists", args: args{aPath: "books"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := absoluteCanonicalPath(tt.args.aPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("absoluteCanonicalPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("absoluteCanonicalPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
