package service

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// Helper: read event from client channel with timeout
func readEvent(t *testing.T, client *SSEClient, timeout time.Duration) (SSEEvent, bool) {
	t.Helper()
	select {
	case data := <-client.Events:
		var event SSEEvent
		if err := json.Unmarshal(data, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		return event, true
	case <-time.After(timeout):
		return SSEEvent{}, false
	}
}

// Helper: assert no event received within timeout
func assertNoEvent(t *testing.T, client *SSEClient, timeout time.Duration) {
	t.Helper()
	select {
	case <-client.Events:
		t.Error("unexpected event received")
	case <-time.After(timeout):
		// OK - no event
	}
}

func TestSSEHub_NewHub(t *testing.T) {
	hub := NewSSEHub()
	if hub == nil {
		t.Fatal("NewSSEHub() returned nil")
	}
	if hub.clients == nil {
		t.Fatal("hub.clients should be initialized")
	}
	if len(hub.clients) != 0 {
		t.Fatalf("expected 0 clients, got %d", len(hub.clients))
	}
}

func TestSSEHub_Subscribe(t *testing.T) {
	hub := NewSSEHub()

	client := hub.Subscribe("user1")
	if client == nil {
		t.Fatal("Subscribe returned nil")
	}
	if client.UserID != "user1" {
		t.Fatalf("expected UserID 'user1', got '%s'", client.UserID)
	}
	if client.Events == nil {
		t.Fatal("client.Events channel should be initialized")
	}
	if cap(client.Events) != 32 {
		t.Fatalf("expected channel capacity 32, got %d", cap(client.Events))
	}
}

func TestSSEHub_SubscribeMultipleClients(t *testing.T) {
	hub := NewSSEHub()

	c1 := hub.Subscribe("user1")
	c2 := hub.Subscribe("user1") // Same user, two connections
	c3 := hub.Subscribe("user2")

	if c1 == c2 {
		t.Fatal("two subscriptions for same user should return different clients")
	}

	// Verify all clients receive events
	event := SSEEvent{Type: "test", Data: map[string]any{"msg": "hello"}}
	hub.BroadcastAll(event)

	_, ok1 := readEvent(t, c1, 100*time.Millisecond)
	if !ok1 {
		t.Error("c1 should receive broadcast event")
	}
	_, ok2 := readEvent(t, c2, 100*time.Millisecond)
	if !ok2 {
		t.Error("c2 should receive broadcast event")
	}
	_, ok3 := readEvent(t, c3, 100*time.Millisecond)
	if !ok3 {
		t.Error("c3 should receive broadcast event")
	}
}

func TestSSEHub_Unsubscribe(t *testing.T) {
	hub := NewSSEHub()

	client := hub.Subscribe("user1")
	hub.Unsubscribe(client)

	// After unsubscribe, client should not receive events
	event := SSEEvent{Type: "test", Data: map[string]any{"msg": "hello"}}
	hub.BroadcastToUser("user1", event)

	assertNoEvent(t, client, 50*time.Millisecond)
}

func TestSSEHub_Unsubscribe_Idempotent(t *testing.T) {
	hub := NewSSEHub()

	client := hub.Subscribe("user1")
	hub.Unsubscribe(client)
	// Second unsubscribe should not panic
	hub.Unsubscribe(client)
}

func TestSSEHub_Unsubscribe_UnknownClient(t *testing.T) {
	hub := NewSSEHub()

	// Unsubscribe a client that was never subscribed
	client := &SSEClient{
		UserID: "nonexistent",
		Events: make(chan []byte, 32),
	}
	// Should not panic
	hub.Unsubscribe(client)
}

func TestSSEHub_BroadcastToUser(t *testing.T) {
	hub := NewSSEHub()

	client1 := hub.Subscribe("user1")
	client2 := hub.Subscribe("user2")

	event := SSEEvent{
		Type: "approval_update",
		Data: map[string]any{"count": 5, "type": "leave"},
	}
	hub.BroadcastToUser("user1", event)

	// user1 should receive
	e, ok := readEvent(t, client1, 100*time.Millisecond)
	if !ok {
		t.Fatal("user1 should receive event")
	}
	if e.Type != "approval_update" {
		t.Fatalf("expected type 'approval_update', got '%s'", e.Type)
	}
	count, _ := e.Data["count"].(float64)
	if count != 5 {
		t.Fatalf("expected count 5, got %v", e.Data["count"])
	}

	// user2 should NOT receive
	assertNoEvent(t, client2, 50*time.Millisecond)
}

func TestSSEHub_BroadcastToUser_NoClients(t *testing.T) {
	hub := NewSSEHub()

	// Should not panic when broadcasting to user with no clients
	event := SSEEvent{Type: "test", Data: map[string]any{}}
	hub.BroadcastToUser("nonexistent", event)
}

func TestSSEHub_BroadcastToUsers(t *testing.T) {
	hub := NewSSEHub()

	c1 := hub.Subscribe("user1")
	c2 := hub.Subscribe("user2")
	c3 := hub.Subscribe("user3")

	event := SSEEvent{Type: "bulk", Data: map[string]any{"bulk": true}}
	hub.BroadcastToUsers([]string{"user1", "user3"}, event)

	_, ok1 := readEvent(t, c1, 100*time.Millisecond)
	if !ok1 {
		t.Error("user1 should receive")
	}
	// user2 should NOT receive
	assertNoEvent(t, c2, 50*time.Millisecond)
	_, ok3 := readEvent(t, c3, 100*time.Millisecond)
	if !ok3 {
		t.Error("user3 should receive")
	}
}

func TestSSEHub_BroadcastAll(t *testing.T) {
	hub := NewSSEHub()

	clients := make([]*SSEClient, 5)
	for i := 0; i < 5; i++ {
		clients[i] = hub.Subscribe("user")
	}

	event := SSEEvent{Type: "global", Data: map[string]any{}}
	hub.BroadcastAll(event)

	for i, c := range clients {
		_, ok := readEvent(t, c, 100*time.Millisecond)
		if !ok {
			t.Fatalf("client %d should receive BroadcastAll event", i)
		}
	}
}

func TestSSEHub_BroadcastAll_Empty(t *testing.T) {
	hub := NewSSEHub()

	// Should not panic when no clients connected
	event := SSEEvent{Type: "global", Data: map[string]any{}}
	hub.BroadcastAll(event)
}

func TestSSEHub_ChannelBufferOverflow(t *testing.T) {
	hub := NewSSEHub()

	client := hub.Subscribe("user1")

	// Fill the buffer (capacity is 32)
	event := SSEEvent{Type: "fill", Data: map[string]any{}}
	for i := 0; i < 35; i++ {
		hub.BroadcastToUser("user1", event) // Some events will be dropped
	}

	// Read events - should get at most 32 (buffer capacity), some may be dropped
	received := 0
	for {
		_, ok := readEvent(t, client, 50*time.Millisecond)
		if !ok {
			break
		}
		received++
	}

	if received > 32 {
		t.Fatalf("expected at most 32 events (buffer capacity), got %d", received)
	}
	if received == 0 {
		t.Error("expected at least some events to be received")
	}
}

func TestSSEHub_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	hub := NewSSEHub()
	var wg sync.WaitGroup

	// Concurrently subscribe/unsubscribe 50 times
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			uid := "user"
			client := hub.Subscribe(uid)
			time.Sleep(time.Millisecond)
			hub.Unsubscribe(client)
		}(i)
	}
	wg.Wait()

	// After all concurrent ops, broadcast should not panic
	event := SSEEvent{Type: "cleanup", Data: map[string]any{}}
	hub.BroadcastAll(event)
}

