package httputil

import (
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	StatusCode  int
	wroteHeader bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

func NewSafeResponseWriter(w http.ResponseWriter) *ResponseWriter {
	if existing, ok := w.(*ResponseWriter); ok {
		return existing
	} else {
		return NewResponseWriter(w)
	}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.StatusCode = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
