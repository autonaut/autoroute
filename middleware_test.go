package autoroute

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSignedHeaderMiddleware(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	shm := NewSignedHeadersMiddleware([]string{"x-api-key"}, "test-key")

	signedStr, err := shm.ks.Sign("is-this-signed")
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewHandler(ts.DoThingSignedMiddleware, WithCodec(JSONCodec), WithMiddleware((shm)))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))

	req.Header.Set("x-api-key", signedStr)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("request failed when it should have passed")
	}
}

func TestSignedHeaderMiddlewareRejections(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	shm := NewSignedHeadersMiddleware([]string{"x-api-key"}, "test-key")

	handler, err := NewHandler(ts.DoThingSignedMiddleware, WithCodec(JSONCodec), WithMiddleware((shm)))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))

	req.Header.Set("x-api-key", "this-is-invalid")
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusForbidden {
		t.Error("did not return status verboten")
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	bam := NewBasicAuthMiddleware("user", "user")

	handler, err := NewHandler(ts.DoThingValueArgs, WithCodec(JSONCodec), WithMiddleware((bam)))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))

	req.SetBasicAuth("user", "user")
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Error("request failed when it should have passed")
	}
}

func TestBasicAuthMiddlewareInvalid(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	bam := NewBasicAuthMiddleware("user", "user")

	handler, err := NewHandler(ts.DoThingValueArgs, WithCodec(JSONCodec), WithMiddleware((bam)))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))

	req.SetBasicAuth("user", "incorrect-pwd")
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusForbidden {
		t.Error("request failed when it should have passed")
	}
}
