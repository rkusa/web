package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCombine(t *testing.T) {
	app := New()

	firstCalled, secondCalled, thirdCalled := false, false, false

	app.Use(Combine(
		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			firstCalled = true
			next(rw, r)
		},

		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			secondCalled = true
			// next not called
		},

		func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			thirdCalled = true
			next(rw, r)
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

	app.Use(Mount("/bar", func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		t.Errorf("unexpected path match")
	}))

	app.Use(Mount("/foo", func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if r.URL.Path != "/bar" {
			t.Errorf("path not trimmed properly, got: %s", r.URL.Path)
		}
	}))

	rec := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/foo/bar", nil)
	if err != nil {
		t.Fatal(err)
	}
	app.ServeHTTP(rec, r)
}
