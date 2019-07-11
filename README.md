autoroute [![Build Status](https://travis-ci.org/autonaut/autoroute.svg?branch=master)](https://travis-ci.org/autonaut/autoroute) [![Documentation](https://godoc.org/github.com/autonaut/autoroute?status.svg)](http://godoc.org/github.com/autonaut/autoroute)

------

WIP: There's a bit left to do here

- [ ] properly parse http Headers into the `autoroute.Header` arg
- [ ] document and test the `autoroute.Router`

^ If you don't mind those two being broken, here's autoroute:

build wicked fast, automatically documented APIs with Go.

Autoroute works by using reflection to automatically create an http.Handler from 
any number of Go functions. It does this by using a concept called an `autoroute.Codec`
which dictates how a certain HTTP `Content-Type` (parsed by `mime.ParseMediaType`) maps to 
a certain class of Go functions.

## Basic Example

In the general case, we can automatically create a JSON Handler from a function like this

```go
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
```

After we build and run the above program, we can try it out

```sh
ian@zuus ~ % curl -XPOST 'http://localhost:8080' -H 'Content-Type: application/json' -d '{"string":"test string split me"}' 
{"split_string":["test","string","split","me"]}
```

In addition to the happy path, our new handler automatically supports many common use cases, such as 

handling incorrect content type (http status code 415):

```sh
ian@zuus ~ % curl -w "%{http_code}" -XPOST 'http://localhost:8080' -H 'Content-Type: application/potatoes' -d '{"string":"test string split me"}'
415
```

limiting input size (this is harder to demo, but you can control it via `autoroute.WithMaxSizeBytes(your-max-byte-size-int64)` when you create a handler)


## Codecs and Roadmap

In addition to the default codec (`autoroute.JSONCodec`), we hope to ship many more codecs soon, such as 

- autoroute.CSVCodec (parse application/csv automatically)
- autoroute.FormCodec (parse application/x-www-form-urlencoded and multipart/form-data), outputting JSON
- autoroute.HTMLCodec ^ same as FormCodec, but output text/ html

We also plan to implement a mechanism for returning binary, custom file types (PDFs, Images, etc) 

LICENSE
======

MIT, see LICENSE