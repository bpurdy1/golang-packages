package headermiddleware

import (
	"context"
	"net/http"
	"reflect"
)

type ctxKey[T any] struct{}

// HeaderStructMiddleware returns middleware that decodes request headers into
// a struct of type T using the `header` struct tag, and stores the result
// in the request context.
//
// Usage:
//
//	type MyHeaders struct {
//	    Authorization string `header:"Authorization"`
//	    RequestID     string `header:"X-Request-Id"`
//	    RetryCount    int    `header:"X-Retry-Count"`
//	}
//
//	mux.Use(headermiddleware.HeaderStructMiddleware[MyHeaders]())
func HeaderStructMiddleware[T any]() func(http.Handler) http.Handler {
	var zero T
	t := reflect.TypeOf(&zero).Elem() // match DecodeHeaders' pointer → elem behavior

	// Prewarm the field cache used by DecodeHeaders
	getFieldInfo(t)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var dst T

			if err := DecodeHeaders(r.Header, &dst); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKey[T]{}, dst)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext retrieves the decoded header struct from the request context.
// Returns the zero value and false if not present.
func FromContext[T any](ctx context.Context) (T, bool) {
	val, ok := ctx.Value(ctxKey[T]{}).(T)
	return val, ok
}
