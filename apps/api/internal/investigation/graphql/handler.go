// Package graphql exposes product-oriented investigation queries.
package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/investigation"
	graphql "github.com/graphql-go/graphql"
)

type endpointReader interface {
	Endpoints(context.Context, string, string, time.Time, time.Time) ([]investigation.EndpointSummary, error)
}
type Diagnostics struct {
	RestartID                              string
	Accepted, Rejected, Throttled, Dropped uint64
}
type diagnosticsReader interface {
	Diagnostics(context.Context) Diagnostics
}

func Handler(reader endpointReader) (http.Handler, error) {
	endpoint := graphql.NewObject(graphql.ObjectConfig{Name: "Endpoint", Fields: graphql.Fields{"route": {Type: graphql.String}, "requests": {Type: graphql.Int}, "errors": {Type: graphql.Int}, "averageDurationNs": {Type: graphql.Int}}})
	page := graphql.NewObject(graphql.ObjectConfig{Name: "EndpointPage", Fields: graphql.Fields{"items": {Type: graphql.NewList(endpoint)}, "nextCursor": {Type: graphql.String}}})
	diagnostic := graphql.NewObject(graphql.ObjectConfig{Name: "PlatformDiagnostics", Fields: graphql.Fields{"restartId": {Type: graphql.String}, "accepted": {Type: graphql.Int}, "rejected": {Type: graphql.Int}, "throttled": {Type: graphql.Int}, "dropped": {Type: graphql.Int}}})
	query := graphql.NewObject(graphql.ObjectConfig{Name: "Query", Fields: graphql.Fields{"endpoints": &graphql.Field{Type: page, Args: graphql.FieldConfigArgument{"service": {Type: graphql.NewNonNull(graphql.String)}, "environment": {Type: graphql.NewNonNull(graphql.String)}, "from": {Type: graphql.NewNonNull(graphql.String)}, "to": {Type: graphql.NewNonNull(graphql.String)}, "first": {Type: graphql.Int}, "after": {Type: graphql.String}}, Resolve: func(params graphql.ResolveParams) (any, error) {
		from, err := time.Parse(time.RFC3339, params.Args["from"].(string))
		if err != nil {
			return nil, err
		}
		to, err := time.Parse(time.RFC3339, params.Args["to"].(string))
		if err != nil {
			return nil, err
		}
		items, err := reader.Endpoints(params.Context, params.Args["service"].(string), params.Args["environment"].(string), from, to)
		if err != nil {
			return nil, err
		}
		start, limit := 0, 100
		if value, ok := params.Args["first"].(int); ok && value > 0 && value <= 100 {
			limit = value
		}
		if cursor, ok := params.Args["after"].(string); ok && cursor != "" {
			if _, err := fmt.Sscanf(cursor, "%d", &start); err != nil {
				return nil, fmt.Errorf("invalid cursor")
			}
		}
		end := start + limit
		if end > len(items) {
			end = len(items)
		}
		if start > len(items) {
			start = len(items)
		}
		result := make([]map[string]any, end-start)
		for i, item := range items[start:end] {
			result[i] = map[string]any{"route": item.Route, "requests": item.Requests, "errors": item.Errors, "averageDurationNs": item.AverageDuration.Nanoseconds()}
		}
		next := ""
		if end < len(items) {
			next = fmt.Sprintf("%d", end)
		}
		return map[string]any{"items": result, "nextCursor": next}, nil
	}}, "platformDiagnostics": &graphql.Field{Type: diagnostic, Resolve: func(params graphql.ResolveParams) (any, error) {
		if source, ok := reader.(diagnosticsReader); ok {
			value := source.Diagnostics(params.Context)
			return map[string]any{"restartId": value.RestartID, "accepted": value.Accepted, "rejected": value.Rejected, "throttled": value.Throttled, "dropped": value.Dropped}, nil
		}
		return map[string]any{"restartId": "unavailable"}, nil
	}}}})
	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: query})
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			http.Error(writer, "method must be POST", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, 1<<20)).Decode(&payload); err != nil {
			http.Error(writer, "invalid GraphQL request", http.StatusBadRequest)
			return
		}
		result := graphql.Do(graphql.Params{Schema: schema, RequestString: payload.Query, Context: request.Context()})
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(result)
	}), nil
}
