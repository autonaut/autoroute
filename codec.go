package autoroute

import (
	"net/http"
	"reflect"
)

// A Codec implements pluggable, mime-type based serialization and deserialization
// for autoroute based Handlers
type Codec interface {
	Mime() string
	ValidFn(fn reflect.Value) error
	HandleRequest(*CodecRequestArgs)
}

// CodecRequestArgs is passed to a Codec when it matches the mime type
// of a given request
type CodecRequestArgs struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	ErrorHandler   ErrorHandler

	// Our nice bit of reflection stuff to work with
	HandlerFn                     reflect.Value
	HandlerType                   reflect.Type
	InputArgCount, OutputArgCount int

	MaxSizeBytes int64
}
