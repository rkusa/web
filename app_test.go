package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	app := New()

	firstCalled, secondCalled, thirdCalled := false, false, false

	app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		firstCalled = true
		next(rw, r)
	})

	app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		secondCalled = true
		// next not called
	})

	app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		thirdCalled = true
		next(rw, r)
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

	app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Set("X-Foo", "bar")
		next(rw, r)
	})

	app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Write([]byte("foobar"))
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

	app.Use(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r)
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, (*http.Request)(nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("unexpected response with status %d", rec.Code)
	}
}
