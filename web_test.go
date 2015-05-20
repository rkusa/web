package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	app := New()

	firstCalled, secondCalled, thirdCalled := false, false, false

	app.Use(func(ctx Context, next Next) {
		firstCalled = true
		next(ctx)
	})

	app.Use(func(ctx Context, next Next) {
		secondCalled = true
		// next not called
	})

	app.Use(func(ctx Context, next Next) {
		thirdCalled = true
		next(ctx)
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, (*http.Request)(nil))

	if firstCalled != true {
		t.Errorf("first middleware not called")
	}

	if secondCalled != true {
		t.Errorf("second middleware not called")
	}

	if thirdCalled != false {
		t.Errorf("third middleware called")
	}
}

func TestServeHTTP(t *testing.T) {
	app := New()

	app.Use(func(ctx Context, next Next) {
		ctx.Header().Set("X-Foo", "bar")
		next(ctx)
	})

	app.Use(func(ctx Context, next Next) {
		ctx.Write([]byte("foobar"))
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, (*http.Request)(nil))

	if rec.Code != http.StatusOK {
		t.Errorf("request failed with status %d", rec.Code)
	}

	if rec.Body.String() != "foobar" {
		t.Errorf("invalid body: %s", rec.Body.String())
	}

	if rec.Header().Get("X-Foo") != "bar" {
		t.Errorf("header not set")
	}
}

func TestNotFound(t *testing.T) {
	app := New()

	app.Use(func(ctx Context, next Next) {
		next(ctx)
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, (*http.Request)(nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("unexpected response with status %d", rec.Code)
	}
}

func TestCombine(t *testing.T) {
	app := New()

	firstCalled, secondCalled, thirdCalled := false, false, false

	app.Use(Combine(
		func(ctx Context, next Next) {
			firstCalled = true
			next(ctx)
		},

		func(ctx Context, next Next) {
			secondCalled = true
			// next not called
		},

		func(ctx Context, next Next) {
			thirdCalled = true
			next(ctx)
		},
	))

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, (*http.Request)(nil))

	if firstCalled != true {
		t.Errorf("first middleware not called")
	}

	if secondCalled != true {
		t.Errorf("second middleware not called")
	}

	if thirdCalled != false {
		t.Errorf("third middleware called")
	}
}

func TestMount(t *testing.T) {
	app := New()

	app.Use(Mount("/bar", func(ctx Context, next Next) {
		t.Errorf("unexpected path match")
	}))

	app.Use(Mount("/foo", func(ctx Context, next Next) {
		if ctx.Req().URL.Path != "/bar" {
			t.Errorf("path not trimmed properly, got: %s", ctx.Req().URL.Path)
		}
	}))

	rec := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	app.ServeHTTP(rec, r)
}
