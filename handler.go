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
	ErrNoFunction          = errors.New("autoroute: not a function passed to NewHandler")
	ErrTooManyInputArgs    = errors.New("autoroute: a function can only have up to three input args")
	ErrTooManyOutputArgs   = errors.New("autoroute: a function can only have up to two output args")
	ErrBadErrorHandlerArgs = errors.New("autoroute: error handlers must have two input args")
	ErrDecodeFailure       = errors.New("autoroute: failure decoding input")
)

type Header map[string]string

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

type Codec interface {
	Mime() string
	ValidFn(fn reflect.Value) error
	HandleRequest(w http.ResponseWriter, r *http.Request, eh ErrorHandler, fn reflect.Value, inputArgs, outputArgs int, maxSizeBytes int64)
}

type Handler struct {
	reflectFn reflect.Value
	fnName    string

	mimeToCodec map[string]Codec

	inputArgCount, outputArgCount int

	maxSizeBytes int64
	errorHandler ErrorHandler
}

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

	codec.HandleRequest(w, r, h.errorHandler, h.reflectFn, h.inputArgCount, h.outputArgCount, h.maxSizeBytes)
}

func newReflectType(t reflect.Type) reflect.Value {
	// Dereference pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return reflect.New(t)
}
