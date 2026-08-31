package elastic

import "fmt"

// BasisPoints is the denominator used for deterministic utilization ratios.
// Consensus-sensitive code must avoid floating point arithmetic.
const BasisPoints uint16 = 10_000

// DomainID identifies an execution domain in settlement state.
type DomainID uint32

// DomainMetrics contains only finalized, consensus-reproducible measurements.
// Local mempool size, wall-clock latency, CPU load and similar node-local data
// must never be used to trigger protocol state transitions.
type DomainMetrics struct {
	DomainID              DomainID
	Epoch                 uint64
	GasUtilizationBps     uint16
	DAUtilizationBps      uint16
	CrossDomainQueueDepth uint64
}

func (m DomainMetrics) Validate() error {
	if m.GasUtilizationBps > BasisPoints {
		return fmt.Errorf("gas utilization %d exceeds %d bps", m.GasUtilizationBps, BasisPoints)
	}
	if m.DAUtilizationBps > BasisPoints {
		return fmt.Errorf("DA utilization %d exceeds %d bps", m.DAUtilizationBps, BasisPoints)
	}
	return nil
}

// ScalingPolicy is intended to live in settlement-layer protocol state.
// All validators must evaluate the same policy against the same finalized data.
type ScalingPolicy struct {
	MinDomains                uint32
	MaxDomains                uint32
	ScaleOutUtilizationBps    uint16
	ScaleInUtilizationBps     uint16
	ScaleOutQueueDepth        uint64
	ScaleOutConsecutiveEpochs uint32
	ScaleInConsecutiveEpochs  uint32
}

func DefaultScalingPolicy() ScalingPolicy {
	return ScalingPolicy{
		MinDomains:                1,
		MaxDomains:                64,
		ScaleOutUtilizationBps:    8_000,
		ScaleInUtilizationBps:     1_500,
		ScaleOutQueueDepth:        1_000,
		ScaleOutConsecutiveEpochs: 100,
		ScaleInConsecutiveEpochs:  500,
	}
}

func (p ScalingPolicy) Validate() error {
	if p.MinDomains == 0 {
		return fmt.Errorf("min domains must be positive")
	}
	if p.MaxDomains < p.MinDomains {
		return fmt.Errorf("max domains %d is below min domains %d", p.MaxDomains, p.MinDomains)
	}
	if p.ScaleOutUtilizationBps > BasisPoints || p.ScaleInUtilizationBps > BasisPoints {
		return fmt.Errorf("utilization thresholds must be <= %d bps", BasisPoints)
	}
	if p.ScaleInUtilizationBps >= p.ScaleOutUtilizationBps {
		return fmt.Errorf("scale-in threshold must be below scale-out threshold")
	}
	if p.ScaleOutConsecutiveEpochs == 0 || p.ScaleInConsecutiveEpochs == 0 {
		return fmt.Errorf("consecutive epoch thresholds must be positive")
	}
	return nil
}
