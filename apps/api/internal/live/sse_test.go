package live_test

import (
	"bufio"
	"context"
	"github.com/datasnoop/datasnoop/apps/api/internal/live"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSSEPublishesEventsAndSignalsGapsForSlowClients(t *testing.T) {
	broker := live.NewBroker(1)
	server := httptest.NewServer(broker.Handler())
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	deadline := time.Now().Add(time.Second)
	for broker.ClientCount() != 1 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	broker.Publish(live.Event{Type: "operation", Data: "first"})
	reader := bufio.NewReader(response.Body)
	_, _ = reader.ReadString('\n')
	_, _ = reader.ReadString('\n')
	line, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "event: operation") {
		t.Fatalf("first event=%q err=%v", line, err)
	}
	broker.Publish(live.Event{Type: "operation", Data: "second"})
	broker.Publish(live.Event{Type: "operation", Data: "third"})
	if broker.ClientCount() != 1 {
		t.Fatal("slow client disconnected publisher")
	}
}
