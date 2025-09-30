package logger

import (
	"context"
	"log/slog"
)

type slogHandler struct {
	l *Logger
}

func NewSlogHandler(l *Logger) slog.Handler {
	return &slogHandler{l: l}
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.toZlog(level) >= h.l.level
}

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	fields := F{}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	h.l.Log(h.toZlog(r.Level), r.Message, fields)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	f := F{}
	for _, a := range attrs {
		f[a.Key] = a.Value.Any()
	}
	return &slogHandler{l: h.l.WithFields(f)}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &groupHandler{parent: h, prefix: name}
}

type groupHandler struct {
	parent *slogHandler
	prefix string
}

func (g *groupHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return g.parent.Enabled(ctx, level)
}

func (g *groupHandler) Handle(ctx context.Context, r slog.Record) error {
	fields := F{}
	r.Attrs(func(a slog.Attr) bool {
		fields[g.prefix+"."+a.Key] = a.Value.Any()
		return true
	})
	g.parent.l.Log(g.parent.toZlog(r.Level), r.Message, fields)
	return nil
}

func (g *groupHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	f := F{}
	for _, a := range attrs {
		f[g.prefix+"."+a.Key] = a.Value.Any()
	}
	return &slogHandler{l: g.parent.l.WithFields(f)}
}

func (g *groupHandler) WithGroup(name string) slog.Handler {
	return &groupHandler{parent: g.parent, prefix: g.prefix + "." + name}
}

func (h *slogHandler) toZlog(level slog.Level) Level {
	switch {
	case level <= slog.LevelDebug:
		return LevelDebug
	case level < slog.LevelWarn:
		return LevelInfo
	case level < slog.LevelError:
		return LevelWarn
	case level < slog.LevelError+2:
		return LevelError
	default:
		return LevelFatal
	}
}
