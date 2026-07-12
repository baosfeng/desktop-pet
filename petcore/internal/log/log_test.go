package log

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := parseLevel(tc.input); got != tc.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestLogFunctionsWithDirectLogger(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	l.Debug("debug msg", "key", "val")
	l.Info("info msg")
	l.Warn("warn msg")
	l.Error("error msg")

	output := buf.String()
	if !strContains(output, "debug msg") {
		t.Error("expected debug msg in output, got:", output)
	}
	if !strContains(output, "info msg") {
		t.Error("expected info msg in output")
	}
	if !strContains(output, "warn msg") {
		t.Error("expected warn msg in output")
	}
	if !strContains(output, "error msg") {
		t.Error("expected error msg in output")
	}
}

func TestNewContext_FromContext(t *testing.T) {
	customLogger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	ctx := NewContext(context.Background(), customLogger)

	got := FromContext(ctx)
	if got != customLogger {
		t.Error("FromContext should return the injected logger")
	}
}

func TestFromContext_NoLogger(t *testing.T) {
	got := FromContext(context.Background())
	if got == nil {
		t.Error("FromContext should not return nil")
	}
}

func strContains(s, substr string) bool {
	return len(s) >= len(substr) && strContainsStr(s, substr)
}

func strContainsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestReplaceAttr_RemovesTimeKey(t *testing.T) {
	attr := replaceAttr([]string{}, slog.Time(slog.TimeKey, time.Time{}))
	if attr.Key != "" {
		t.Errorf("expected empty key for TimeKey, got %q", attr.Key)
	}
}

func TestReplaceAttr_KeepsOtherKeys(t *testing.T) {
	attr := replaceAttr([]string{}, slog.String("msg", "hello"))
	if attr.Key != "msg" {
		t.Errorf("expected key='msg', got %q", attr.Key)
	}
	if attr.Value.String() != "hello" {
		t.Errorf("expected value='hello', got %q", attr.Value.String())
	}
}

func TestParseLevel_EnvVarIntegration(t *testing.T) {
	// Direct tests for the env-to-level mapping already covered by TestParseLevel
	// Verify the function handles all env values it can parse
	if got := parseLevel("debug"); got != slog.LevelDebug {
		t.Errorf("parseLevel('debug') = %v, want %v", got, slog.LevelDebug)
	}
	if got := parseLevel("info"); got != slog.LevelInfo {
		t.Errorf("parseLevel('info') = %v, want %v", got, slog.LevelInfo)
	}
	if got := parseLevel("warn"); got != slog.LevelWarn {
		t.Errorf("parseLevel('warn') = %v, want %v", got, slog.LevelWarn)
	}
	if got := parseLevel("error"); got != slog.LevelError {
		t.Errorf("parseLevel('error') = %v, want %v", got, slog.LevelError)
	}
}
