/*
 * HTTP Log
 *
 * Analysis http request and sync with middleware printer or sync logs
 *
 * API version: 1.0.0
 * Contact: welington@alboompro.com
 */
package httplog

import (
	"log"
	"net/http"
	"os"
)

// OutputConsole Default output logger writting in Stdout
type OutputConsole struct {
	syslog *log.Logger
}

// Send interface of `Middleware`
func (oc *OutputConsole) Send(l Log) error {
	oc.syslog.Printf(l.ToString())
	return nil
}

// NewLog receives a http request to create log
func (oc *OutputConsole) NewLog(r *http.Request, sw *statusWriter) Log {
	return NewLog(r, sw)
}

// NewOutputConsole creates a default logger
func NewOutputConsole() Middleware {
	a := &OutputConsole{
		syslog: log.New(os.Stderr, "", log.LstdFlags),
	}
	return a
}
