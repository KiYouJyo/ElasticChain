package elastic

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type MessageStatus uint8

const (
	MessagePending MessageStatus = iota
	MessageFinalized
	MessageConsumed
)

// CrossDomainMessage is committed by the source domain and consumed exactly once
// by the destination after settlement finality.
type CrossDomainMessage struct {
	ID               [32]byte
	SourceDomain     DomainID
	DestinationDomain DomainID
	Nonce            uint64
	PayloadHash      [32]byte
	SettlementHeight uint64
	Status           MessageStatus
}

// MessageQueue is a prototype settlement-layer inbox/outbox registry.
type MessageQueue struct {
	messages map[[32]byte]CrossDomainMessage
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{messages: make(map[[32]byte]CrossDomainMessage)}
}

func MessageID(source, destination DomainID, nonce uint64, payload []byte) [32]byte {
	payloadHash := sha256.Sum256(payload)
	buf := make([]byte, 4+4+8+32)
	binary.BigEndian.PutUint32(buf[0:4], uint32(source))
	binary.BigEndian.PutUint32(buf[4:8], uint32(destination))
	binary.BigEndian.PutUint64(buf[8:16], nonce)
	copy(buf[16:], payloadHash[:])
	return sha256.Sum256(buf)
}

func (q *MessageQueue) Submit(source, destination DomainID, nonce uint64, payload []byte) (CrossDomainMessage, error) {
	if source == destination {
		return CrossDomainMessage{}, fmt.Errorf("cross-domain message source and destination must differ")
	}
	id := MessageID(source, destination, nonce, payload)
	if _, exists := q.messages[id]; exists {
		return CrossDomainMessage{}, fmt.Errorf("message %x already exists", id)
	}
	message := CrossDomainMessage{
		ID:                id,
		SourceDomain:      source,
		DestinationDomain: destination,
		Nonce:             nonce,
		PayloadHash:       sha256.Sum256(payload),
		Status:            MessagePending,
	}
	q.messages[id] = message
	return message, nil
}

// Finalize records the settlement block that made the source-domain commitment final.
func (q *MessageQueue) Finalize(id [32]byte, settlementHeight uint64) error {
	message, ok := q.messages[id]
	if !ok {
		return fmt.Errorf("message %x does not exist", id)
	}
	if message.Status != MessagePending {
		return fmt.Errorf("message %x is not pending", id)
	}
	message.SettlementHeight = settlementHeight
	message.Status = MessageFinalized
	q.messages[id] = message
	return nil
}

// Consume enforces destination binding, settlement finality depth and exactly-once delivery.
func (q *MessageQueue) Consume(id [32]byte, destination DomainID, finalizedHeight, minFinalityBlocks uint64) error {
	message, ok := q.messages[id]
	if !ok {
		return fmt.Errorf("message %x does not exist", id)
	}
	if message.DestinationDomain != destination {
		return fmt.Errorf("message %x belongs to destination %d, not %d", id, message.DestinationDomain, destination)
	}
	if message.Status == MessageConsumed {
		return fmt.Errorf("message %x was already consumed", id)
	}
	if message.Status != MessageFinalized {
		return fmt.Errorf("message %x is not finalized", id)
	}
	if finalizedHeight < message.SettlementHeight {
		return fmt.Errorf("finalized height %d precedes settlement height %d", finalizedHeight, message.SettlementHeight)
	}
	if finalizedHeight-message.SettlementHeight < minFinalityBlocks {
		return fmt.Errorf("message %x has insufficient settlement depth", id)
	}
	message.Status = MessageConsumed
	q.messages[id] = message
	return nil
}

func (q *MessageQueue) Get(id [32]byte) (CrossDomainMessage, bool) {
	message, ok := q.messages[id]
	return message, ok
}
