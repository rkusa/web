package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gocontext "golang.org/x/net/context"
)

func TestWriteHeader(t *testing.T) {
	ctx, _ := createContext(t)

	ctx.WriteHeader(http.StatusBadRequest)
	if ctx.Status() != http.StatusBadRequest {
		t.Errorf("setting status code failed")
	}
}

func TestWrite(t *testing.T) {
	ctx, rec := createContext(t)

	n, err := ctx.Write([]byte("foobar"))
	if err != nil {
		t.Fatal(err)
	}

	if n != 6 {
		t.Errorf("wrong length")
	}

	if ctx.Status() != http.StatusOK {
		t.Errorf("setting status code failed")
	}

	if rec.Body.String() != "foobar" {
		t.Errorf("writing body failed")
	}
}

func TestEvolve(t *testing.T) {
	a, _ := createContext(t)
	b := a.Evolve(gocontext.WithValue(a, "foo", "bar"))

	if val, ok := b.Value("foo").(string); ok != true || val != "bar" {
		t.Errorf("evolve failed")
	}
}

func TestWithValue(t *testing.T) {
	a, _ := createContext(t)
	b := a.WithValue("foo", "bar")

	if val, ok := b.Value("foo").(string); ok != true || val != "bar" {
		t.Errorf("evolve with value failed")
	}
}

func TestString(t *testing.T) {
	a, _ := createContext(t)
	b := a.WithValue("foo", "bar")

	if b.String() != `context.Background.WithValue("foo", "bar").WebContext` {
		t.Errorf("wrong string presentation")
	}
}

func TestBeforeFunc(t *testing.T) {
	ctx, _ := createContext(t)

	beforeCalled := false
	ctx.Before(func(rw http.ResponseWriter) {
		beforeCalled = true
	})

	ctx.WriteHeader(http.StatusOK)
	if beforeCalled != true {
		t.Errorf("before function not called")
	}
}

func createContext(t *testing.T) (Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	return CreateContext(rec, (*http.Request)(nil)), rec
}
