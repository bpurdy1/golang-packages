package jwtmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type testClaims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

var testKey = []byte("test-secret-key")

func makeToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(testKey)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestJWTMiddleware(t *testing.T) {
	tokenStr := makeToken(t, jwt.MapClaims{
		"sub":   "user-123",
		"email": "test@example.com",
		"role":  "admin",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	handler := JWTMiddleware[testClaims](testKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := FromContext[testClaims](r.Context())
		if !ok {
			t.Fatal("expected claims in context")
		}
		if claims.UserID != "user-123" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
		}
		if claims.Email != "test@example.com" {
			t.Errorf("Email = %q, want %q", claims.Email, "test@example.com")
		}
		if claims.Role != "admin" {
			t.Errorf("Role = %q, want %q", claims.Role, "admin")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestJWTMiddleware_MissingToken(t *testing.T) {
	handler := JWTMiddleware[testClaims](testKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestJWTMiddleware_InvalidScheme(t *testing.T) {
	tokenStr := makeToken(t, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	handler := JWTMiddleware[testClaims](testKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic "+tokenStr)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	handler := JWTMiddleware[testClaims](testKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestJWTMiddleware_WrongKey(t *testing.T) {
	tokenStr := makeToken(t, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	wrongKey := []byte("wrong-key")
	handler := JWTMiddleware[testClaims](wrongKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	tokenStr := makeToken(t, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	handler := JWTMiddleware[testClaims](testKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestJWTMiddleware_CustomHeader(t *testing.T) {
	tokenStr := makeToken(t, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	handler := JWTMiddleware[testClaims](testKey, WithHeader("X-Token"), WithScheme(""))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := FromContext[testClaims](r.Context())
		if !ok {
			t.Fatal("expected claims in context")
		}
		if claims.UserID != "user-123" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Token", tokenStr)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestFromContext_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := FromContext[testClaims](req.Context())
	if ok {
		t.Fatal("expected ok=false when no middleware ran")
	}
}

// --- Benchmarks ---

func BenchmarkJWTMiddleware(b *testing.B) {
	claims := jwt.MapClaims{
		"sub":   "user-123",
		"email": "test@example.com",
		"role":  "admin",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(testKey)

	handler := JWTMiddleware[testClaims](testKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}
