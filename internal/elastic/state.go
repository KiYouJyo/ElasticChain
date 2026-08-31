package elastic

import (
	"bytes"
	"fmt"
	"sort"
)

const snapshotVersion uint32 = 1

// TopologySnapshot is the canonical persistence/export boundary for execution
// topology state. Domains are always emitted in ascending DomainID order.
type TopologySnapshot struct {
	Version      uint32
	NextDomainID DomainID
	Domains      []Domain
}

func (t *Topology) Snapshot() TopologySnapshot {
	domains := make([]Domain, 0, len(t.nodes))
	for _, domain := range t.nodes {
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].ID < domains[j].ID })
	return TopologySnapshot{
		Version:      snapshotVersion,
		NextDomainID: t.nextDomainID,
		Domains:      domains,
	}
}

// RestoreTopology reconstructs topology state only after validating all
// consensus-relevant tree invariants. Malformed imported state must fail closed.
func RestoreTopology(snapshot TopologySnapshot) (*Topology, error) {
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}

	nodes := make(map[DomainID]Domain, len(snapshot.Domains))
	for _, domain := range snapshot.Domains {
		nodes[domain.ID] = domain
	}
	return &Topology{nodes: nodes, nextDomainID: snapshot.NextDomainID}, nil
}

func (s TopologySnapshot) Validate() error {
	if s.Version != snapshotVersion {
		return fmt.Errorf("unsupported topology snapshot version %d", s.Version)
	}
	if len(s.Domains) == 0 {
		return fmt.Errorf("topology snapshot has no domains")
	}
	if s.NextDomainID == 0 {
		return fmt.Errorf("next domain id must be positive")
	}

	nodes := make(map[DomainID]Domain, len(s.Domains))
	var maxID DomainID
	for _, domain := range s.Domains {
		if _, exists := nodes[domain.ID]; exists {
			return fmt.Errorf("duplicate domain id %d", domain.ID)
		}
		if domain.ID >= s.NextDomainID {
			return fmt.Errorf("domain id %d is not below next domain id %d", domain.ID, s.NextDomainID)
		}
		if domain.PrefixBits > 64 {
			return fmt.Errorf("domain %d has invalid prefix length %d", domain.ID, domain.PrefixBits)
		}
		if domain.PrefixBits < 64 && domain.Prefix >= (uint64(1)<<domain.PrefixBits) {
			return fmt.Errorf("domain %d prefix %d exceeds %d bits", domain.ID, domain.Prefix, domain.PrefixBits)
		}
		if domain.Active && domain.HasChildren {
			return fmt.Errorf("domain %d cannot be active and have current children", domain.ID)
		}
		nodes[domain.ID] = domain
		if domain.ID > maxID {
			maxID = domain.ID
		}
	}
	if s.NextDomainID <= maxID {
		return fmt.Errorf("next domain id %d must exceed max domain id %d", s.NextDomainID, maxID)
	}

	root, ok := nodes[0]
	if !ok {
		return fmt.Errorf("topology snapshot is missing root domain 0")
	}
	if root.HasParent || root.Prefix != 0 || root.PrefixBits != 0 {
		return fmt.Errorf("root domain has invalid parent or prefix metadata")
	}

	for _, domain := range nodes {
		if domain.ID != 0 && domain.HasParent {
			parent, ok := nodes[domain.ParentID]
			if !ok {
				return fmt.Errorf("domain %d references missing parent %d", domain.ID, domain.ParentID)
			}
			if domain.PrefixBits != parent.PrefixBits+1 {
				return fmt.Errorf("domain %d prefix length is not parent length + 1", domain.ID)
			}
			if domain.Prefix>>1 != parent.Prefix {
				return fmt.Errorf("domain %d prefix does not descend from parent %d", domain.ID, parent.ID)
			}
		}
		if domain.ID != 0 && !domain.HasParent {
			return fmt.Errorf("non-root domain %d has no parent", domain.ID)
		}
		if domain.HasChildren {
			if domain.LeftID == domain.RightID {
				return fmt.Errorf("domain %d child ids must differ", domain.ID)
			}
			left, leftOK := nodes[domain.LeftID]
			right, rightOK := nodes[domain.RightID]
			if !leftOK || !rightOK {
				return fmt.Errorf("domain %d references missing current children", domain.ID)
			}
			if !left.HasParent || left.ParentID != domain.ID || !right.HasParent || right.ParentID != domain.ID {
				return fmt.Errorf("domain %d child parent linkage is inconsistent", domain.ID)
			}
			if left.Prefix != domain.Prefix<<1 || right.Prefix != (domain.Prefix<<1)|1 {
				return fmt.Errorf("domain %d children do not form left/right prefix partition", domain.ID)
			}
		}
	}

	// Validate that the *current* routing tree rooted at domain 0 is complete:
	// every current node is either an active leaf or an inactive node with two
	// current children. Historical inactive children retained after a merge are
	// intentionally allowed to sit outside this current tree.
	visiting := make(map[DomainID]bool)
	var walk func(DomainID) error
	walk = func(id DomainID) error {
		if visiting[id] {
			return fmt.Errorf("cycle detected at domain %d", id)
		}
		visiting[id] = true
		defer delete(visiting, id)

		domain := nodes[id]
		if domain.Active {
			if domain.HasChildren {
				return fmt.Errorf("active domain %d has children", id)
			}
			return nil
		}
		if !domain.HasChildren {
			return fmt.Errorf("current routing domain %d is neither active nor split", id)
		}
		if err := walk(domain.LeftID); err != nil {
			return err
		}
		return walk(domain.RightID)
	}
	return walk(0)
}

