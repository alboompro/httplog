/*
 * HTTP Log
 *
 * Analysis http request and sync with middleware printer or sync logs
 *
 * API version: 1.0.0
 * Contact: welington@alboompro.com
 */
package httplog

import "net/http"

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
