package workload

import "testing"

func TestSoakLifecycleUsesWorkloadExplainedBounds(t *testing.T) {
	if err := VerifySoak([]SoakSample{{100, 10, 2}, {120, 12, 3}}, 30, 3, 2); err != nil {
		t.Fatal(err)
	}
	if err := VerifySoak([]SoakSample{{100, 10, 2}, {200, 20, 5}}, 30, 3, 2); err == nil {
		t.Fatal("unbounded soak accepted")
	}
}
