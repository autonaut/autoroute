package autonaut

import (
	"errors"
	"net/http"
)

var (
	ErrAlreadyRegistered = errors.New("autonaut: route already registered")
	ErrInvalidMethod     = errors.New("autonaut: not a valid method")
)

// Router implements an autonaut aware grouping of autonaut.Handler's
type Router struct {
	// map[http.Method]map[Path]*Handler
	routeMap map[string]map[string]*Handler

	defaultErrorHandler ErrorHandler
	notFoundHandler     http.Handler

	decodeHeaders []string
	keySigner     *keySigner
}

func NewRouter() (*Router, error) {
	defaultRouteMap := make(map[string]map[string]*Handler)
	defaultRouteMap[http.MethodGet] = make(map[string]*Handler)
	defaultRouteMap[http.MethodPut] = make(map[string]*Handler)
	defaultRouteMap[http.MethodPost] = make(map[string]*Handler)
	defaultRouteMap[http.MethodPatch] = make(map[string]*Handler)
	defaultRouteMap[http.MethodDelete] = make(map[string]*Handler)

	return &Router{
		routeMap:        defaultRouteMap,
		notFoundHandler: http.NotFoundHandler(),
	}, nil
}

func (ro *Router) Register(method string, path string, x interface{}, extraOptions ...HandlerOption) error {
	defaultOptions := []HandlerOption{
		WithErrorHandler(ro.defaultErrorHandler),
	}

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
		ro.notFoundHandler.ServeHTTP(w, r)
		return
	}

	handler.ServeHTTP(w, r)
}