func TestSSEHub_ConcurrentBroadcastAndSubscribe(t *testing.T) {
	hub := NewSSEHub()
	var wg sync.WaitGroup

	// Subscribe clients first
	clients := make([]*SSEClient, 10)
	for i := 0; i < 10; i++ {
		clients[i] = hub.Subscribe("user")
	}

	// Concurrently broadcast and subscribe/unsubscribe
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := SSEEvent{Type: "concurrent", Data: map[string]any{}}
			hub.BroadcastAll(event)
		}()
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := hub.Subscribe("user")
			time.Sleep(time.Millisecond)
			hub.Unsubscribe(c)
		}(i)
	}

	wg.Wait()

	// Clean up original clients
	for _, c := range clients {
		hub.Unsubscribe(c)
	}

	// Verify hub is empty
	if len(hub.clients) != 0 {
		t.Fatalf("expected 0 users after cleanup, got %d", len(hub.clients))
	}
}

func TestSSEHub_GlobalSingleton(t *testing.T) {
	// Reset global
	GlobalSSEHub = nil

	hub := InitGlobalSSEHub()
	if hub == nil {
		t.Fatal("InitGlobalSSEHub returned nil")
	}

	got := GetSSEHub()
	if got != hub {
		t.Fatal("GetSSEHub should return the same instance")
	}

	if GlobalSSEHub != hub {
		t.Fatal("GlobalSSEHub should point to the same instance")
	}
}

func TestSSEHub_EventStructure(t *testing.T) {
	hub := NewSSEHub()
	client := hub.Subscribe("user")

	event := SSEEvent{
		Type: "approval_update",
		Data: map[string]any{
			"action": "approved",
			"count":  3,
			"nested": map[string]any{"key": "val"},
		},
	}
	hub.BroadcastToUser("user", event)

	e, ok := readEvent(t, client, 100*time.Millisecond)
	if !ok {
		t.Fatal("should receive event")
	}
	if e.Type != "approval_update" {
		t.Fatalf("expected type 'approval_update', got '%s'", e.Type)
	}
	if e.Data["action"] != "approved" {
		t.Fatalf("expected action 'approved', got '%v'", e.Data["action"])
	}
}
