package httputil

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.themarshians.com/dinglebit/log"
)

func TestAccessControlHandler(t *testing.T) {
	h := AccessControlHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello, world")
	}), "*")

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, r)

	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin was set to '%v', expected '*'",
			rr.Header().Get("Access-Control-Allow-Origin"))
	}
}

var testHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	_ = r.Body.Close()
	w.WriteHeader(400)
	h := w.Header()
	_, _ = io.WriteString(w, h.Get("Host"))
	_, _ = io.WriteString(w, " - echo: ")
	_, _ = w.Write(body)
})

func TestLogging(t *testing.T) {

	// Setup request.
	data := bytes.NewBuffer([]byte("hello, server"))
	r, err := http.NewRequest("POST", "/", data)
	if err != nil {
		t.Fatalf("failed making request: %v", err)
	}
	rr := httptest.NewRecorder()

	// create the logger.
	buf := &bytes.Buffer{}
	l := log.New(buf)

	h := LogHandler(testHandler, l)

	// Make the request and verify the values.
	h.ServeHTTP(rr, r)
	line := buf.String()
	parts := strings.Split(line, " ")
	t.Logf(line)
	if len(parts) != 10 {
		t.Fatalf("didn't get a full log line (10 parts): %v %v", len(parts), parts)
	}
	if parts[4] != "POST" {
		t.Errorf("Method was not POST: %v", parts[4])
	}
	if parts[5] != "/" {
		t.Errorf("URL was not '/': %v", parts[5])
	}
	if parts[6] != "400" {
		t.Errorf("code was not 400: %v", parts[6])
	}
	if parts[8] != "13" {
		t.Errorf("request size was not 13: %v", parts[8])
	}
	if parts[9] != "22\"\n" {
		t.Errorf("response was not 22: %v", parts[9])
	}
	if rr.Body.String() != " - echo: hello, server" {
		t.Errorf("response body was not ' - echo: hello, server': %v",
			rr.Body.String())
	}
}
