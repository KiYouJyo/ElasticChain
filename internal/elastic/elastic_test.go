package elastic

import "testing"

func TestTopologySplitRouteAndMerge(t *testing.T) {
	topology := NewTopology()

	var low [32]byte
	var high [32]byte
	high[0] = 0x80

	if got, err := topology.Route(low); err != nil || got != 0 {
		t.Fatalf("root route = %d, %v; want domain 0", got, err)
	}

	left, right, err := topology.Split(0)
	if err != nil {
		t.Fatalf("split root: %v", err)
	}
	if got, err := topology.Route(low); err != nil || got != left.ID {
		t.Fatalf("low route = %d, %v; want %d", got, err, left.ID)
	}
	if got, err := topology.Route(high); err != nil || got != right.ID {
		t.Fatalf("high route = %d, %v; want %d", got, err, right.ID)
	}

	if !topology.CanMerge(0) {
		t.Fatal("root should be mergeable after one split")
	}
	if err := topology.Merge(0); err != nil {
		t.Fatalf("merge root: %v", err)
	}
	if got, err := topology.Route(high); err != nil || got != 0 {
		t.Fatalf("route after merge = %d, %v; want domain 0", got, err)
	}
}

func TestPlannerSplitThenMerge(t *testing.T) {
	policy := DefaultScalingPolicy()
	policy.MaxDomains = 4
	policy.ScaleOutConsecutiveEpochs = 2
	policy.ScaleInConsecutiveEpochs = 2

	planner, err := NewPlanner(policy)
	if err != nil {
		t.Fatalf("new planner: %v", err)
	}
	topology := NewTopology()

	hot := func(epoch uint64, id DomainID) DomainMetrics {
		return DomainMetrics{DomainID: id, Epoch: epoch, GasUtilizationBps: 9_000, DAUtilizationBps: 2_000}
	}
	cold := func(epoch uint64, id DomainID) DomainMetrics {
		return DomainMetrics{DomainID: id, Epoch: epoch, GasUtilizationBps: 500, DAUtilizationBps: 500}
	}

	if actions, err := planner.Plan(topology, []DomainMetrics{hot(1, 0)}); err != nil || len(actions) != 0 {
		t.Fatalf("epoch 1 actions = %#v, %v; want none", actions, err)
	}
	actions, err := planner.Plan(topology, []DomainMetrics{hot(2, 0)})
	if err != nil {
		t.Fatalf("epoch 2 plan: %v", err)
	}
	if len(actions) != 1 || actions[0].Kind != ActionSplit || actions[0].DomainID != 0 {
		t.Fatalf("epoch 2 actions = %#v; want split domain 0", actions)
	}
	if err := planner.Apply(topology, actions); err != nil {
		t.Fatalf("apply split: %v", err)
	}

	active := topology.ActiveDomains()
	if len(active) != 2 {
		t.Fatalf("active domain count = %d; want 2", len(active))
	}

	if actions, err := planner.Plan(topology, []DomainMetrics{cold(3, active[0]), cold(3, active[1])}); err != nil || len(actions) != 0 {
		t.Fatalf("epoch 3 actions = %#v, %v; want none", actions, err)
	}
	actions, err = planner.Plan(topology, []DomainMetrics{cold(4, active[0]), cold(4, active[1])})
	if err != nil {
		t.Fatalf("epoch 4 plan: %v", err)
	}
	if len(actions) != 1 || actions[0].Kind != ActionMerge || actions[0].DomainID != 0 {
		t.Fatalf("epoch 4 actions = %#v; want merge parent 0", actions)
	}
	if err := planner.Apply(topology, actions); err != nil {
		t.Fatalf("apply merge: %v", err)
	}
	if got := len(topology.ActiveDomains()); got != 1 {
		t.Fatalf("active domain count after merge = %d; want 1", got)
	}
}

func TestMessageRequiresFinalityAndConsumesExactlyOnce(t *testing.T) {
	queue := NewMessageQueue()
	message, err := queue.Submit(1, 2, 7, []byte("transfer:100"))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if err := queue.Consume(message.ID, 2, 100, 2); err == nil {
		t.Fatal("consume before finalization succeeded")
	}
	if err := queue.Finalize(message.ID, 100); err != nil {
		t.Fatalf("finalize: %v", err)
	}
	if err := queue.Consume(message.ID, 3, 102, 2); err == nil {
		t.Fatal("consume by wrong destination succeeded")
	}
	if err := queue.Consume(message.ID, 2, 101, 2); err == nil {
		t.Fatal("consume before minimum finality depth succeeded")
	}
	if err := queue.Consume(message.ID, 2, 102, 2); err != nil {
		t.Fatalf("consume at sufficient finality: %v", err)
	}
	if err := queue.Consume(message.ID, 2, 103, 2); err == nil {
		t.Fatal("duplicate consume succeeded")
	}
}

func TestMessageRejectsSourceNonceReuse(t *testing.T) {
	queue := NewMessageQueue()
	if _, err := queue.Submit(1, 2, 7, []byte("first")); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	if _, err := queue.Submit(1, 3, 7, []byte("different payload and destination")); err == nil {
		t.Fatal("reused source-domain nonce succeeded")
	}
	if _, err := queue.Submit(2, 3, 7, []byte("same numeric nonce, different source")); err != nil {
		t.Fatalf("nonce should be reusable by a different source domain: %v", err)
	}
}

func TestMessageIDIsDeterministic(t *testing.T) {
	first := MessageID(1, 2, 99, []byte("hello"))
	second := MessageID(1, 2, 99, []byte("hello"))
	if first != second {
		t.Fatal("same message inputs produced different IDs")
	}
	third := MessageID(1, 2, 100, []byte("hello"))
	if first == third {
		t.Fatal("different nonce produced the same message ID")
	}
}
