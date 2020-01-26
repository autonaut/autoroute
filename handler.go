package autoroute

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"reflect"
	"runtime"
)

var (
	ErrNoFunction    = errors.New("autoroute: not a function passed to NewHandler")
	ErrDecodeFailure = errors.New("autoroute: failure decoding input")
)

type Header map[string]string

var headerValueKey = struct{}{}

func GetHeaders(ctx context.Context) Header {
	v := ctx.Value(headerValueKey)
	return v.(Header)
}

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()
var headerType = reflect.TypeOf(make(Header))

type HandlerOption func(h *Handler)

func WithErrorHandler(errH ErrorHandler) HandlerOption {
	return func(h *Handler) {
		h.errorHandler = errH
	}
}

func WithMaxSizeBytes(maxSizeBytes int64) HandlerOption {
	return func(h *Handler) {
		h.maxSizeBytes = maxSizeBytes
	}
}

func WithCodec(c Codec) HandlerOption {
	return func(h *Handler) {
		h.mimeToCodec[c.Mime()] = c
	}
}

func WithMiddleware(m Middleware) HandlerOption {
	return func(h *Handler) {
		h.middlewares = append(h.middlewares, m)
	}
}

// Handler wraps a function and uses reflection and pluggable codecs to
// automatically create a fast, safe http.Handler from the function
type Handler struct {
	reflectFn     reflect.Value
	reflectFnType reflect.Type
	fnName        string

	mimeToCodec map[string]Codec

	inputArgCount, outputArgCount int

	middlewares []Middleware

	maxSizeBytes int64
	errorHandler ErrorHandler
}

// NewHandler creates an http.Handler from a function that fits a codec-specified
// layout.
func NewHandler(x interface{}, opts ...HandlerOption) (*Handler, error) {
	reflectFn := reflect.ValueOf(x)
	fnName := runtime.FuncForPC(reflectFn.Pointer()).Name()

	if reflectFn.Kind() != reflect.Func {
		return nil, ErrNoFunction
	}

	inputArgCount := reflectFn.Type().NumIn()
	outputArgCount := reflectFn.Type().NumOut()

	h := &Handler{
		reflectFn:     reflectFn,
		reflectFnType: reflectFn.Type(),
		fnName:        fnName,
		inputArgCount: inputArgCount,
		// 65336 bytes
		maxSizeBytes:   2 << 15,
		outputArgCount: outputArgCount,
		mimeToCodec:    make(map[string]Codec),
		errorHandler:   DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(h)
	}

	// prevalidate all loaded codecs
	for _, codec := range h.mimeToCodec {
		err := codec.ValidFn(h.reflectFn)
		if err != nil {
			return nil, err
		}
	}

	return h, nil
}

const MimeTypeHeader = "Content-Type"

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, mw := range h.middlewares {
		err := mw.Before(r, h)
		if err != nil {
			mwe, ok := err.(MiddlewareError)
			if ok {
				w.WriteHeader(mwe.StatusCode)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			h.errorHandler(w, err)
			return
		}
	}

	canonicalMime, _, err := mime.ParseMediaType(r.Header.Get(MimeTypeHeader))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.errorHandler(w, err)
		return
	}

	codec, ok := h.mimeToCodec[canonicalMime]
	if !ok {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var header = make(Header)
	for k := range r.Header {
		hVal := r.Header.Get(k)
		header[k] = hVal
	}

	codec.HandleRequest(&CodecRequestArgs{
		ResponseWriter: w,
		Request:        r,
		Header:         header,
		ErrorHandler:   h.errorHandler,
		HandlerFn:      h.reflectFn,
		HandlerType:    h.reflectFnType,
		InputArgCount:  h.inputArgCount,
		OutputArgCount: h.outputArgCount,
		MaxSizeBytes:   h.maxSizeBytes,
	})
}

func newReflectType(t reflect.Type) reflect.Value {
	// Dereference pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return reflect.New(t)
}
