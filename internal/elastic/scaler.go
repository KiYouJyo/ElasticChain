package elastic

import (
	"fmt"
	"sort"
)

type ActionKind string

const (
	ActionSplit ActionKind = "split"
	ActionMerge ActionKind = "merge"
)

// ScalingAction is a deterministic protocol proposal evaluated at epoch boundaries.
// For split actions DomainID is the active leaf to split. For merge actions DomainID
// is the inactive parent whose two active leaf children should be merged.
type ScalingAction struct {
	Kind     ActionKind
	DomainID DomainID
}

type pressureState struct {
	hotEpochs  uint32
	coldEpochs uint32
	lastEpoch  uint64
	seen       bool
}

// Planner keeps finalized pressure history. Its state must be persisted by the
// settlement layer in a production implementation; it is in-memory in this prototype.
type Planner struct {
	policy   ScalingPolicy
	pressure map[DomainID]pressureState
}

func NewPlanner(policy ScalingPolicy) (*Planner, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &Planner{policy: policy, pressure: make(map[DomainID]pressureState)}, nil
}

// Plan consumes exactly one finalized metric sample per active domain for an epoch.
// Scale-out takes priority over scale-in so overload cannot be masked by simultaneous
// consolidation. Results are sorted and therefore identical on every validator.
func (p *Planner) Plan(topology *Topology, metrics []DomainMetrics) ([]ScalingAction, error) {
	active := topology.ActiveDomains()
	if len(metrics) != len(active) {
		return nil, fmt.Errorf("received %d metric samples for %d active domains", len(metrics), len(active))
	}

	byID := make(map[DomainID]DomainMetrics, len(metrics))
	var epoch uint64
	for i, metric := range metrics {
		if err := metric.Validate(); err != nil {
			return nil, fmt.Errorf("domain %d: %w", metric.DomainID, err)
		}
		if i == 0 {
			epoch = metric.Epoch
		} else if metric.Epoch != epoch {
			return nil, fmt.Errorf("metrics span multiple epochs: %d and %d", epoch, metric.Epoch)
		}
		if _, exists := byID[metric.DomainID]; exists {
			return nil, fmt.Errorf("duplicate metrics for domain %d", metric.DomainID)
		}
		byID[metric.DomainID] = metric
	}

	for _, id := range active {
		metric, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("missing metrics for active domain %d", id)
		}
		state := p.pressure[id]
		if state.seen && metric.Epoch != state.lastEpoch+1 {
			return nil, fmt.Errorf("domain %d metric epoch jumped from %d to %d", id, state.lastEpoch, metric.Epoch)
		}

		hot := metric.GasUtilizationBps >= p.policy.ScaleOutUtilizationBps ||
			metric.DAUtilizationBps >= p.policy.ScaleOutUtilizationBps ||
			metric.CrossDomainQueueDepth >= p.policy.ScaleOutQueueDepth
		cold := metric.GasUtilizationBps <= p.policy.ScaleInUtilizationBps &&
			metric.DAUtilizationBps <= p.policy.ScaleInUtilizationBps &&
			metric.CrossDomainQueueDepth == 0

		switch {
		case hot:
			state.hotEpochs++
			state.coldEpochs = 0
		case cold:
			state.coldEpochs++
			state.hotEpochs = 0
		default:
			state.hotEpochs = 0
			state.coldEpochs = 0
		}
		state.lastEpoch = metric.Epoch
		state.seen = true
		p.pressure[id] = state
	}

	if uint32(len(active)) < p.policy.MaxDomains {
		remainingCapacity := int(p.policy.MaxDomains) - len(active)
		splits := make([]ScalingAction, 0)
		for _, id := range active {
			if remainingCapacity == 0 {
				break
			}
			if p.pressure[id].hotEpochs >= p.policy.ScaleOutConsecutiveEpochs {
				splits = append(splits, ScalingAction{Kind: ActionSplit, DomainID: id})
				remainingCapacity-- // one split replaces one active domain with two: net +1
			}
		}
		if len(splits) > 0 {
			return splits, nil
		}
	}

	if uint32(len(active)) <= p.policy.MinDomains {
		return nil, nil
	}

	parents := topology.MergeableParents()
	merges := make([]ScalingAction, 0)
	projectedActive := len(active)
	for _, parentID := range parents {
		if uint32(projectedActive) <= p.policy.MinDomains {
			break
		}
		parent, _ := topology.Node(parentID)
		leftState := p.pressure[parent.LeftID]
		rightState := p.pressure[parent.RightID]
		if leftState.coldEpochs >= p.policy.ScaleInConsecutiveEpochs &&
			rightState.coldEpochs >= p.policy.ScaleInConsecutiveEpochs {
			merges = append(merges, ScalingAction{Kind: ActionMerge, DomainID: parentID})
			projectedActive--
		}
	}

	sort.Slice(merges, func(i, j int) bool { return merges[i].DomainID < merges[j].DomainID })
	return merges, nil
}

// Apply executes previously planned actions at an epoch boundary. Applying actions
// is separated from planning so settlement can commit the decision before topology changes.
func (p *Planner) Apply(topology *Topology, actions []ScalingAction) error {
	for _, action := range actions {
		switch action.Kind {
		case ActionSplit:
			left, right, err := topology.Split(action.DomainID)
			if err != nil {
				return err
			}
			delete(p.pressure, action.DomainID)
			p.pressure[left.ID] = pressureState{}
			p.pressure[right.ID] = pressureState{}
		case ActionMerge:
			parent, ok := topology.Node(action.DomainID)
			if !ok {
				return fmt.Errorf("merge parent %d does not exist", action.DomainID)
			}
			if err := topology.Merge(action.DomainID); err != nil {
				return err
			}
			delete(p.pressure, parent.LeftID)
			delete(p.pressure, parent.RightID)
			p.pressure[parent.ID] = pressureState{}
		default:
			return fmt.Errorf("unknown scaling action %q", action.Kind)
		}
	}
	return nil
}
