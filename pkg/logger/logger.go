package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(env, level string) zerolog.Logger {
	return NewWithWriter(env, level, os.Stderr)
}

func NewWithWriter(env, level string, w io.Writer) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	lvl := parseLevel(level)

	var out io.Writer = w
	if env != "production" && isTTY(w) {
		cw := zerolog.ConsoleWriter{
			Out:        w,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		cw.FormatLevel = formatLevel
		out = cw
	}

	return zerolog.New(out).
		Level(lvl).
		With().
		Timestamp().
		Logger()
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace":
		return zerolog.TraceLevel
	case "debug", "dbg":
		return zerolog.DebugLevel
	case "info", "inf":
		return zerolog.InfoLevel
	case "warn", "war", "warning":
		return zerolog.WarnLevel
	case "error", "err":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func formatLevel(i any) string {
	level, _ := i.(string)
	switch strings.ToLower(level) {
	case "debug":
		return "DBG"
	case "info":
		return "INF"
	case "warn":
		return "WAR"
	case "error":
		return "ERR"
	case "fatal":
		return "ERR"
	case "panic":
		return "ERR"
	default:
		return strings.ToUpper(level)
	}
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}
