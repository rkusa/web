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
app.Use(render.Middleware(render.Options{
    Directory: "views",
}))
app.Use(sessions.Middleware(
    cookieName,
    sessions.NewCookieStore([]byte("secret")),
))
app.Use(routes.Public())
if err := app.Run("0.0.0.0:3000"); err != nil {
    log.Fatal(err)
}
```

### Works well with

- [http-assert](https://github.com/rkusa/http-assert)
- [serve](https://github.com/rkusa/serve)
- [logger](https://github.com/rkusa/logger)
- [sessions](https://github.com/rkusa/sessions)
- [router](https://github.com/rkusa/router)
- [http-json](https://github.com/rkusa/http-json)

## License

[MIT](LICENSE)

[travis]: https://api.travis-ci.org/rkusa/web.svg
[godoc]: http://img.shields.io/badge/godoc-reference-blue.svg
