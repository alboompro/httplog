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

// Middleware defines function requireds to send log
type Middleware interface {
	// Send function to write log result
	Send(l Log) error

	// NewLog receives a http request to create log
	NewLog(r *http.Request, sw *statusWriter) Log
}

// MiddlewareTimed interface to add duration into the log
type MiddlewareTimed interface {
	SetDuration(int64)
}

// MiddlewareRequestID interface to add request ID into the log
type MiddlewareRequestID interface {
	SetRequestID(string)
}

// MiddlewareParams interface to add request parameters into the log
type MiddlewareParams interface {
	SetParams(interface{})
}

// MiddlewareNamed interface to add name of route into the log
type MiddlewareNamed interface {
	SetName(name string)
}

// MiddlewareChecker check if the resquest should be send do logger
type MiddlewareChecker interface {
	Check(interface{}) bool
}

var middlewares map[string]Middleware

// RegisterMiddleware registry a new log middleware in store
func RegisterMiddleware(name string, m Middleware) {
	middlewares[name] = m
}
