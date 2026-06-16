package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/KENTA0326/run-sync-pro/internal/config"
)

// InitFromEnv は LOG_LEVEL / LOG_FORMAT / LOG_SOURCE で slog.Default を初期化する。
// main の先頭（database.Connect より前）で一度呼ぶ。
//
//	LOG_LEVEL   … debug | info | warn | error（既定: info）
//	LOG_FORMAT  … json | text（既定: json）
//	LOG_SOURCE  … true でソース位置を付与（既定: false）
func InitFromEnv() {
	level := parseLevel(config.ResolveString("", "LOG_LEVEL", "info"))
	format := strings.ToLower(strings.TrimSpace(config.ResolveString("", "LOG_FORMAT", "json")))
	addSource := parseBool(config.ResolveString("", "LOG_SOURCE", "false"))

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	}
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(h))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func parseBool(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "1" || s == "true" || s == "yes"
}
