package datasnoop

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type RouteResolver func(*http.Request) string

func HTTP(handler http.Handler, route RouteResolver) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		resolvedRoute := request.URL.Path
		if route != nil && route(request) != "" {
			resolvedRoute = route(request)
		}
		start := time.Now()
		context, span := otel.Tracer("datasnoop/http").Start(request.Context(), request.Method+" "+resolvedRoute)
		defer span.End()
		response := &statusWriter{ResponseWriter: writer, status: http.StatusOK}
		handler.ServeHTTP(response, request.WithContext(context))
		span.SetAttributes(
			attribute.String("http.request.method", request.Method),
			attribute.String("http.route", resolvedRoute),
			attribute.Int("http.response.status_code", response.status),
			attribute.Int64("http.server.duration_ns", time.Since(start).Nanoseconds()),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (writer *statusWriter) WriteHeader(status int) {
	if writer.wrote {
		return
	}
	writer.status = status
	writer.wrote = true
	writer.ResponseWriter.WriteHeader(status)
}

func (writer *statusWriter) Write(payload []byte) (int, error) {
	if !writer.wrote {
		writer.WriteHeader(writer.status)
	}
	return writer.ResponseWriter.Write(payload)
}
