package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/autonaut/autoroute"
)

type LoginResponse struct {
	APIKey string
}

type SpecialResponse struct {
	Restricted bool
}

// DoSomethingSpecial is some restricted access.
// Hypothetically, we could add autoroute.Header as the second argument to this function
// and then look at the actual x-api-key value itself!
func DoSomethingSpecial(ctx context.Context) *SpecialResponse {
	return &SpecialResponse{
		Restricted: true,
	}
}

type apiKeyVerifier struct{}

func (akv *apiKeyVerifier) Before(r *http.Request, h *autoroute.Handler) error {
	val := r.Header.Get("x-api-key")
	if val != "hurray!" {
		return autoroute.MiddlewareError{
			StatusCode: http.StatusForbidden,
			Err:        errors.New("invalid api key"),
		}
	}

	return nil
}

func main() {
	signedHeaderMiddleware := autoroute.NewSignedHeadersMiddleware([]string{"x-api-key"}, "test-key")

	login := func(ctx context.Context, input struct {
		Username string
		Password string
	}) (*LoginResponse, error) {
		// use the middleware directly in our login handler so the value is nice and signed going out
		signedAPIKey, err := signedHeaderMiddleware.Sign("hurray!")
		if err != nil {
			return nil, err
		}

		return &LoginResponse{
			APIKey: signedAPIKey,
		}, nil
	}

	// applies the JSONCodec to all sub routes
	r, err := autoroute.NewRouter(autoroute.WithCodec(autoroute.JSONCodec))
	if err != nil {
		log.Fatal(err)
	}

	// register our two routes
	r.Register(http.MethodPost, "/login", login)
	r.Register(http.MethodPost, "/special", DoSomethingSpecial,
		// signed header middleware runs first and ensures the api key is signed for this route
		autoroute.WithMiddleware(signedHeaderMiddleware),
		// apikeyVerifier gets the clean, unsigned version of this header
		autoroute.WithMiddleware(&apiKeyVerifier{}),
	)

	log.Fatal(http.ListenAndServe(":8080", r))
}

// here's the curl guide to how this works
// ian@zuus ~ % curl -XPOST localhost:8080/login -H 'Content-Type: application/json' -d '{"Username": "ian", "Password": "password"}'
// {"APIKey":"hurray!.6fe7c33d56bab58e99b494c558720aae045a1fca2cf5902a3bf8ea6a1ae623ae"}

// with no header sent we get rejected
// ian@zuus ~ % curl -XPOST localhost:8080/special
// {"error":"invalid token"}

// with an unsigned header we get rejected
// ian@zuus ~ % curl -XPOST localhost:8080/special -H 'x-api-key: hurray!'
// {"error":"invalid token"}

// ian@zuus ~ % curl -XPOST localhost:8080/special -H 'Content-Type: application/json' -H 'x-api-key: hurray!.6fe7c33d56bab58e99b494c558720aae045a1fca2cf5902a3bf8ea6a1ae623ae'
// {"Restricted":true}
