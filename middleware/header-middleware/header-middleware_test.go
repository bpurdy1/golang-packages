package headermiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testHeaders struct {
	Authorization string  `header:"Authorization"`
	RequestID     string  `header:"X-Request-Id"`
	RetryCount    int     `header:"X-Retry-Count"`
	Score         float64 `header:"X-Score"`
	Debug         bool    `header:"X-Debug"`
	Ignored       string  // no tag — should be ignored
	Skipped       string  `header:"-"`
}

func TestHeaderStructMiddleware(t *testing.T) {
	handler := HeaderStructMiddleware[testHeaders]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := FromContext[testHeaders](r.Context())
		if !ok {
			t.Fatal("expected header struct in context")
		}
		if h.Authorization != "Bearer tok" {
			t.Errorf("Authorization = %q, want %q", h.Authorization, "Bearer tok")
		}
		if h.RequestID != "abc-123" {
			t.Errorf("RequestID = %q, want %q", h.RequestID, "abc-123")
		}
		if h.RetryCount != 3 {
			t.Errorf("RetryCount = %d, want 3", h.RetryCount)
		}
		if h.Score != 9.5 {
			t.Errorf("Score = %f, want 9.5", h.Score)
		}
		if !h.Debug {
			t.Error("Debug = false, want true")
		}
		if h.Ignored != "" {
			t.Errorf("Ignored = %q, want empty", h.Ignored)
		}
		if h.Skipped != "" {
			t.Errorf("Skipped = %q, want empty", h.Skipped)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	req.Header.Set("X-Request-Id", "abc-123")
	req.Header.Set("X-Retry-Count", "3")
	req.Header.Set("X-Score", "9.5")
	req.Header.Set("X-Debug", "true")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestHeaderStructMiddleware_MissingHeaders(t *testing.T) {
	handler := HeaderStructMiddleware[testHeaders]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := FromContext[testHeaders](r.Context())
		if !ok {
			t.Fatal("expected header struct in context")
		}
		if h.Authorization != "" {
			t.Errorf("Authorization = %q, want empty", h.Authorization)
		}
		if h.RetryCount != 0 {
			t.Errorf("RetryCount = %d, want 0", h.RetryCount)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestHeaderStructMiddleware_InvalidInt(t *testing.T) {
	handler := HeaderStructMiddleware[testHeaders]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on decode error")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Retry-Count", "not-a-number")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestHeaderStructMiddleware_InvalidBool(t *testing.T) {
	handler := HeaderStructMiddleware[testHeaders]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on decode error")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Debug", "not-a-bool")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestFromContext_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := FromContext[testHeaders](req.Context())
	if ok {
		t.Fatal("expected ok=false when no middleware ran")
	}
}

type uintHeaders struct {
	Port uint16 `header:"X-Port"`
}

func TestHeaderStructMiddleware_Uint(t *testing.T) {
	handler := HeaderStructMiddleware[uintHeaders]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, ok := FromContext[uintHeaders](r.Context())
		if !ok {
			t.Fatal("expected header struct in context")
		}
		if h.Port != 8080 {
			t.Errorf("Port = %d, want 8080", h.Port)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Port", "8080")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

// --- Benchmarks ---

type benchHeaders struct {
	Authorization string `header:"Authorization"`
	RequestID     string `header:"X-Request-Id"`
	ContentType   string `header:"Content-Type"`
	RetryCount    int    `header:"X-Retry-Count"`
	Debug         bool   `header:"X-Debug"`
}

func BenchmarkHeaderStructMiddleware(b *testing.B) {
	middleware := HeaderStructMiddleware[benchHeaders]()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer some-long-token-value-here")
	req.Header.Set("X-Request-Id", "550e8400-e29b-41d4-a716-446655440000")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Retry-Count", "2")
	req.Header.Set("X-Debug", "true")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkDecode(b *testing.B) {
	header := make(http.Header)
	header.Set("Authorization", "Bearer some-long-token-value-here")
	header.Set("X-Request-Id", "550e8400-e29b-41d4-a716-446655440000")
	header.Set("Content-Type", "application/json")
	header.Set("X-Retry-Count", "2")
	header.Set("X-Debug", "true")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var dst benchHeaders
		_ = DecodeHeaders(header, &dst)
	}
}

func BenchmarkFromContext(b *testing.B) {
	middleware := HeaderStructMiddleware[benchHeaders]()
	var captured *http.Request
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer tok")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	ctx := captured.Context()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromContext[benchHeaders](ctx)
	}
}
