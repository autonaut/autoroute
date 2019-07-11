package autoroute

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

// An ErrorHandler is responsible for writing an error back to the calling
// http client
type ErrorHandler func(w http.ResponseWriter, e error)

// Handle is a convienience method on ErrorHandler that allows it to call itself
// with a reflect.Value
func (eh ErrorHandler) Handle(w http.ResponseWriter, errorValue reflect.Value) {
	errConv := errorValue.Convert(errorType)
	ehFn := reflect.ValueOf(eh)
	ehFn.Call([]reflect.Value{reflect.ValueOf(w), errConv})
}

// DefaultErrorHandler writes json `{"error": "errString"}`
func DefaultErrorHandler(w http.ResponseWriter, x error) {
	if x == ErrDecodeFailure {
		w.WriteHeader(http.StatusBadRequest)
	}

	outputErrorString := fmt.Sprintf("%s", x)
	json.NewEncoder(w).Encode(map[string]interface{}{"error": outputErrorString})
}
