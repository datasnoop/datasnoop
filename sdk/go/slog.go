package datasnoop

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type LogRecord struct {
	Timestamp  time.Time
	Severity   slog.Level
	Message    string
	Attributes []slog.Attr
	TraceID    string
	SpanID     string
}

type LogExporter func(context.Context, LogRecord)

func Slog(next slog.Handler, export LogExporter) slog.Handler {
	return &logHandler{next: next, export: export}
}

type logHandler struct {
	next   slog.Handler
	export LogExporter
	attrs  []slog.Attr
	groups []string
}

func (handler *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.next == nil || handler.next.Enabled(ctx, level)
}

func (handler *logHandler) Handle(ctx context.Context, record slog.Record) error {
	if handler.next != nil {
		if err := handler.next.Handle(ctx, record); err != nil {
			return err
		}
	}
	if handler.export == nil {
		return nil
	}
	attributes := append([]slog.Attr{}, handler.attrs...)
	record.Attrs(func(attribute slog.Attr) bool {
		attributes = append(attributes, attribute)
		return true
	})
	spanContext := trace.SpanContextFromContext(ctx)
	traceID, spanID := "", ""
	if spanContext.IsValid() {
		traceID, spanID = spanContext.TraceID().String(), spanContext.SpanID().String()
	}
	handler.export(ctx, LogRecord{Timestamp: record.Time, Severity: record.Level, Message: record.Message, Attributes: attributes, TraceID: traceID, SpanID: spanID})
	return nil
}

func (handler *logHandler) WithAttrs(attributes []slog.Attr) slog.Handler {
	copy := *handler
	copy.attrs = append(append([]slog.Attr{}, handler.attrs...), attributes...)
	if handler.next != nil {
		copy.next = handler.next.WithAttrs(attributes)
	}
	return &copy
}

func (handler *logHandler) WithGroup(name string) slog.Handler {
	copy := *handler
	copy.groups = append(append([]string{}, handler.groups...), name)
	if handler.next != nil {
		copy.next = handler.next.WithGroup(name)
	}
	return &copy
}
