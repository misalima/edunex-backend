package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
)

const defaultTicketTTL = 30 * time.Second

// TicketStore issues short-lived, single-use tickets that map to a userID.
// Used so the browser never puts a long-lived JWT in the WebSocket URL.
type TicketStore struct {
	mu      sync.Mutex
	tickets map[string]ticketEntry
	ttl     time.Duration
}

type ticketEntry struct {
	userID    uuid.UUID
	expiresAt time.Time
}

func NewTicketStore() *TicketStore {
	return &TicketStore{
		tickets: make(map[string]ticketEntry),
		ttl:     defaultTicketTTL,
	}
}

// Issue creates a new opaque ticket for userID. The ticket expires after ttl
// and can be consumed at most once.
func (s *TicketStore) Issue(userID uuid.UUID) (string, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	id := hex.EncodeToString(raw[:])

	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked(time.Now())
	s.tickets[id] = ticketEntry{
		userID:    userID,
		expiresAt: time.Now().Add(s.ttl),
	}
	return id, nil
}

// Consume validates and removes the ticket in one step (single-use).
// Returns the associated userID and true on success.
func (s *TicketStore) Consume(ticket string) (uuid.UUID, bool) {
	if ticket == "" {
		return uuid.Nil, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.tickets[ticket]
	if !ok {
		return uuid.Nil, false
	}
	delete(s.tickets, ticket)

	if time.Now().After(entry.expiresAt) {
		return uuid.Nil, false
	}
	return entry.userID, true
}

func (s *TicketStore) purgeExpiredLocked(now time.Time) {
	for id, entry := range s.tickets {
		if now.After(entry.expiresAt) {
			delete(s.tickets, id)
		}
	}
}
