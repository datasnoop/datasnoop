package datasnoop

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

type Config struct {
	ServiceName string
	Endpoint    string
	Token       string
}

type Client struct{ provider *sdktrace.TracerProvider }

func Bootstrap(ctx context.Context, config Config) (*Client, error) {
	endpoint, err := validateConfig(config)
	if err != nil {
		return nil, err
	}
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure(), otlptracegrpc.WithHeaders(map[string]string{"x-datasnoop-token": config.Token}))
	if err != nil {
		return nil, fmtError("create telemetry exporter", err)
	}
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(resource.NewWithAttributes("", semconv.ServiceName(config.ServiceName))))
	otel.SetTracerProvider(provider)
	return &Client{provider: provider}, nil
}

func (client *Client) Shutdown(ctx context.Context) error {
	if client == nil || client.provider == nil {
		return nil
	}
	return client.provider.Shutdown(ctx)
}

func validateConfig(config Config) (string, error) {
	if strings.TrimSpace(config.ServiceName) == "" {
		return "", errors.New("DataSnoop service name is required")
	}
	if strings.TrimSpace(config.Token) == "" {
		return "", errors.New("DataSnoop credential is required")
	}
	parsed, err := url.Parse(config.Endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("DataSnoop endpoint must be an absolute http or https URL")
	}
	return parsed.Host, nil
}

func fmtError(prefix string, err error) error { return errors.New(prefix + ": " + err.Error()) }
