package autoroute

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
)

type JSONCodec struct{}

func (js JSONCodec) Mime() string {
	return "application/json"
}

func (js JSONCodec) ValidFn(fn reflect.Value) error {
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

func (js JSONCodec) HandleRequest(w http.ResponseWriter, r *http.Request, eh ErrorHandler, fn reflect.Value, inputArgs, outputArgs int, maxSizeBytes int64) {
	ctx := reflect.ValueOf(r.Context())
	fnType := fn.Type()
	callArgs := make([]reflect.Value, inputArgs)
	switch inputArgs {
	case 3:
		if fnType.In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(fnType.In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			panic("functions with two or more input args must have the first one be a context.Context")
		}

		if !(headerType == fnType.In(1)) {
			panic("autoroute: functions with three input args must have the second be an autoroute.Header")
		} else {
			callArgs[1] = reflect.ValueOf(make(Header))
		}

		callArg, err := js.decode(fnType.In(2), r.Body, maxSizeBytes)
		if err != nil {
			eh.Handle(w, reflect.ValueOf(err))
			return
		}

		callArgs[2] = callArg
	case 2:
		if fnType.In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(fnType.In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			panic("functions with two or more input args must have the first one be a context.Context")
		}

		callArg, err := js.decode(fnType.In(1), r.Body, maxSizeBytes)
		if err != nil {
			eh.Handle(w, reflect.ValueOf(err))
			return
		}

		callArgs[1] = callArg
	case 1:
		// here, our first arg is an interface
		if fnType.In(0).Kind() == reflect.Interface {
			// if it implements context.Context
			if contextType.Implements(fnType.In(0)) {
				callArgs[0] = ctx
			} else {
				panic("got a non context.Context interface type as the first arg")
			}
		} else {
			callArg, err := js.decode(fnType.In(0), r.Body, maxSizeBytes)
			if err != nil {
				eh.Handle(w, reflect.ValueOf(err))
				return
			}

			callArgs[0] = callArg
		}
	case 0:
		// do nothing
	default:
		panic("autoroute: can only have up to three input args")
	}

	outputValues := fn.Call(callArgs)
	w.Header().Set("Content-Type", "application/json")
	switch outputArgs {
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
				eh.Handle(w, outputValues[1])
				return
			}
		}
	case 1:
		if outputValues[0].Kind() == reflect.Interface {
			if outputValues[0].Type().ConvertibleTo(errorType) {
				eh.Handle(w, outputValues[0])
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

func (js JSONCodec) decode(inArg reflect.Type, body io.ReadCloser, maxSizeBytes int64) (reflect.Value, error) {
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
