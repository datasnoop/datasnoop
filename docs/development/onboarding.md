# Onboarding DataSnoop

## One-shot Go path

Start the API and database, then configure the Go SDK with service identity, endpoint, and credential. No manual provider or Collector configuration is required.

```go
client, err := datasnoop.Bootstrap(ctx, datasnoop.Config{
    ServiceName: "checkout",
    Endpoint: "http://localhost:4317",
    Token: "replace-me",
})
defer client.Shutdown(ctx)
```

Wrap the HTTP handler with `datasnoop.HTTP`, add `datasnoop.Slog` to the logger, and enable `datasnoop.NewHostCollector` for application-host measurements. Run `go run ./sdk/go/cmd/single-service-demo` to produce the documented incident dataset, then open the endpoint overview.

## Advanced direct-OTLP path

Any standard OTLP/gRPC exporter can send the supported logs, server operations, and host gauges directly to the endpoint. Set one `x-datasnoop-token` metadata value and use the documented resource attributes (`service.name`, optional deployment environment, and host identity). The receiver validates records individually and reports partial success for mixed batches.

The direct path does not require the DataSnoop Go SDK or an OpenTelemetry Collector. Use the correlated trace fixture and the direct-client contract test as a reference for the accepted subset.
