package autoroute

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

// JSONCodec implements autoroute functionality for the mime type application/json
// and functions that have up to three input args and two output args.
// a function with three input args must look like func(context.Context, autoroute.Header, anyStructOrPointer)
// with two it can be func(context.Context, autoroute.Header) or func(context.Context, anyStructOrPointer)
// with one it can be any of func(context.Context), func(autoroute.Header), or func(anyStructOrPointer)
// in terms of output values, a function can return `(anyStructOrPointer, error)`, `(anyStructOrPointer)`,
// (error), or nothing.
// the JSONCodec will attempt to decode values in two ways
// 1. use encoding/json on the request body
// 2. decode the URL parameters of a GET request into the struct
// and it will always JSON encode the output value.
var JSONCodec Codec = jsonCodec{}

type jsonCodec struct{}

func (js jsonCodec) Mime() string {
	return "application/json"
}

func (js jsonCodec) ValidFn(fn reflect.Value) error {
	inputArgCount := fn.Type().NumIn()
	if inputArgCount > 3 {
		return ErrTooManyInputArgs
	}

	outputArgCount := fn.Type().NumOut()
	if outputArgCount > 2 {
		return ErrTooManyOutputArgs
	}

	return nil
}

func (js jsonCodec) HandleRequest(cra *CodecRequestArgs) {
	ctx := reflect.ValueOf(cra.Request.Context())
	callArgs := make([]reflect.Value, cra.InputArgCount)
	switch cra.InputArgCount {
	case 3:
		if cra.HandlerType.In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(cra.HandlerType.In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			panic("functions with two or more input args must have the first one be a context.Context")
		}

		if !(headerType == cra.HandlerType.In(1)) {
			panic("autoroute: functions with three input args must have the second be an autoroute.Header")
		} else {
			callArgs[1] = reflect.ValueOf(cra.Header)
		}

		callArg, err := js.decode(cra.HandlerType.In(2), cra.Request.Body, cra.MaxSizeBytes)
		if err != nil {
			cra.ErrorHandler.Handle(cra.ResponseWriter, reflect.ValueOf(err))
			return
		}

		callArgs[2] = callArg
	case 2:
		if cra.HandlerType.In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(cra.HandlerType.In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			panic("functions with two or more input args must have the first one be a context.Context")
		}

		handlerInArgType := cra.HandlerType.In(1)

		var callArgOne reflect.Value
		var err error
		if headerType == handlerInArgType {
			callArgOne = reflect.ValueOf(cra.Header)
		} else {
			callArgOne, err = js.decode(handlerInArgType, cra.Request.Body, cra.MaxSizeBytes)
		}
		if err != nil {
			cra.ErrorHandler.Handle(cra.ResponseWriter, reflect.ValueOf(err))
			return
		}

		callArgs[1] = callArgOne
	case 1:
		inArg := cra.HandlerType.In(0)
		// here, our first arg is an interface
		if inArg.Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(inArg) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else if headerType == inArg {
			callArgs[0] = reflect.ValueOf(cra.Header)
		} else {
			callArg, err := js.decode(inArg, cra.Request.Body, cra.MaxSizeBytes)
			if err != nil {
				cra.ErrorHandler.Handle(cra.ResponseWriter, reflect.ValueOf(err))
				return
			}

			callArgs[0] = callArg
		}
	case 0:
		// do nothing
	default:
		panic("autoroute: can only have up to three input args")
	}

	outputValues := cra.HandlerFn.Call(callArgs)
	cra.ResponseWriter.Header().Set("Content-Type", "application/json")
	switch cra.OutputArgCount {
	case 2:
		// if err == nil
		if outputValues[1].IsNil() {
			err := json.NewEncoder(cra.ResponseWriter).Encode(outputValues[0].Interface())
			if err != nil {
				panic(err)
			}
			return
		}

		if outputValues[1].Kind() == reflect.Interface {
			if outputValues[1].Type().ConvertibleTo(errorType) {
				cra.ErrorHandler.Handle(cra.ResponseWriter, outputValues[1])
				return
			}
		}
	case 1:
		if outputValues[0].Kind() == reflect.Interface {
			if outputValues[0].Type().ConvertibleTo(errorType) {
				cra.ErrorHandler.Handle(cra.ResponseWriter, outputValues[0])
				return
			}
		}

		err := json.NewEncoder(cra.ResponseWriter).Encode(outputValues[0].Interface())
		if err != nil {
			panic(err)
		}
	case 0:
		cra.ResponseWriter.WriteHeader(http.StatusOK)
	}
}

func (js jsonCodec) decode(inArg reflect.Type, body io.ReadCloser, maxSizeBytes int64) (reflect.Value, error) {
	var object reflect.Value

	switch inArg.Kind() {
	case reflect.Struct:
		object = newReflectType(inArg)
	case reflect.Ptr:
		object = newReflectType(inArg)
	}

	oi := object.Interface()
	err := json.NewDecoder(io.LimitReader(body, maxSizeBytes)).Decode(&oi)
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
