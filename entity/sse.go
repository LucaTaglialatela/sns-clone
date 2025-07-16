package entity

import (
	"fmt"
	"net/http"
	"sync"
)

type SSEEvent struct {
	Name string
	Data string
}

type Broker struct {
	Clients        map[chan SSEEvent]bool
	NewClients     chan chan SSEEvent
	ClosingClients chan chan SSEEvent
	Messages       chan SSEEvent
	Lock           sync.RWMutex
}

func NewBroker() *Broker {
	b := &Broker{
		Clients:        make(map[chan SSEEvent]bool),
		NewClients:     make(chan chan SSEEvent),
		ClosingClients: make(chan chan SSEEvent),
		Messages:       make(chan SSEEvent),
	}
	go b.listen()
	return b
}

func (b *Broker) listen() {
	for {
		select {
		case s := <-b.NewClients:
			b.Lock.Lock()
			// A new client connected
			b.Clients[s] = true
			b.Lock.Unlock()
		case s := <-b.ClosingClients:
			b.Lock.Lock()
			// A client disconnected
			delete(b.Clients, s)
			b.Lock.Unlock()
		case msg := <-b.Messages:
			b.Lock.RLock()
			// When receiving a message, broadcast it to all connected clients
			for clientChan := range b.Clients {
				select {
				case clientChan <- msg:
				default:
				}
			}
			b.Lock.RUnlock()
		}
	}
}

func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Make sure that the writer supports flushing.
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// TODO set this to the actual base url
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for the client
	messageChan := make(chan SSEEvent)
	b.NewClients <- messageChan

	defer func() {
		b.ClosingClients <- messageChan
	}()

	// Block until the client disconnects
	ctx := r.Context()
	for {
		select {
		case event := <-messageChan:
			// Respond with the message in SSE format
			_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Name, event.Data)
			if err != nil {
				fmt.Printf("Error writing to client: %v", err)
				return
			}
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}
