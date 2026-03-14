package requestidmiddleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey struct{}

func RequestIdMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uuid, _ := uuid.NewV7()
			ctx := context.WithValue(r.Context(), ctxKey{}, uuid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(ctx context.Context) uuid.UUID {
	return ctx.Value(ctxKey{}).(uuid.UUID)
}
