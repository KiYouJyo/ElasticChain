package elastic

import "testing"

func TestTopologySnapshotRoundTripPreservesRoutingAndNextID(t *testing.T) {
	topology := NewTopology()
	left, right, err := topology.Split(0)
	if err != nil {
		t.Fatalf("split root: %v", err)
	}
	if _, _, err := topology.Split(right.ID); err != nil {
		t.Fatalf("split right: %v", err)
	}

	snapshot := topology.Snapshot()
	if err := snapshot.Validate(); err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}
	for i := 1; i < len(snapshot.Domains); i++ {
		if snapshot.Domains[i-1].ID >= snapshot.Domains[i].ID {
			t.Fatal("topology snapshot is not ordered by domain id")
		}
	}

	restored, err := RestoreTopology(snapshot)
	if err != nil {
		t.Fatalf("restore topology: %v", err)
	}

	var low, middle, high [32]byte
	middle[0] = 0x80
	high[0] = 0xC0
	for _, key := range [][32]byte{low, middle, high} {
		want, err := topology.Route(key)
		if err != nil {
			t.Fatalf("route original: %v", err)
		}
		got, err := restored.Route(key)
		if err != nil {
			t.Fatalf("route restored: %v", err)
		}
		if got != want {
			t.Fatalf("restored route %d, want %d", got, want)
		}
	}

	newLeft, _, err := restored.Split(left.ID)
	if err != nil {
		t.Fatalf("split restored topology: %v", err)
	}
	if newLeft.ID != snapshot.NextDomainID {
		t.Fatalf("new domain id %d, want persisted next id %d", newLeft.ID, snapshot.NextDomainID)
	}
}

func TestTopologySnapshotRoundTripAfterMergeAllowsHistoricalNodes(t *testing.T) {
	topology := NewTopology()
	if _, _, err := topology.Split(0); err != nil {
		t.Fatalf("split: %v", err)
	}
	if err := topology.Merge(0); err != nil {
		t.Fatalf("merge: %v", err)
	}

	snapshot := topology.Snapshot()
	if len(snapshot.Domains) != 3 {
		t.Fatalf("snapshot domains = %d, want root plus two historical children", len(snapshot.Domains))
	}
	if _, err := RestoreTopology(snapshot); err != nil {
		t.Fatalf("restore merged topology with historical nodes: %v", err)
	}
}

func TestTopologySnapshotRejectsMalformedCurrentTree(t *testing.T) {
	snapshot := NewTopology().Snapshot()
	snapshot.Domains[0].Active = false
	if err := snapshot.Validate(); err == nil {
		t.Fatal("inactive unsplit root snapshot validated")
	}

	snapshot = NewTopology().Snapshot()
	snapshot.Domains[0].HasChildren = true
	snapshot.Domains[0].LeftID = 1
	snapshot.Domains[0].RightID = 2
	if err := snapshot.Validate(); err == nil {
		t.Fatal("active root with children snapshot validated")
	}
}

func TestMessageSnapshotRoundTripPreservesReplayProtection(t *testing.T) {
	queue := NewMessageQueue()
	first, err := queue.Submit(1, 2, 7, []byte("first"))
	if err != nil {
		t.Fatalf("submit first: %v", err)
	}
	if err := queue.Finalize(first.ID, 10); err != nil {
		t.Fatalf("finalize first: %v", err)
	}
	second, err := queue.Submit(2, 3, 3, []byte("second"))
	if err != nil {
		t.Fatalf("submit second: %v", err)
	}

	snapshot := queue.Snapshot()
	if err := snapshot.Validate(); err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}
	for i := 1; i < len(snapshot.Messages); i++ {
		if string(snapshot.Messages[i-1].ID[:]) >= string(snapshot.Messages[i].ID[:]) {
			t.Fatal("message snapshot is not ordered by message id")
		}
	}

	restored, err := RestoreMessageQueue(snapshot)
	if err != nil {
		t.Fatalf("restore messages: %v", err)
	}
	if got, ok := restored.Get(first.ID); !ok || got.Status != MessageFinalized {
		t.Fatalf("restored first message = %#v, %v", got, ok)
	}
	if got, ok := restored.Get(second.ID); !ok || got.Status != MessagePending {
		t.Fatalf("restored second message = %#v, %v", got, ok)
	}
	if _, err := restored.Submit(1, 3, 7, []byte("replay")); err == nil {
		t.Fatal("restored queue accepted reused source nonce")
	}
}

func TestMessageSnapshotRejectsDuplicateSourceNonce(t *testing.T) {
	queue := NewMessageQueue()
	first, err := queue.Submit(1, 2, 9, []byte("one"))
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	second := first
	second.ID[0] ^= 0xFF
	second.DestinationDomain = 3

	snapshot := MessageQueueSnapshot{
		Version:  snapshotVersion,
		Messages: []CrossDomainMessage{first, second},
	}
	if err := snapshot.Validate(); err == nil {
		t.Fatal("duplicate source nonce snapshot validated")
	}
}
