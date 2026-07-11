package log

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestDebugOutput(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	l.Debug("debug msg", "k", "v")

	out := buf.String()
	if !strings.Contains(out, "debug msg") {
		t.Errorf("expected 'debug msg', got: %s", out)
	}
	if !strings.Contains(out, "v") {
		t.Errorf("expected 'v', got: %s", out)
	}
}

func TestInfoOutput(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, nil))
	l.Info("info msg", "user", "alice")

	out := buf.String()
	if !strings.Contains(out, "info msg") {
		t.Errorf("expected 'info msg', got: %s", out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected 'alice', got: %s", out)
	}
}

func TestWarnOutput(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, nil))
	l.Warn("warn msg", "threshold", 0.8)

	out := buf.String()
	if !strings.Contains(out, "warn msg") {
		t.Errorf("expected 'warn msg', got: %s", out)
	}
}

func TestErrorOutput(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, nil))
	l.Error("error msg", "err", "something broke")

	out := buf.String()
	if !strings.Contains(out, "error msg") {
		t.Errorf("expected 'error msg', got: %s", out)
	}
	if !strings.Contains(out, "something broke") {
		t.Errorf("expected 'something broke', got: %s", out)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	l.Debug("should not appear")
	l.Info("should not appear either")
	l.Warn("this should appear")

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Errorf("debug/info should be filtered at warn level")
	}
	if !strings.Contains(out, "this should appear") {
		t.Errorf("warn should appear at warn level")
	}
}

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

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
