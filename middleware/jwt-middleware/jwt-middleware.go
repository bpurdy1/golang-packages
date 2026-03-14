package jwtmiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingToken  = errors.New("missing token")
	ErrInvalidScheme = errors.New("invalid token scheme")
	ErrInvalidToken  = errors.New("invalid token")
	ErrInvalidClaims = errors.New("invalid claims")
	ErrDecodeClaims  = errors.New("failed to decode claims")
)

type ctxKey[T any] struct{}

type Option func(*config)

type config struct {
	header string
	scheme string
}

// WithHeader sets the header to read the token from (default: "Authorization").
func WithHeader(header string) Option {
	return func(c *config) {
		c.header = header
	}
}

// WithScheme sets the expected scheme prefix (default: "Bearer").
// Set to "" to read the raw header value with no prefix stripping.
func WithScheme(scheme string) Option {
	return func(c *config) {
		c.scheme = scheme
	}
}

// JWTMiddleware returns middleware that parses a JWT from the request,
// validates it, decodes claims into a struct of type T, and stores
// the result in the request context.
//
// T must be a struct with `json` tags matching the JWT claim names.
// The signing key is used to validate the token signature.
//
// Usage:
//
//	type MyClaims struct {
//	    UserID string `json:"sub"`
//	    Email  string `json:"email"`
//	    Role   string `json:"role"`
//	}
//
//	mux.Use(jwtmiddleware.JWTMiddleware[MyClaims](signingKey))
func JWTMiddleware[T any](signingKey any, opts ...Option) func(http.Handler) http.Handler {
	cfg := &config{
		header: "Authorization",
		scheme: "Bearer",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get(cfg.header)
			if raw == "" {
				http.Error(w, ErrMissingToken.Error(), http.StatusUnauthorized)
				return
			}

			tokenStr := raw
			if cfg.scheme != "" {
				prefix := cfg.scheme + " "
				if !strings.HasPrefix(raw, prefix) {
					http.Error(w, ErrInvalidScheme.Error(), http.StatusUnauthorized)
					return
				}
				tokenStr = raw[len(prefix):]
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				return signingKey, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, ErrInvalidClaims.Error(), http.StatusUnauthorized)
				return
			}

			// Marshal claims map to JSON, then unmarshal into T.
			claimsJSON, err := json.Marshal(claims)
			if err != nil {
				http.Error(w, ErrDecodeClaims.Error(), http.StatusInternalServerError)
				return
			}

			var dst T
			if err := json.Unmarshal(claimsJSON, &dst); err != nil {
				http.Error(w, ErrDecodeClaims.Error(), http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKey[T]{}, dst)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext retrieves the decoded JWT claims struct from the request context.
// Returns the zero value and false if not present.
func FromContext[T any](ctx context.Context) (T, bool) {
	val, ok := ctx.Value(ctxKey[T]{}).(T)
	return val, ok
}
