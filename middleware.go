package autoroute

import (
	"net/http"

	"github.com/autonaut/autoroute/internal/keysigner"
)

// Middleware is used for things such as authentication / authorization controls
// checking specific headers of a request
// etc
type Middleware interface {
	// Before can modify an incoming request in the middleware chain
	Before(r *http.Request, h *Handler) error
	After(r *http.Request, w http.ResponseWriter) error
}

type MiddlewareError struct {
	StatusCode int
	Err        error
}

func (mwe MiddlewareError) Error() string {
	return mwe.Err.Error()
}

// SignedHeadersMiddleware validates that all incoming headers are signed using a certain key
// if they're set as a header outgoing, they'll also be signed on the way out.
// this works great for cookies.
type SignedHeadersMiddleware struct {
	ks      *keysigner.KeySigner
	headers []string
}

func NewSignedHeadersMiddleware(headers []string, key string) *SignedHeadersMiddleware {
	ks := keysigner.NewKeySigner(key)

	return &SignedHeadersMiddleware{
		headers: headers,
		ks:      ks,
	}
}

func (shm *SignedHeadersMiddleware) Before(r *http.Request, h *Handler) error {
	for _, h := range shm.headers {
		hVal := r.Header.Get(h)

		verified, err := shm.ks.Verify(hVal)
		if err != nil {
			return MiddlewareError{
				StatusCode: http.StatusForbidden,
				Err:        err,
			}
		}

		r.Header.Set(h, verified)
	}

	return nil
}

func (shm *SignedHeadersMiddleware) After(r *http.Request, w http.ResponseWriter) error {
	for _, h := range shm.headers {
		hVal := w.Header().Get(h)
		if hVal == "" {
			continue
		}

		signed, err := shm.ks.Sign(hVal)
		if err != nil {
			return err
		}

		r.Header.Set(h, signed)
	}

	return nil
}
