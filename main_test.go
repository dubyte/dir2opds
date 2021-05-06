package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
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
