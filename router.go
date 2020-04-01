package autoroute

import (
	"bytes"
	"errors"
	"net/http"
)

var (
	ErrAlreadyRegistered = errors.New("autoroute: route already registered")
	ErrInvalidMethod     = errors.New("autoroute: not a valid method")
)

// Router implements an autoroute aware grouping of autoroute.Handler's
type Router struct {
	// map[http.Method]map[Path]*Handler
	routeMap map[string]map[string]*Handler

	defaultHandlerOptions []HandlerOption

	defaultErrorHandler ErrorHandler
	NotFoundHandler     http.Handler
}

func NewRouter(handlerOptions ...HandlerOption) (*Router, error) {
	defaultRouteMap := make(map[string]map[string]*Handler)
	defaultRouteMap[http.MethodGet] = make(map[string]*Handler)
	defaultRouteMap[http.MethodPut] = make(map[string]*Handler)
	defaultRouteMap[http.MethodPost] = make(map[string]*Handler)
	defaultRouteMap[http.MethodPatch] = make(map[string]*Handler)
	defaultRouteMap[http.MethodDelete] = make(map[string]*Handler)

	return &Router{
		routeMap:              defaultRouteMap,
		defaultHandlerOptions: handlerOptions,
		defaultErrorHandler:   DefaultErrorHandler,
		NotFoundHandler:       http.NotFoundHandler(),
	}, nil
}

func (ro *Router) Register(method string, path string, x interface{}, extraOptions ...HandlerOption) error {
	defaultOptions := []HandlerOption{
		WithErrorHandler(ro.defaultErrorHandler),
	}

	defaultOptions = append(defaultOptions, ro.defaultHandlerOptions...)
	defaultOptions = append(defaultOptions, extraOptions...)
	h, err := NewHandler(x, defaultOptions...)
	if err != nil {
		return err
	}

	if !methodAllowed(method) {
		return ErrInvalidMethod
	}

	// check no route exists  yet
	routesForMethod := ro.routeMap[method]
	_, ok := routesForMethod[path]
	if ok {
		return ErrAlreadyRegistered
	}

	ro.routeMap[method][path] = h

	return nil
}

func methodAllowed(method string) bool {
	return method == http.MethodDelete ||
		method == http.MethodGet ||
		method == http.MethodPatch ||
		method == http.MethodPost ||
		method == http.MethodPut
}

func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	routesForMethod := ro.routeMap[r.Method]
	handler, ok := routesForMethod[r.URL.Path]
	if !ok {
		ro.NotFoundHandler.ServeHTTP(w, r)

		if r.Body == nil || r.Body == http.NoBody {
			// do nothing
		} else {
			// chew up the rest of the body
			var buf bytes.Buffer
			buf.ReadFrom(r.Body)
			r.Body.Close()
		}
		return
	}

	handler.ServeHTTP(w, r)

	if r.Body == nil || r.Body == http.NoBody {
		// do nothing
	} else {
		// chew up the rest of the body
		var buf bytes.Buffer
		buf.ReadFrom(r.Body)
		r.Body.Close()
	}
}
