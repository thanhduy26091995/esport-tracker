package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHub_RegisterAndBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	recv := make(chan []byte, 1)
	c := &client{send: make(chan []byte, 8)}
	hub.register <- c

	// Give Run() goroutine time to process registration
	time.Sleep(10 * time.Millisecond)

	event := ActivityEvent{
		Type:     "bet_placed",
		UserName: "Ric Phan",
		Stake:    5,
	}
	hub.Broadcast(event)

	// Drain send channel
	select {
	case msg := <-c.send:
		recv <- msg
	case <-time.After(200 * time.Millisecond):
		t.Fatal("broadcast not received within timeout")
	}

	var got ActivityEvent
	assert.NoError(t, json.Unmarshal(<-recv, &got))
	assert.Equal(t, "bet_placed", got.Type)
	assert.Equal(t, "Ric Phan", got.UserName)
	assert.Equal(t, 5, got.Stake)
}

func TestHub_Unregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	c := &client{send: make(chan []byte, 8)}
	hub.register <- c
	time.Sleep(10 * time.Millisecond)

	hub.unregister <- c
	time.Sleep(10 * time.Millisecond)

	// After unregister the channel should be closed
	_, open := <-c.send
	assert.False(t, open, "send channel should be closed after unregister")
}

func TestHub_BroadcastNoClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Should not panic with zero clients
	assert.NotPanics(t, func() {
		hub.Broadcast(ActivityEvent{Type: "bet_placed"})
		time.Sleep(10 * time.Millisecond)
	})
}