// MessageQueueSnapshot is the canonical persistence/export boundary for
// cross-domain messages. Messages are emitted in lexicographic ID order.
type MessageQueueSnapshot struct {
	Version  uint32
	Messages []CrossDomainMessage
}

func (q *MessageQueue) Snapshot() MessageQueueSnapshot {
	messages := make([]CrossDomainMessage, 0, len(q.messages))
	for _, message := range q.messages {
		messages = append(messages, message)
	}
	sort.Slice(messages, func(i, j int) bool {
		return bytes.Compare(messages[i].ID[:], messages[j].ID[:]) < 0
	})
	return MessageQueueSnapshot{Version: snapshotVersion, Messages: messages}
}

// RestoreMessageQueue rebuilds the replay-protection index from canonical
// messages instead of persisting two independently mutable indexes.
func RestoreMessageQueue(snapshot MessageQueueSnapshot) (*MessageQueue, error) {
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}
	queue := NewMessageQueue()
	for _, message := range snapshot.Messages {
		queue.messages[message.ID] = message
		queue.sourceNonces[sourceNonceKey{Source: message.SourceDomain, Nonce: message.Nonce}] = message.ID
	}
	return queue, nil
}

func (s MessageQueueSnapshot) Validate() error {
	if s.Version != snapshotVersion {
		return fmt.Errorf("unsupported message snapshot version %d", s.Version)
	}
	ids := make(map[[32]byte]struct{}, len(s.Messages))
	nonces := make(map[sourceNonceKey][32]byte, len(s.Messages))
	for _, message := range s.Messages {
		if message.SourceDomain == message.DestinationDomain {
			return fmt.Errorf("message %x has identical source and destination %d", message.ID, message.SourceDomain)
		}
		if message.Status > MessageConsumed {
			return fmt.Errorf("message %x has unknown status %d", message.ID, message.Status)
		}
		if message.Status == MessagePending && message.SettlementHeight != 0 {
			return fmt.Errorf("pending message %x has settlement height %d", message.ID, message.SettlementHeight)
		}
		if _, exists := ids[message.ID]; exists {
			return fmt.Errorf("duplicate message id %x", message.ID)
		}
		ids[message.ID] = struct{}{}

		key := sourceNonceKey{Source: message.SourceDomain, Nonce: message.Nonce}
		if existingID, exists := nonces[key]; exists {
			return fmt.Errorf("source domain %d nonce %d is shared by %x and %x", message.SourceDomain, message.Nonce, existingID, message.ID)
		}
		nonces[key] = message.ID
	}
	return nil
}
