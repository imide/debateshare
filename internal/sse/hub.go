package sse

import "sync"

type Event struct {
	Name string
	Data string // payload (json)
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[chan Event]struct{}),
	}
}

func (h *Hub) Subscribe(code string) (ch chan Event, unsub func()) {
	ch = make(chan Event, 8)
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[code] == nil {
		h.rooms[code] = make(map[chan Event]struct{})
	}
	h.rooms[code][ch] = struct{}{}

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		subs, ok := h.rooms[code]
		if !ok {
			return // already closed by CloseRoom, no need to do anything
		}
		if _, ok := subs[ch]; !ok {
			return
		}
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(h.rooms, code)
		}
	}
}

func (h *Hub) Broadcast(code string, ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.rooms[code] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// broadcast final event to every sub before tearing that shit down
func (h *Hub) CloseRoom(code string, ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for ch := range h.rooms[code] {
		select {
		case ch <- ev:
		default:
		}
		close(ch)
	}
	delete(h.rooms, code)
}
