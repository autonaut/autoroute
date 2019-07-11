package autoroute

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

type ErrorHandler func(w http.ResponseWriter, e error)

func (eh ErrorHandler) Handle(w http.ResponseWriter, errorValue reflect.Value) {
	errConv := errorValue.Convert(errorType)
	ehFn := reflect.ValueOf(eh)
	ehFn.Call([]reflect.Value{reflect.ValueOf(w), errConv})
}

func DefaultErrorHandler(w http.ResponseWriter, x error) {
	if x == ErrDecodeFailure {
		w.WriteHeader(http.StatusBadRequest)
	}

	outputErrorString := fmt.Sprintf("%s", x)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": outputErrorString})
}
