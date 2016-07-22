// Package web is a minimal web middleware toolkit build upon build upon the
// Go 1.7 immutable [context](https://golang.org/pkg/context).
//
//  app := web.New()
//  app.Use(assert.Middleware())
//  app.Use(web.Mount("/assets", serve.Dir("public")))
//  app.Use(logger.Middleware())
//  app.Use(timeout.Timeout("15s"))
//  app.Use(sessions.Middleware(
//      cookieName,
//      sessions.NewCookieStore([]byte("secret")),
//  ))
//  app.Use(routes.Public())
//  if err := app.Run("0.0.0.0:3000"); err != nil {
//    log.Fatal(err)
//  }
//
package web

import (
	"bufio"
	"log"
	"net"
	"net/http"
	// TODO: re-consider go-errors
)

// TODO: Redirect
//

// A middleware.
type Middleware func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)

// The App is used to register middlewares and serve them through HTTP.
type App interface {
	// Add a middleware to the middleware stack.
	Use(Middleware)

	// UseHandler can be used to use a http.Handler as a middleware.
	UseHandler(http.Handler)

	// UseFunc can be used to use a http.HandlerFunc as a middleware.
	UseFunc(http.HandlerFunc)

	// Execute the middleware stack using the provided context.
	Execute(http.ResponseWriter, *http.Request, http.HandlerFunc)

	// Serve the middleware stack through HTTP.
	ServeHTTP(http.ResponseWriter, *http.Request)

	// Run a http server on the given address.
	Run(addr string) error

	// Run a https server on the given address using the provided certificate.
	RunTLS(addr, certFile, keyFile string) error
}

type link struct {
	middleware Middleware
	next       *link
}

type app struct {
	first *link
	last  *link
}

// New creates a new App.
func New() App {
	return &app{nil, nil}
}

func (a *app) Use(mw Middleware) {
	if a.first == nil {
		a.first = &link{mw, nil}
		a.last = a.first
	} else {
		a.last.next = &link{mw, nil}
		a.last = a.last.next
	}
}

func (a *app) UseHandler(handler http.Handler) {
	a.Use(Handler(handler))
}

func (a *app) UseFunc(fn http.HandlerFunc) {
	a.Use(Func(fn))
}

func (a *app) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw = &responseWriter{rw, 0, false}

	c := make(chan error, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				c <- err.(error)
			} else {
				c <- nil
			}
		}()

		a.Execute(rw, r, func(rw http.ResponseWriter, r *http.Request) {
			http.NotFound(rw, r)
		})
	}()

	a.handleError(<-c, rw)
}

func (a *app) Execute(rw http.ResponseWriter, r *http.Request, done http.HandlerFunc) {
	link := a.first

	var next http.HandlerFunc
	next = func(rw http.ResponseWriter, r *http.Request) {
		if link == nil {
			done(rw, r)
			return
		}

		current := link
		link = current.next
		current.middleware(rw, r, next)
	}

	next(rw, r)
}

func (a *app) handleError(err error, rw http.ResponseWriter) {
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("PANIC: %s", err.Error())
	}
}

func (a *app) Run(addr string) error {
	return http.ListenAndServe(addr, a)
}

func (a *app) RunTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, a)
}

// Custom response writer to keep track whether something has already been written.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	rw.written = true
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.written = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return rw.ResponseWriter.(http.Hijacker).Hijack()
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Written() bool {
	return rw.written
}
