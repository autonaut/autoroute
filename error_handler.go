package autoroute

import "net/http"

type ErrorHandler func(w http.ResponseWriter, e error)