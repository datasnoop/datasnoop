package datasnoop_test

import (
	"context"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/sdk/go"
)

func TestBootstrapAcceptsMinimumConfiguration(t *testing.T) {
	client, err := datasnoop.Bootstrap(context.Background(), datasnoop.Config{ServiceName: "checkout", Endpoint: "http://127.0.0.1:4317", Token: "test-token"})
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	context, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(context); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestBootstrapRejectsInvalidConfiguration(t *testing.T) {
	for _, config := range []datasnoop.Config{
		{Endpoint: "http://127.0.0.1:4317", Token: "test-token"},
		{ServiceName: "checkout", Endpoint: "not an endpoint", Token: "test-token"},
		{ServiceName: "checkout", Endpoint: "http://127.0.0.1:4317"},
	} {
		if _, err := datasnoop.Bootstrap(context.Background(), config); err == nil {
			t.Fatalf("bootstrap accepted invalid config %#v", config)
		}
	}
}
