// Package retention owns validated telemetry retention policy.
package retention

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const Default = 30 * 24 * time.Hour

type Policy struct{ Duration time.Duration }

func Validate(duration time.Duration) (Policy, error) {
	if duration < 24*time.Hour || duration > 365*24*time.Hour {
		return Policy{}, errors.New("retention must be between 1 and 365 days")
	}
	return Policy{Duration: duration}, nil
}

type Store struct{ policy Policy }

func NewStore() *Store           { return &Store{policy: Policy{Duration: Default}} }
func (store *Store) Get() Policy { return store.policy }
func (store *Store) Set(authorised bool, duration time.Duration) (Policy, error) {
	if !authorised {
		return Policy{}, errors.New("retention update is not authorized")
	}
	policy, err := Validate(duration)
	if err != nil {
		return Policy{}, err
	}
	store.policy = policy
	return policy, nil
}

// Handler exposes the active policy without exposing storage implementation.
func (store *Store) Handler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(writer).Encode(map[string]string{"retention": store.Get().Duration.String()})
	})
}
