package workload

import "fmt"

type SoakSample struct {
	MemoryBytes uint64
	Goroutines  int
	Connections int
}

func VerifySoak(samples []SoakSample, memoryGrowth uint64, goroutineGrowth, connectionGrowth int) error {
	if len(samples) < 2 {
		return fmt.Errorf("soak requires at least two samples")
	}
	first, last := samples[0], samples[len(samples)-1]
	if last.MemoryBytes > first.MemoryBytes+memoryGrowth || last.Goroutines > first.Goroutines+goroutineGrowth || last.Connections > first.Connections+connectionGrowth {
		return fmt.Errorf("soak lifecycle exceeded workload-explained bound")
	}
	return nil
}
