package web

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	gocontext "golang.org/x/net/context"
)

type BeforeResponseFunc func(http.ResponseWriter)

type Context interface {
	gocontext.Context
	http.ResponseWriter
	http.Hijacker

	// Evolve returns a new context with the provided context being set as parent.
	Evolve(ctx gocontext.Context) Context

	// WithValue returns a new context created from the given key/value.
	WithValue(interface{}, interface{}) Context

	// Req returns the HTTP request.
	Req() *http.Request

	// Before registers the given function to be called before the status code
	// is set, ie, before WriteHeader or Write are executed.
	Before(fn BeforeResponseFunc)

	// Status returns the current status code.
	Status() int

	// Redirect is a helper method the redirects the current request.
	Redirect(path string)

	// String returns a string representation of the current context
	// (for debug purposes)
	String() string
}

type context struct {
	gocontext.Context
	http.ResponseWriter

	req         *http.Request
	status      *int
	beforeFuncs []BeforeResponseFunc
}

func (c *context) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *context) WriteHeader(code int) {
	for _, fn := range c.beforeFuncs {
		fn(c.ResponseWriter)
	}

	*c.status = code
	c.ResponseWriter.WriteHeader(*c.status)
}

func (c *context) Before(fn BeforeResponseFunc) {
	c.beforeFuncs = append(c.beforeFuncs, fn)
}

func (c *context) Write(b []byte) (int, error) {
	if *c.status == 0 {
		c.WriteHeader(http.StatusOK)
	}

	return c.ResponseWriter.Write(b)
}

func (c *context) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// c.size = 0
	return c.ResponseWriter.(http.Hijacker).Hijack()
}

func (c *context) Status() int {
	return *c.status
}

func (c *context) Req() *http.Request {
	return c.req
}

func (c *context) Evolve(ctx gocontext.Context) Context {
	return &context{ctx, c.ResponseWriter, c.req, c.status, c.beforeFuncs}
}

func (c *context) WithValue(key, val interface{}) Context {
	return &context{gocontext.WithValue(c.Context, key, val), c.ResponseWriter, c.req, c.status, c.beforeFuncs}
}

func (c *context) String() string {
	return fmt.Sprintf("%v.WebContext", c.Context)
}

func (c *context) Redirect(path string) {
	defer http.Redirect(c, c.req, path, http.StatusFound)
}

// Create creates and initializes a new Context fromt the given
// http.ResponseWriter and http.Request.
func CreateContext(rw http.ResponseWriter, r *http.Request) Context {
	return CreateContextWith(gocontext.Background(), rw, r)
}

func CreateContextWith(ctx gocontext.Context, rw http.ResponseWriter, r *http.Request) Context {
	status := 0
	return &context{
		Context:        ctx,
		ResponseWriter: rw,
		req:            r,
		status:         &status,
		beforeFuncs:    make([]BeforeResponseFunc, 0),
	}
}
