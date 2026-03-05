package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNewHub tests creating a new Hub
func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("Expected hub to be created")
	}

	if hub.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if hub.broadcast == nil {
		t.Error("Expected broadcast channel to be initialized")
	}

	if hub.register == nil {
		t.Error("Expected register channel to be initialized")
	}

	if hub.unregister == nil {
		t.Error("Expected unregister channel to be initialized")
	}
}

// TestHub_StartStop tests starting and stopping the Hub
func TestHub_StartStop(t *testing.T) {
	hub := NewHub()

	// Start hub in goroutine
	go hub.Run()

	// Give it time to start
	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	running := hub.running
	hub.mu.RUnlock()

	if !running {
		t.Error("Expected hub to be running")
	}

	// Stop hub
	hub.Stop()

	hub.mu.RLock()
	running = hub.running
	hub.mu.RUnlock()

	if running {
		t.Error("Expected hub to not be running after Stop")
	}
}

// TestHub_RegisterUnregister tests client registration
func TestHub_RegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// Create mock client
	client := &Client{
		ID:            "test-client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	// Register client
	hub.Register(client)

	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	count := len(hub.clients)
	hub.mu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}

	// Unregister client
	hub.Unregister(client)

	time.Sleep(10 * time.Millisecond)

	hub.mu.RLock()
	count = len(hub.clients)
	hub.mu.RUnlock()

	if count != 0 {
		t.Errorf("Expected 0 clients after unregister, got %d", count)
	}
}

// TestHub_Broadcast tests broadcasting messages
func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// Create client subscribed to "traffic" channel
	client := &Client{
		ID:            "test-client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"traffic": true},
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Broadcast message
	msg := &Message{
		Type:      "data",
		Channel:   "traffic",
		Data:      map[string]string{"url": "https://example.com"},
		Timestamp: time.Now(),
	}

	hub.Broadcast(msg)

	// Wait for message to be sent
	select {
	case data := <-client.send:
		var received Message
		if err := json.Unmarshal(data, &received); err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}
		if received.Channel != "traffic" {
			t.Errorf("Expected channel 'traffic', got '%s'", received.Channel)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive message, but got none")
	}
}

// TestHub_BroadcastFiltered tests that messages are filtered by subscription
func TestHub_BroadcastFiltered(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// Create client subscribed to "vulns" channel only
	client := &Client{
		ID:            "test-client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: map[string]bool{"vulns": true},
	}

	hub.Register(client)
	time.Sleep(10 * time.Millisecond)

	// Broadcast message to "traffic" channel (client not subscribed)
	msg := &Message{
		Type:      "data",
		Channel:   "traffic",
		Data:      map[string]string{"url": "https://example.com"},
		Timestamp: time.Now(),
	}

	hub.Broadcast(msg)

	// Should not receive message
	select {
	case <-client.send:
		t.Error("Should not receive message for unsubscribed channel")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message received
	}
}

// TestClient_SubscribeUnsubscribe tests client subscriptions
func TestClient_SubscribeUnsubscribe(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:            "test-client",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	// Subscribe
	client.Subscribe("traffic")
	if !client.IsSubscribed("traffic") {
		t.Error("Expected client to be subscribed to traffic")
	}

	// Subscribe to another channel
	client.Subscribe("vulns")
	if !client.IsSubscribed("vulns") {
		t.Error("Expected client to be subscribed to vulns")
	}

	// Unsubscribe
	client.Unsubscribe("traffic")
	if client.IsSubscribed("traffic") {
		t.Error("Expected client to NOT be subscribed to traffic")
	}
	if !client.IsSubscribed("vulns") {
		t.Error("Expected client to still be subscribed to vulns")
	}
}

// TestClient_SubscribeMultiple tests subscribing to multiple channels
func TestClient_SubscribeMultiple(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:            "test-client",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	// Subscribe to multiple channels
	channels := []string{"traffic", "vulns", "intercept"}
	for _, ch := range channels {
		client.Subscribe(ch)
	}

	for _, ch := range channels {
		if !client.IsSubscribed(ch) {
			t.Errorf("Expected to be subscribed to %s", ch)
		}
	}
}

// TestMessage tests message structure
func TestMessage(t *testing.T) {
	msg := &Message{
		Type:      "data",
		Action:    "create",
		Channel:   "traffic",
		Data:      map[string]interface{}{"url": "https://example.com", "method": "GET"},
		Timestamp: time.Now(),
		ClientID:  "client-123",
	}

	if msg.Type != "data" {
		t.Errorf("Expected Type 'data', got '%s'", msg.Type)
	}

	if msg.Channel != "traffic" {
		t.Errorf("Expected Channel 'traffic', got '%s'", msg.Channel)
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var unmarshaled Message
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if unmarshaled.Type != msg.Type {
		t.Errorf("Expected Type '%s', got '%s'", msg.Type, unmarshaled.Type)
	}
}

// TestHub_ClientCount tests getting client count
func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", hub.ClientCount())
	}

	// Register clients
	client1 := &Client{
		ID:            "client-1",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}
	client2 := &Client{
		ID:            "client-2",
		hub:           hub,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	hub.Register(client1)
	hub.Register(client2)

	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 2 {
		t.Errorf("Expected 2 clients, got %d", hub.ClientCount())
	}
}

// TestHub_MultipleClients tests handling multiple clients
func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// Create multiple clients with different subscriptions
	clients := make([]*Client, 3)
	for i := 0; i < 3; i++ {
		clients[i] = &Client{
			ID:            string(rune('A' + i)),
			hub:           hub,
			send:          make(chan []byte, 256),
			subscriptions: map[string]bool{"traffic": true},
		}
		hub.Register(clients[i])
	}

	time.Sleep(10 * time.Millisecond)

	// Broadcast to all
	msg := &Message{
		Type:      "data",
		Channel:   "traffic",
		Data:      "test",
		Timestamp: time.Now(),
	}
	hub.Broadcast(msg)

	// All clients should receive
	for i, client := range clients {
		select {
		case <-client.send:
			// Expected
		case <-time.After(50 * time.Millisecond):
			t.Errorf("Client %d did not receive message", i)
		}
	}
}
