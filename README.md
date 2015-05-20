# web

Minimal web toolkit build upon [golang.org/x/net/context](https://godoc.org/golang.org/x/net/context).

[![Build Status][drone]](https://ci.rkusa.st/github.com/rkgo/web)
[![GoDoc][godoc]](https://godoc.org/github.com/rkgo/web)

### Example

```go
app := web.New()
app.Use(assert.Middleware())
app.Use(web.Mount("/assets", serve.Dir("public")))
app.Use(logger.Middleware())
app.Use(timeout.Timeout("15s"))
app.Use(render.Middleware(render.Options{
    Directory: "views",
}))
app.Use(sessions.Middleware(
    cookieName,
    sessions.NewCookieStore([]byte("secret")),
))
app.Use(routes.Public())
app.Run("0.0.0.0:3000")
```

[drone]: http://ci.rkusa.st/api/badge/github.com/rkgo/web/status.svg?branch=master&style=flat-square
[godoc]: http://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square