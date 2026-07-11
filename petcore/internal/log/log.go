// Package log 提供结构化日志功能。
//
// 基于 Go 1.21 标准库 log/slog，提供统一的日志接口。
//
// 使用方式：
//
//	import "petcore/internal/log"
//
//	log.Info("引擎启动", "state", "ready")
//	log.Error("请求失败", "error", err, "req_id", id)
//
// 日志级别通过环境变量 LOG_LEVEL 控制（debug / info / warn / error）。
// 输出格式通过环境变量 LOG_FORMAT 控制（text / json）。
package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	initOnce sync.Once
	logger   *slog.Logger
)

// InitLogger 初始化全局日志器。
// 自动读取环境变量 LOG_LEVEL 和 LOG_FORMAT。
func InitLogger() {
	initOnce.Do(func() {
		level := parseLevel(os.Getenv("LOG_LEVEL"))
		format := os.Getenv("LOG_FORMAT")

		var handler slog.Handler
		switch strings.ToLower(format) {
		case "json":
			handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				Level: level,
			})
		default:
			handler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level:       level,
				ReplaceAttr: replaceAttr,
			})
		}

		logger = slog.New(handler)
	})
}

// L 返回全局日志器。
func L() *slog.Logger {
	InitLogger()
	return logger
}

// ─── 便捷函数 ──────────────────────────────

// Debug 输出一条调试级别日志。
func Debug(msg string, args ...any) { L().Debug(msg, args...) }

// Info 输出一条信息级别日志。
func Info(msg string, args ...any) { L().Info(msg, args...) }

// Warn 输出一条警告级别日志。
func Warn(msg string, args ...any) { L().Warn(msg, args...) }

// Error 输出一条错误级别日志。
func Error(msg string, args ...any) { L().Error(msg, args...) }

// ─── 带上下文的日志 ─────────────────────────

type ctxKey struct{}

// NewContext 将日志器注入到 context 中。
func NewContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext 从 context 中提取日志器，不存在时返回全局日志器。
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return L()
}

// ─── 内部 ──────────────────────────────────

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// replaceAttr 是 slog.HandlerOptions.ReplaceAttr 回调，
// 用于移除默认时间戳属性。
func replaceAttr(_ []string, a slog.Attr) slog.Attr {
	// 时间戳格式化为易读格式
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}
