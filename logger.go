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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/google/uuid"
	"github.com/tomasen/realip"
)

// ContextName is a context typed
type ContextName string

var (
	syslog            *log.Logger
	defaultMiddleware Middleware = NewOutputConsole()

	// ResourceRequestID use request id in log
	ResourceRequestID = 2

	// ResourceDuration use request id in log
	ResourceDuration = 4

	// ResourceName use name in log
	ResourceName = 8

	// ResourceParams use params in log
	ResourceParams = 16

	// Resources defines what resources should be use
	Resources = ResourceRequestID | ResourceDuration | ResourceName | ResourceParams

	// ContextRequestID Context name of header request id
	ContextRequestID = ContextName("request_id")

	// ContextRouteName Context name of route name
	ContextRouteName = ContextName("route_name")

	// ContextParams Context name of parameters
	ContextParams = ContextName("params")
)

// Log is an interface of Log object
type Log interface {
	// ToString converts the log into string
	ToString() string

	// ToJSON converts the log into json format
	ToJSON() string
}

// LogRequest structure of log to print in terminal or send to cloudwatch like json format
type LogRequest struct {
	Duration      int64       `json:"duration"`
	Method        string      `json:"method"`
	Params        interface{} `json:"params,omitempty"`
	ContentLength int         `json:"content_length"`
	Path          string      `json:"path"`
	QueryString   string      `json:"query_string,omitempty"`
	RemoteIP      string      `json:"remote_ip,omitempty"`
	RouteName     string      `json:"route_name,omitempty"`
	RequestID     string      `json:"request_id"`
	Status        int         `json:"status,omitempty"`
	UserAgent     string      `json:"user_agent"`
	Dump          interface{} `json:"dump,omitempty"`
	Error         interface{} `json:"error,omitmepty"`
}

// ToString converts structure to simples string like Apache log
func (l *LogRequest) ToString() string {
	return fmt.Sprintf(
		"[%s] [%s] %s - %d %d %d - %s %s \"%s\" \"%s\" - \"%s\"",
		l.RequestID,
		l.RouteName,
		l.RemoteIP,
		l.Status,
		l.Duration,
		l.ContentLength,
		l.Method,
		l.Path,
		l.QueryString,
		l.UserAgent,
		l.Error,
	)
}

// ToJSON converts the structure to JSON string
func (l *LogRequest) ToJSON() string {
	content, err := json.Marshal(l)

	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}

	return fmt.Sprintf("%s", content)
}

// SetDuration interface to add duration time in log
func (l *LogRequest) SetDuration(duration int64) {
	l.Duration = duration
}

// SetRequestID interface to add request ID in log
func (l *LogRequest) SetRequestID(id string) {
	l.RequestID = id
}

// SetName interface to add name in log
func (l *LogRequest) SetName(name string) {
	l.RouteName = name
}

// SetParams interface to add name in log
func (l *LogRequest) SetParams(params interface{}) {
	l.Params = params
}

// SetError interface to add error message in log
func (l *LogRequest) SetError(err interface{}) {
	l.Error = err
}

// NewLog Default and basic log message
func NewLog(r *http.Request, sw *statusWriter) (logMessage *LogRequest) {
	logMessage = &LogRequest{
		Method:      r.Method,
		Path:        r.URL.Path,
		QueryString: r.URL.RawQuery,
		RemoteIP:    realip.FromRequest(r),
		UserAgent:   r.UserAgent(),
	}

	if sw != nil {
		logMessage.Status = sw.status
		logMessage.ContentLength = sw.length
	}

	dump, err := httputil.DumpRequest(r, true)
	if err == nil {
		logMessage.Dump = dump
	}

	logMessage.Params = r.Context().Value("params")

	return
}

// With default method to inject middleware in http.Router
func With(inner http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(handler(inner))
}

// WithNamed introducing of name if log
func WithNamed(name string, fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ContextRouteName, name)
		r = r.WithContext(ctx)
		With(fn).ServeHTTP(w, r.WithContext(ctx))
	})
}

// Inject default method to inject middleware in http.Router
func Inject(inner http.Handler) http.Handler {
	return http.HandlerFunc(handler(inner))
}

func handler(inner http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}

		if (Resources & ResourceRequestID) != 0 {
			id := uuid.New().String()
			ctx := context.WithValue(r.Context(), ContextRequestID, id)
			r = r.WithContext(ctx)
		}

		defer func() {
			err := recover()
			if err != nil {
				fmt.Println(err) // May be log this error? Send to sentry?

				sw.WriteHeader(http.StatusInternalServerError)
				sw.Write([]byte("There was an internal server error"))

				l := prepareLog(r, sw, start)

				// Add Error message to log
				if li, ok := l.(MiddlewareError); ok {
					li.SetError(err)
				}

				go defaultMiddleware.Send(l)
			}

		}()

		inner.ServeHTTP(sw, r)

		l := prepareLog(r, sw, start)

		if li, ok := l.(MiddlewareChecker); ok {
			if li.Check(l) {
				go defaultMiddleware.Send(l)
			}
		} else {
			go defaultMiddleware.Send(l)
		}

	}
}

func prepareLog(r *http.Request, sw *statusWriter, start time.Time) Log {
	l := defaultMiddleware.NewLog(r, sw)

	// Add duration if timed log
	if (Resources & ResourceDuration) != 0 {
		if li, ok := l.(MiddlewareTimed); ok {
			li.SetDuration(time.Since(start).Milliseconds())
		}
	}

	// Add request id if exists log
	if (Resources & ResourceRequestID) != 0 {
		if li, ok := l.(MiddlewareRequestID); ok {
			li.SetRequestID(r.Context().Value(ContextRequestID).(string))
		}
	}

	// Add name if named log
	if (Resources & ResourceName) != 0 {
		if li, ok := l.(MiddlewareNamed); ok {
			if name := r.Context().Value(ContextRouteName); name != nil {
				li.SetName(name.(string))
			}

		}
	}

	// Add parameters if exists
	if (Resources & ResourceParams) != 0 {
		if li, ok := l.(MiddlewareParams); ok {
			li.SetParams(r.Context().Value(ContextParams))
		}
	}

	return l
}

// Use defines which middleware should be use to send logs
func Use(name string) {
	a := middlewares[name]
	if a != nil {
		defaultMiddleware = a
	}
}

// Default returns default middleware logger
func Default() Middleware {
	return defaultMiddleware
}
