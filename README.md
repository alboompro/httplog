# HTTP Log

Simples extensible http middleware to print or send Http Request Log. Has support to middleware sync like Cloudwatch, Sumo Logic, LogStach etc.

## Install

Installing with a simple terminal command.

```bash
go get -u github.com/openalboompro/httplog
```

## Using


**net/http** example
```go
func main() {
    http.Handle("/foo", httplog.With(fooHandler))
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}
```

**Mux** example using `github.com/gorilla/mux`
```go
func main() {
    r := mux.NewRouter()

    r.Use(httplog.Inject)
    r.HandleFunc("/", handler)
    // OR
    r.HandleFunc("/another", httplog.With(handler))

    log.Fatal(http.ListenAndServe(":8080", r))
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}
```


## Middleware

With middlewares you can send your logs to anywhere. And have a simple structure to create a new middleware, see bellow:

```go
type MyMiddleware struct {
    /* store your credentials to sync log */
    URL      string
    Username string
    Password string
}

// Send function executed when finish request
// param +l+ is an interface  
type (m *MyMiddleware) Send(l *httplog.Log) error {
    // Execute your request to send log
    return nil
}

// NewLog receives a http request to create log
type (m *MyMiddleware) NewLog(r *http.Request, sw *statusWriter) httplog.Log {
    // Execute your request to send log
    return httplog.NewLog(r, sw)
}

// Use Register function to add your custom middleware in httplog
httplog.RegisterMiddleware("my_middleware", &MyMiddleware{
    URL: "https://httpbin.org/post",
    Username: "admin",
    Password: "admin",
})

// Use is responsible to define what is the default middleware responsible
// to send all logs
httplog.Use("my_middleware")
```
