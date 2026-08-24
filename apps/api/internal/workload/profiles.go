// Package workload defines deterministic system workload profiles and their semantic oracle.
package workload

import "fmt"

type Profile struct {
	Name       string
	Operations int
	Errors     int
	Retries    int
	Logs       int
	Metrics    int
}

var Profiles = []Profile{
	{Name: "smoke-v1", Operations: 3, Errors: 2, Retries: 1, Logs: 2, Metrics: 3},
	{Name: "steady-v1", Operations: 60, Errors: 12, Retries: 3, Logs: 12, Metrics: 30},
	{Name: "burst-v1", Operations: 240, Errors: 48, Retries: 12, Logs: 48, Metrics: 30},
	{Name: "saturation-v1", Operations: 400, Errors: 80, Retries: 20, Logs: 80, Metrics: 30},
	{Name: "concurrent-retention-v1", Operations: 120, Errors: 24, Retries: 6, Logs: 24, Metrics: 30},
	{Name: "recovery-v1", Operations: 80, Errors: 16, Retries: 8, Logs: 16, Metrics: 30},
	{Name: "soak-v1", Operations: 600, Errors: 120, Retries: 30, Logs: 120, Metrics: 300},
}

func Oracle(profile Profile) error {
	if profile.Operations < 1 || profile.Errors < 0 || profile.Errors > profile.Operations || profile.Retries < 0 || profile.Logs < profile.Errors || profile.Metrics < 3 {
		return fmt.Errorf("invalid workload profile %q", profile.Name)
	}
	return nil
}
