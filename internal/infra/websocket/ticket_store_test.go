package websocket

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTicketStore_IssueAndConsume(t *testing.T) {
	store := NewTicketStore()
	userID := uuid.New()

	ticket, err := store.Issue(userID)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if ticket == "" {
		t.Fatal("expected non-empty ticket")
	}

	got, ok := store.Consume(ticket)
	if !ok {
		t.Fatal("expected consume to succeed")
	}
	if got != userID {
		t.Fatalf("userID: got %s want %s", got, userID)
	}

	if _, ok := store.Consume(ticket); ok {
		t.Fatal("ticket must be single-use")
	}
}

func TestTicketStore_Expired(t *testing.T) {
	store := NewTicketStore()
	store.ttl = time.Millisecond

	ticket, err := store.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	time.Sleep(5 * time.Millisecond)

	if _, ok := store.Consume(ticket); ok {
		t.Fatal("expired ticket must be rejected")
	}
}

func TestTicketStore_Unknown(t *testing.T) {
	store := NewTicketStore()
	if _, ok := store.Consume("does-not-exist"); ok {
		t.Fatal("unknown ticket must be rejected")
	}
}
