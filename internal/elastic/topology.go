package elastic

import (
	"encoding/binary"
	"fmt"
	"sort"
)

// Domain represents a node in the routing-prefix tree. Prefix contains the
// high-order PrefixBits of the first 64 bits of the routing hash, right-aligned.
type Domain struct {
	ID          DomainID
	Prefix      uint64
	PrefixBits  uint8
	ParentID    DomainID
	HasParent   bool
	LeftID      DomainID
	RightID     DomainID
	HasChildren bool
	Active      bool
}

// Topology stores the canonical execution-domain split tree.
type Topology struct {
	nodes        map[DomainID]Domain
	nextDomainID DomainID
}

func NewTopology() *Topology {
	root := Domain{ID: 0, Prefix: 0, PrefixBits: 0, Active: true}
	return &Topology{
		nodes:        map[DomainID]Domain{root.ID: root},
		nextDomainID: 1,
	}
}

func (t *Topology) Node(id DomainID) (Domain, bool) {
	d, ok := t.nodes[id]
	return d, ok
}

func (t *Topology) ActiveDomains() []DomainID {
	ids := make([]DomainID, 0, len(t.nodes))
	for id, domain := range t.nodes {
		if domain.Active {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (t *Topology) Split(id DomainID) (Domain, Domain, error) {
	parent, ok := t.nodes[id]
	if !ok {
		return Domain{}, Domain{}, fmt.Errorf("domain %d does not exist", id)
	}
	if !parent.Active {
		return Domain{}, Domain{}, fmt.Errorf("domain %d is not active", id)
	}
	if parent.PrefixBits >= 64 {
		return Domain{}, Domain{}, fmt.Errorf("domain %d cannot be split beyond 64 prefix bits", id)
	}
	if parent.HasChildren {
		return Domain{}, Domain{}, fmt.Errorf("domain %d already has child domains", id)
	}

	leftID := t.nextDomainID
	rightID := t.nextDomainID + 1
	t.nextDomainID += 2

	left := Domain{
		ID:         leftID,
		Prefix:     parent.Prefix << 1,
		PrefixBits: parent.PrefixBits + 1,
		ParentID:   parent.ID,
		HasParent:  true,
		Active:     true,
	}
	right := Domain{
		ID:         rightID,
		Prefix:     (parent.Prefix << 1) | 1,
		PrefixBits: parent.PrefixBits + 1,
		ParentID:   parent.ID,
		HasParent:  true,
		Active:     true,
	}

	parent.Active = false
	parent.HasChildren = true
	parent.LeftID = left.ID
	parent.RightID = right.ID
	t.nodes[parent.ID] = parent
	t.nodes[left.ID] = left
	t.nodes[right.ID] = right

	return left, right, nil
}

func (t *Topology) CanMerge(parentID DomainID) bool {
	parent, ok := t.nodes[parentID]
	if !ok || parent.Active || !parent.HasChildren {
		return false
	}
	left, leftOK := t.nodes[parent.LeftID]
	right, rightOK := t.nodes[parent.RightID]
	return leftOK && rightOK && left.Active && right.Active && !left.HasChildren && !right.HasChildren
}

func (t *Topology) Merge(parentID DomainID) error {
	if !t.CanMerge(parentID) {
		return fmt.Errorf("domain %d does not have two active leaf children", parentID)
	}
	parent := t.nodes[parentID]
	left := t.nodes[parent.LeftID]
	right := t.nodes[parent.RightID]

	left.Active = false
	right.Active = false
	parent.Active = true
	parent.HasChildren = false
	parent.LeftID = 0
	parent.RightID = 0

	t.nodes[left.ID] = left
	t.nodes[right.ID] = right
	t.nodes[parent.ID] = parent
	return nil
}

func (t *Topology) MergeableParents() []DomainID {
	ids := make([]DomainID, 0)
	for id := range t.nodes {
		if t.CanMerge(id) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Route maps a 32-byte routing hash to the active execution domain whose
// prefix contains it. Routing is deterministic and independent of node-local state.
func (t *Topology) Route(routingHash [32]byte) (DomainID, error) {
	key := binary.BigEndian.Uint64(routingHash[:8])
	var (
		bestID   DomainID
		bestBits uint8
		found    bool
	)

	for id, domain := range t.nodes {
		if !domain.Active || !prefixMatches(key, domain.Prefix, domain.PrefixBits) {
			continue
		}
		if !found || domain.PrefixBits > bestBits || (domain.PrefixBits == bestBits && id < bestID) {
			bestID = id
			bestBits = domain.PrefixBits
			found = true
		}
	}
	if !found {
		return 0, fmt.Errorf("no active domain matches routing hash")
	}
	return bestID, nil
}

func prefixMatches(key, prefix uint64, bits uint8) bool {
	if bits == 0 {
		return true
	}
	return key>>(64-bits) == prefix
}
