# web

Minimal web toolkit build upon the Go 1.7 immutable [context](https://golang.org/pkg/context).

[![Build Status][travis]](https://travis-ci.org/rkusa/web)
[![GoDoc][godoc]](https://godoc.org/github.com/rkusa/web)

### Example

```go
app := web.New()
app.Use(assert.Middleware())
app.Use(web.Mount("/assets", serve.Dir("public")))
app.Use(logger.Middleware())
app.Use(timeout.Timeout("15s"))
app.Use(sessions.Middleware(
    cookieName,
    sessions.NewCookieStore([]byte("secret")),
))
app.Use(routes.Public())
if err := app.Run("0.0.0.0:3000"); err != nil {
    log.Fatal(err)
}
```

### Request-scoped data

The following shows some examples where request-scoped data is used.

#### Router

The following example shows how the context can be utilized for middlewares to add own data
to the request context. In this example, a [router](https://github.com/rkusa/router) middleware
adds the path parameters to the context.

```go
r := router.New()
// Use http.HandlerFunc footprint, but ...
r.GET("/user/:id", func(rw http.ResponseWriter, r *http.Request) {
  // ... still be able to retrieve router specific values like params
  id := router.Param(r, "id")
  // id := router.ParamFromContext(r.Context(), "id")
})
```

#### Authentication / Current User

The following example authenticates the currently logged in user. On successfull authentication
the user is added to the context. This allows accessing the current user on succeeding
middlewares and routes.

```go
type User struct {}

type key int
var userKey key = 1

func (user *User) NewContext(ctx context.Context) context.Context {
  return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) *User {
  user, ok := ctx.Value(userKey).(*User)
  if !ok {
    return nil
  }
  return user
}

// middleware to ensure authentication
app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  // ...
  if user.IsAuthenticated() {
    // continue in the middleware stack with the user being added to the context
    next(rw, r.WithContext(user.NewContext(r.Context())))
  } else {
    rw.WriteHeader(http.StatusUnauthorized)
  }
})

// route that uses the previuously added user
r.POST("/articles", func(rw http.ResponseWriter, r *http.Request) {
  user := UserFromContext(r.Context())
  // ...
})
```

### Works well with

- [http-assert](https://github.com/rkusa/http-assert)
- [http-json](https://github.com/rkusa/http-json)
- [logger](https://github.com/rkusa/logger)
- [router](https://github.com/rkusa/router)
- [serve](https://github.com/rkusa/serve)
- [sessions](https://github.com/rkusa/sessions)
- [timeout](https://github.com/rkusa/timeout)

## License

[MIT](LICENSE)

[travis]: https://api.travis-ci.org/rkusa/web.svg
[godoc]: http://img.shields.io/badge/godoc-reference-blue.svg
