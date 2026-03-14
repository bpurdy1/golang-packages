package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/posthog/posthog-go"
)

type contextKey struct{}

// DistinctIDFunc extracts a PostHog distinct ID from the HTTP request.
// Return an empty string to skip event capture for that request.
type DistinctIDFunc func(r *http.Request) string

// PostHog is an HTTP middleware that captures request events to PostHog
// and makes the PostHog client available via request context.
type PostHog struct {
	client     posthog.Client
	distinctID DistinctIDFunc
	eventName  string
	properties func(r *http.Request) posthog.Properties
}

type Option func(*PostHog)

// WithDistinctID sets the function used to extract a distinct user ID from requests.
func WithDistinctID(fn DistinctIDFunc) Option {
	return func(p *PostHog) {
		p.distinctID = fn
	}
}

// WithEventName overrides the default event name ("$pageview").
func WithEventName(name string) Option {
	return func(p *PostHog) {
		p.eventName = name
	}
}

// WithProperties sets a function that returns extra properties to include on each captured event.
func WithProperties(fn func(r *http.Request) posthog.Properties) Option {
	return func(p *PostHog) {
		p.properties = fn
	}
}

// NewPostHog creates a PostHog middleware.
// The client is stored in request context for downstream handlers to use via ClientFromContext.
// If a DistinctIDFunc is provided, each request is automatically captured as an event.
func NewPostHog(client posthog.Client, opts ...Option) *PostHog {
	ph := &PostHog{
		client:    client,
		eventName: "$pageview",
	}
	for _, opt := range opts {
		opt(ph)
	}
	return ph
}

// ClientFromContext retrieves the PostHog client from the request context.
func ClientFromContext(ctx context.Context) posthog.Client {
	if c, ok := ctx.Value(contextKey{}).(posthog.Client); ok {
		return c
	}
	return nil
}

// Handler returns HTTP middleware that places the PostHog client in the request
// context and optionally captures request events after the response is written.
func (p *PostHog) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextKey{}, p.client)
		r = r.WithContext(ctx)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rw, r)

		if p.distinctID == nil {
			return
		}

		distinctID := p.distinctID(r)
		if distinctID == "" {
			return
		}

		props := posthog.NewProperties().
			Set("$current_url", r.URL.String()).
			Set("method", r.Method).
			Set("path", r.URL.Path).
			Set("status_code", rw.statusCode).
			Set("duration_ms", time.Since(start).Milliseconds())

		if ua := r.UserAgent(); ua != "" {
			props.Set("$user_agent", ua)
		}

		if p.properties != nil {
			for k, v := range p.properties(r) {
				props.Set(k, v)
			}
		}

		p.client.Enqueue(posthog.Capture{
			DistinctId: distinctID,
			Event:      p.eventName,
			Properties: props,
		})
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter, enabling http.ResponseController
// and interface assertions (e.g. http.Flusher) to work through the wrapper.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
