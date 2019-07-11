package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/autonaut/autoroute"
)

type SplitStringInput struct {
	String string `json:"string"`
}

type SplitStringOutput struct {
	Split []string `json:"split_string"`
}

func SplitString(ssi *SplitStringInput) *SplitStringOutput {
	return &SplitStringOutput{
		Split: strings.Split(ssi.String, " "),
	}
}

func main() {
	// autoroute includes a powerful Router of its own, that's deeply
	// integrated with autoroute's handlers and provides many mechanisms
	// for avoiding duplicated code. For this example, we're showing off
	// our use of standard net/http interfaces by using the standard library
	// http mux.
	mux := http.NewServeMux()

	// create a new *autoroute.Handler (compatible with http.Handler from the stdlib)
	// even though autoroute uses reflection, it provides a handy set of pre-validations
	// to ensure you know ASAP if you've passed something incorrect.
	h, err := autoroute.NewHandler(SplitString, autoroute.WithCodec(autoroute.JSONCodec))
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", h)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
