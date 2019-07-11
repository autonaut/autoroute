package autoroute

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type TestServer struct {
	requests int
	input    string
}

type TestInput struct {
	Input string `json:"input"`
}

type TestOutput struct {
	Output string `json:"output"`
}

func (t *TestServer) DoThing(ti *TestInput) *TestOutput {
	t.requests += 1
	t.input = ti.Input

	return &TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingCtx(ctx context.Context, ti *TestInput) *TestOutput {
	t.requests += 1
	t.input = ti.Input

	return &TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingCtxValueArgs(ctx context.Context, ti TestInput) TestOutput {
	t.requests += 1
	t.input = ti.Input

	return TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingNoInputArgs() TestOutput {
	t.requests += 1

	return TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingNoInputArgsTwoOutput() (*TestOutput, error) {
	t.requests += 1

	return &TestOutput{
		Output: "hi",
	}, nil
}

func (t *TestServer) DoThingNoInputArgsTwoOutputError() (*TestOutput, error) {
	t.requests += 1

	return nil, errors.New("sup")
}

func (t *TestServer) DoThingErrorReturn() error {
	t.requests += 1

	return errors.New("sup")
}

func (t *TestServer) DoThingAllArgs(ctx context.Context, h Header, ti *TestInput) TestOutput {
	t.requests += 1
	t.input = ti.Input

	return TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingNoInputArgsPtrOutput() *TestOutput {
	t.requests += 1

	return &TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingNoInputArgsMapOutput() map[string]string {
	t.requests += 1

	return map[string]string{
		"output": "hi",
	}
}

func (t *TestServer) DoThingValueArgs(ti TestInput) TestOutput {
	t.requests += 1
	t.input = ti.Input

	return TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingTooManyArgs(a, b, c, d int) TestOutput {
	t.requests += 1

	return TestOutput{
		Output: "hi",
	}
}

func (t *TestServer) DoThingInvalidTwoArgs(ein int, ti *TestInput) *TestOutput {
	t.requests += 1
	t.input = ti.Input

	return &TestOutput{
		Output: "hi",
	}
}

func TestHandlerBasic(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThing, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.input != "yo" {
		t.Fatal("did not decode input properly")
	}

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerCtxArg(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingCtx, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.input != "yo" {
		t.Fatal("did not decode input properly")
	}

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerCtxValueArgs(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingCtxValueArgs, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.input != "yo" {
		t.Fatal("did not decode input properly")
	}

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerValueArgs(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingValueArgs, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.input != "yo" {
		t.Fatal("did not decode input properly")
	}

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerNoInputArgs(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingNoInputArgs, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerNoInputArgsPtrOutput(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingNoInputArgsPtrOutput, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerNoInputArgsMapOutput(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingNoInputArgsMapOutput, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerAllArgs(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingAllArgs, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"input": "yo"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.input != "yo" {
		t.Fatal("did not decode input properly")
	}

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerErrorReturn(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingErrorReturn, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"error":"sup"}`+"\n", w.Body.String())

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerTwoOutputErrorReturn(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingNoInputArgsTwoOutputError, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"error":"sup"}`+"\n", w.Body.String())

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerTwoOutputErrorReturnCustomErrorHandler(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	var innerErr error
	handler, err := NewHandler(ts.DoThingNoInputArgsTwoOutputError, WithCodec(JSONCodec), WithErrorHandler(func(w http.ResponseWriter, err error) {
		innerErr = err
	}))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	if innerErr.Error() != "sup" {
		t.Fatalf("did not get error back properly, got %s", innerErr)
	}

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerTwoOutputStructReturn(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	handler, err := NewHandler(ts.DoThingNoInputArgsTwoOutput, WithCodec(JSONCodec))
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(w, req)

	diffJSON(t, `{"output":"hi"}`+"\n", w.Body.String())

	if ts.requests != 1 {
		t.Fatal("did not actually call function")
	}
}

func TestHandlerArgLengthValidation(t *testing.T) {
	t.Parallel()
	ts := &TestServer{}

	_, err := NewHandler(ts.DoThingTooManyArgs, WithCodec(JSONCodec))
	if err == nil {
		t.Fatal("did not get error for too many args")
	}
}
