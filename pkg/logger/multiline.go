package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
)

type MultilineHandler struct {
	out  io.Writer
	opts *slog.HandlerOptions
}

func NewMultilineHandler(out io.Writer, opts *slog.HandlerOptions) *MultilineHandler {
	return &MultilineHandler{out: out, opts: opts}
}

func (h *MultilineHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *MultilineHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MultilineHandler{
		out:  h.out,
		opts: h.opts,
	}
}

func (h *MultilineHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *MultilineHandler) Handle(ctx context.Context, record slog.Record) error {
	_, file, line, ok := runtime.Caller(4)
	caller := "unknown:0"
	if ok {
		parts := strings.Split(file, "/")
		caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
	}

	ts := record.Time.Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] %-5s [%s]", ts, record.Level.String(), caller)

	msg := record.Message
	lines := strings.Split(msg, "\n")
	var b strings.Builder
	if len(lines) > 0 {
		fmt.Fprintf(&b, "%s %s\n", prefix, lines[0])
		for _, extra := range lines[1:] {
			if extra != "" {
				fmt.Fprintf(&b, "      %s\n", extra)
			}
		}
	}

	_, err := h.out.Write([]byte(b.String()))
	return err
}
