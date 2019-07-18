// Package httputil contains utility functions for the standard library http
// package.
package httputil

import (
	"io"
	"net/http"
	"time"

	"git.themarshians.com/dinglebit/log"
)

// AccessControlHandler wraps the given handler with an
// 'Access-Control-Allow-Origin' header allowing the given access.
func AccessControlHandler(h http.Handler, access string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", access)
		h.ServeHTTP(w, r)
	})
}

// ResponseWriterStats stores information about data being written to an
// http.ResponseWriter. The information is not locked with any syncing, so you
// should only check the results when you are done writing.
type ResponseWriterStats struct {
	w            http.ResponseWriter
	ResponseCode int
	Total        int
}

// NewResponseWriterStats creates a new ResponseWriterStats that wraps the given
// http.ResponseWriter.
func NewResponseWriterStats(w http.ResponseWriter) *ResponseWriterStats {
	return &ResponseWriterStats{w: w, ResponseCode: 200}
}

// Header implements the ResponseWriter interface.
func (c *ResponseWriterStats) Header() http.Header {
	return c.w.Header()
}

// WriteHeader implements the ResponseWriter interface.
func (c *ResponseWriterStats) WriteHeader(h int) {
	c.ResponseCode = h
	c.w.WriteHeader(h)
}

// Write implements the ResponseWriter interface.
func (c *ResponseWriterStats) Write(data []byte) (int, error) {
	c.Total += len(data)
	return c.w.Write(data)
}

// RequestBodyStats stores information about data being read from an
// http.Request.Body. The information is not locked with any syncing, so you
// should only check the results when you are done reading.
type RequestBodyStats struct {
	r     io.ReadCloser
	Total int
}

// NewRequestBodyStats creates a new RequestBodyStats that wraps the given
// io.ReadCloser.
func NewRequestBodyStats(r io.ReadCloser) *RequestBodyStats {
	return &RequestBodyStats{r: r}
}

// Close implements the io.ReadCloser interface.
func (r *RequestBodyStats) Close() error {
	return r.r.Close()
}

// Read implements the io.ReadCloser interface.
func (r *RequestBodyStats) Read(data []byte) (int, error) {
	n, err := r.r.Read(data)
	r.Total += n
	return n, err
}

// LogHandler creates a handler that logs the time, bytes read, bytes written,
// and more to the given logger.
func LogHandler(h http.Handler, l log.Interface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		out := NewResponseWriterStats(w)
		in := NewRequestBodyStats(r.Body)
		r.Body = in
		h.ServeHTTP(out, r)
		diff := time.Now().Sub(start)
		l.Infof("%v %v %v %v %v %v %v %v", r.RemoteAddr, r.Proto, r.Method, r.URL,
			out.ResponseCode, diff, in.Total, out.Total)
	})
}
