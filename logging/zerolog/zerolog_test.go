package zerologlogger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewLogger_DefaultLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithLevel("info"))

	logger.Info().Msg("hello")
	logger.Debug().Msg("should not appear")

	output := buf.String()
	if !strings.Contains(output, "hello") {
		t.Fatal("expected info message in output")
	}
	if strings.Contains(output, "should not appear") {
		t.Fatal("debug should be filtered at info level")
	}
}

func TestNewLogger_DebugLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithLevel("debug"))

	logger.Debug().Msg("debug msg")

	if !strings.Contains(buf.String(), "debug msg") {
		t.Fatal("expected debug message in output")
	}
}

func TestNewLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithLevel("info"))

	logger.Info().Msg("json test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid JSON output, got: %s", buf.String())
	}
	if m["message"] != "json test" {
		t.Errorf("message = %v, want %q", m["message"], "json test")
	}
	if m["level"] != "info" {
		t.Errorf("level = %v, want %q", m["level"], "info")
	}
}

func TestNewLogger_WithConsole(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithConsole(), WithLevel("info"))

	logger.Info().Msg("console test")

	output := buf.String()
	if !strings.Contains(output, "console test") {
		t.Fatal("expected message in console output")
	}
	// Console writer output is human-readable, not JSON
	var m map[string]any
	if err := json.Unmarshal([]byte(output), &m); err == nil {
		t.Fatal("expected non-JSON output from console writer")
	}
}

func TestNewLogger_CallerField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithLevel("info"))

	logger.Info().Msg("caller test")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid JSON: %s", buf.String())
	}
	caller, ok := m["caller"].(string)
	if !ok || caller == "" {
		t.Fatal("expected caller field in output")
	}
	if !strings.Contains(caller, ".go:") {
		t.Errorf("caller = %q, expected file:line format", caller)
	}
}

func TestNewLogger_ShortCaller(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithShortCaller(), WithLevel("info"))

	logger.Info().Msg("short caller")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid JSON: %s", buf.String())
	}
	caller := m["caller"].(string)
	if strings.Contains(caller, "/") {
		t.Errorf("short caller should not contain '/', got %q", caller)
	}
}

func TestNewLogger_FileBaseCaller(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithFileBaseCaller(), WithLevel("info"))

	logger.Info().Msg("full caller")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid JSON: %s", buf.String())
	}
	caller := m["caller"].(string)
	if !strings.Contains(caller, "/") {
		t.Errorf("file base caller should contain full path, got %q", caller)
	}
}

func TestFromContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithLevel("info"))

	ctx := WithContext(context.Background(), logger)
	l := FromContext(ctx)

	l.Info().Msg("from context")

	if !strings.Contains(buf.String(), "from context") {
		t.Fatal("expected message from context logger")
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithLevel("info"))

	ctx := WithContext(context.Background(), logger)
	ctx = WithFields(ctx, map[string]any{"user_id": 42, "action": "test"})

	FromContext(ctx).Info().Msg("with fields")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected valid JSON: %s", buf.String())
	}
	if m["user_id"] != float64(42) {
		t.Errorf("user_id = %v, want 42", m["user_id"])
	}
	if m["action"] != "test" {
		t.Errorf("action = %v, want %q", m["action"], "test")
	}
}

func TestNewLogger_InvalidLevel_DefaultsToInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WithWriter(&buf), WithLevel("invalid"))

	logger.Debug().Msg("should not appear")
	logger.Info().Msg("should appear")

	output := buf.String()
	if strings.Contains(output, "should not appear") {
		t.Fatal("invalid level should default to info, filtering debug")
	}
	if !strings.Contains(output, "should appear") {
		t.Fatal("expected info message with default level")
	}
}
