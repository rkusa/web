// Minimal web toolkit build upon
// [golang.org/x/net/context](https://godoc.org/golang.org/x/net/context)
//
//  app := web.New()
//  app.Use(assert.Middleware())
//  app.Use(web.Mount("/assets", serve.Dir("public")))
//  app.Use(logger.Middleware())
//  app.Use(timeout.Timeout("15s"))
//  app.Use(render.Middleware(render.Options{
//      Directory: "views",
//  }))
//  app.Use(sessions.Middleware(
//      cookieName,
//      sessions.NewCookieStore([]byte("secret")),
//  ))
//  app.Use(routes.Public())
//  app.Run("0.0.0.0:3000")
//
package web

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-errors/errors"
	gocontext "golang.org/x/net/context"
)

// The function is used to call the next middleware.
type Next func(Context)

// A middleware.
type Middleware func(ctx Context, next Next)

type link struct {
	middleware Middleware
	next       *link
}

// The App is used to register middlewares and serve them through HTTP.
type App interface {
	// Add a middleware to the middleware stack.
	Use(Middleware)

	// Handler can be used to use a http.Handler as a middleware.
	Handler(http.Handler) Middleware

	// Execute the middleware stack using the provided context.
	Execute(Context, Next)

	// Serve the middleware stack through HTTP.
	ServeHTTP(http.ResponseWriter, *http.Request)

	// Run a http server on the given address.
	Run(addr string)
}

type app struct {
	first *link
	last  *link
}

// New creates a new App.
func New() App {
	return &app{}
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

func (a *app) Handler(handler http.Handler) Middleware {
	return func(ctx Context, next Next) {
		handler.ServeHTTP(ctx, ctx.Req())
		next(ctx)
	}
}

func (a *app) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	cancelable, cancel := gocontext.WithCancel(gocontext.Background())
	defer cancel()

	ctx := CreateContextWith(cancelable, rw, r)

	c := make(chan error, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				c <- errors.Wrap(err, 0)
			} else {
				c <- nil
			}
		}()

		a.Execute(ctx, func(ctx Context) {
			http.NotFound(ctx, ctx.Req())
		})
	}()

	select {
	case <-ctx.Done():
		handleError(ctx.Err(), rw)
	case err := <-c:
		handleError(err, rw)
	}
}

func (a *app) Execute(ctx Context, done Next) {
	link := a.first

	var next Next
	next = func(ctx Context) {
		if link == nil {
			done(ctx)
			return
		}

		current := link
		link = current.next
		current.middleware(ctx, next)
	}

	next(ctx)
}

// Combine can be used to combine multiple middlewares into one stack.
func Combine(mw ...Middleware) Middleware {
	combined := New()

	for _, m := range mw {
		combined.Use(m)
	}

	return func(ctx Context, next Next) {
		combined.Execute(ctx, next)
	}
}

// Mount can be used to mount a middleware to the specified path / path-prefix.
// If the the path matches, the matched part ist trimmed and the middleware is
// called. Otherwise, the middleware is skipped.
func Mount(path string, mw Middleware) Middleware {
	path = strings.TrimSuffix(path, "/")

	return func(ctx Context, next Next) {
		url := ctx.Req().URL
		if url.Path == path || strings.HasPrefix(url.Path, path+"/") {
			before := url.Path

			url.Path = strings.TrimPrefix(url.Path, path)
			if len(url.Path) == 0 {
				url.Path = "/"
			}

			mw(ctx, func(ctx Context) {
				url.Path = before
				next(ctx)
			})
		} else {
			next(ctx)
		}
	}
}

func handleError(e error, rw http.ResponseWriter) {
	if e != nil {
		err := errors.Wrap(e, 0)
		http.Error(rw, err.Error(), http.StatusInternalServerError)

		l := log.New(os.Stdout, "[web] ", 0)
		l.Printf("PANIC: %s\n%s", err.Error(), err.ErrorStack())
	}
}

func (a *app) Run(addr string) {
	l := log.New(os.Stdout, "[web] ", 0)
	l.Printf("listening on %s", addr)
	l.Fatal(http.ListenAndServe(addr, a))
}
