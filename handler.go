package autoroute

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		h.errorHandlerFn = reflect.ValueOf(errH)
	}
}

func WithMaxSizeBytes(maxSizeBytes uint64) HandlerOption {
	return func(h *Handler) {
		h.maxSizeBytes = maxSizeBytes
	}
}

type Handler struct {
	reflectFn reflect.Value
	fnName    string

	inputArgCount, outputArgCount int

	maxSizeBytes   uint64
	errorHandlerFn reflect.Value
}

func NewHandler(x interface{}, opts ...HandlerOption) (*Handler, error) {
	reflectFn := reflect.ValueOf(x)
	fnName := runtime.FuncForPC(reflectFn.Pointer()).Name()

	if reflectFn.Kind() != reflect.Func {
		return nil, ErrNoFunction
	}

	inputArgCount := reflectFn.Type().NumIn()
	if inputArgCount > 3 {
		return nil, ErrTooManyInputArgs
	}

	outputArgCount := reflectFn.Type().NumOut()
	if outputArgCount > 2 {
		return nil, ErrTooManyOutputArgs
	}

	h := &Handler{
		reflectFn:     reflectFn,
		fnName:        fnName,
		inputArgCount: inputArgCount,
		// 65336 bytes
		maxSizeBytes:   2 << 15,
		outputArgCount: outputArgCount,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ctx := reflect.ValueOf(r.Context())
	callArgs := make([]reflect.Value, h.inputArgCount)
	switch h.inputArgCount {
	case 3:
		if h.reflectFn.Type().In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(h.reflectFn.Type().In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			panic("functions with two or more input args must have the first one be a context.Context")
		}

		if !(headerType == h.reflectFn.Type().In(1)) {
			panic("autoroute: functions with three input args must have the second be an autoroute.Header")
		} else {
			callArgs[1] = reflect.ValueOf(make(Header))
		}

		callArg, err := h.decode(h.reflectFn.Type().In(2), r.Body)
		if err != nil {
			h.handleErr(w, reflect.ValueOf(err))
			return
		}

		callArgs[2] = callArg
	case 2:
		if h.reflectFn.Type().In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(h.reflectFn.Type().In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			panic("functions with two or more input args must have the first one be a context.Context")
		}

		callArg, err := h.decode(h.reflectFn.Type().In(1), r.Body)
		if err != nil {
			h.handleErr(w, reflect.ValueOf(err))
			return
		}

		callArgs[1] = callArg
	case 1:
		// here, our first arg is an interface
		if h.reflectFn.Type().In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(h.reflectFn.Type().In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			callArg, err := h.decode(h.reflectFn.Type().In(0), r.Body)
			if err != nil {
				h.handleErr(w, reflect.ValueOf(err))
				return
			}

			callArgs[0] = callArg
		}
	case 0:
		// do nothing
	default:
		panic("autoroute: can only have up to three input args")
	}

	outputValues := h.reflectFn.Call(callArgs)
	switch h.outputArgCount {
	case 2:
		// if err == nil
		if outputValues[1].IsNil() {
			err := json.NewEncoder(w).Encode(outputValues[0].Interface())
			if err != nil {
				panic(err)
			}
			return
		}

		if outputValues[1].Kind() == reflect.Interface {
			if outputValues[1].Type().ConvertibleTo(errorType) {
				h.handleErr(w, outputValues[1])
				return
			}
		}
	case 1:
		if outputValues[0].Kind() == reflect.Interface {
			if outputValues[0].Type().ConvertibleTo(errorType) {
				h.handleErr(w, outputValues[0])
				return
			}
		}

		err := json.NewEncoder(w).Encode(outputValues[0].Interface())
		if err != nil {
			panic(err)
		}
	case 0:
		w.WriteHeader(http.StatusOK)
	}
}

func newReflectType(t reflect.Type) reflect.Value {
	// Dereference pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return reflect.New(t)
}

func (h *Handler) defaultErrorHandler(w http.ResponseWriter, x error) {
	if x == ErrDecodeFailure {
		w.WriteHeader(http.StatusBadRequest)
	}

	outputErrorString := fmt.Sprintf("%s", x)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": outputErrorString})
}

func (h *Handler) handleErr(w http.ResponseWriter, errorValue reflect.Value) {
	errConv := errorValue.Convert(errorType)
	if h.errorHandlerFn.IsValid() {
		h.errorHandlerFn.Call([]reflect.Value{reflect.ValueOf(w), errConv})
	} else {
		defaultErrorHandler := reflect.ValueOf(h.defaultErrorHandler)
		defaultErrorHandler.Call([]reflect.Value{reflect.ValueOf(w), errConv})
	}
}

func (h *Handler) decode(inArg reflect.Type, body io.ReadCloser) (reflect.Value, error) {
	var object reflect.Value

	switch inArg.Kind() {
	case reflect.Struct:
		object = newReflectType(inArg)
	case reflect.Ptr:
		object = newReflectType(inArg)
	}

	oi := object.Interface()
	err := json.NewDecoder(io.LimitReader(body, int64(h.maxSizeBytes))).Decode(&oi)
	if err != nil {
		if err == io.EOF {
			return reflect.Value{}, ErrDecodeFailure
			// literally do nothing if we got no body
		} else {
			return reflect.Value{}, err
		}
	}

	switch inArg.Kind() {
	case reflect.Struct:
		return reflect.ValueOf(oi).Elem(), nil
	case reflect.Ptr:
		return reflect.ValueOf(oi), nil
	}

	return reflect.Value{}, ErrDecodeFailure
}
