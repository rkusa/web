package web

import (
	"net/http"
	"strings"
)

// Combine can be used to combine multiple middlewares into one stack.
func Combine(mw ...Middleware) Middleware {
	combined := New()

	for _, m := range mw {
		combined.Use(m)
	}

	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		combined.Execute(rw, r, next)
	}
}

// Mount can be used to mount a middleware to the specified path / path-prefix.
// If the the path matches, the matched part ist trimmed and the middleware is
// called. Otherwise, the middleware is skipped.
func Mount(path string, mw Middleware) Middleware {
	path = strings.TrimSuffix(path, "/")

	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		url := r.URL
		if url.Path == path || strings.HasPrefix(url.Path, path+"/") {
			before := url.Path

			url.Path = strings.TrimPrefix(url.Path, path)
			if len(url.Path) == 0 {
				url.Path = "/"
			}

			mw(rw, r, func(rw http.ResponseWriter, r *http.Request) {
				url.Path = before
				next(rw, r)
			})
		} else {
			next(rw, r)
		}
	}
}

// UseHandler wraps a http.Handler into a Middleware.
func Handler(handler http.Handler) Middleware {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		handler.ServeHTTP(rw, r)
		if rw, ok := rw.(defaultResponseWriter); ok && !rw.Written() {
			next(rw, r)
		}
	}
}

// UseHandler wraps a http.HandlerFunc into a Middleware.
func Func(fn http.HandlerFunc) Middleware {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		fn(rw, r)
		if rw, ok := rw.(defaultResponseWriter); ok && !rw.Written() {
			next(rw, r)
		}
	}
}
