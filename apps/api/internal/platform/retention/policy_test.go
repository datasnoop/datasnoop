package retention

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPolicyValidationAndAuthorization(t *testing.T) {
	store := NewStore()
	if store.Get().Duration != Default {
		t.Fatal("default policy")
	}
	if _, err := store.Set(false, 7*24*time.Hour); err == nil {
		t.Fatal("unauthorized update")
	}
	if _, err := store.Set(true, time.Hour); err == nil {
		t.Fatal("unsupported policy")
	}
	if policy, err := store.Set(true, 7*24*time.Hour); err != nil || policy.Duration != 7*24*time.Hour {
		t.Fatalf("policy=%#v err=%v", policy, err)
	}
}

func TestPolicyHandlerExposesActivePolicy(t *testing.T) {
	response := httptest.NewRecorder()
	NewStore().Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/retention", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "720h0m0s") {
		t.Fatalf("response=%s", response.Body.String())
	}
}
