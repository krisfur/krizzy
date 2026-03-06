package services

import (
	"sync"
	"time"
)

type BoardEvent struct {
	Type         string `json:"type"`
	BoardID      int64  `json:"board_id"`
	CardID       int64  `json:"card_id,omitempty"`
	ColumnID     int64  `json:"column_id,omitempty"`
	FromColumnID int64  `json:"from_column_id,omitempty"`
	ToColumnID   int64  `json:"to_column_id,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	OccurredAt   int64  `json:"occurred_at"`
}

type BoardEventHub struct {
	mu          sync.RWMutex
	subscribers map[int64]map[chan BoardEvent]struct{}
}

func NewBoardEventHub() *BoardEventHub {
	return &BoardEventHub{
		subscribers: make(map[int64]map[chan BoardEvent]struct{}),
	}
}

func (h *BoardEventHub) Subscribe(boardID int64) chan BoardEvent {
	ch := make(chan BoardEvent, 16)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscribers[boardID]; !ok {
		h.subscribers[boardID] = make(map[chan BoardEvent]struct{})
	}
	h.subscribers[boardID][ch] = struct{}{}

	return ch
}

func (h *BoardEventHub) Unsubscribe(boardID int64, ch chan BoardEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	boardSubs, ok := h.subscribers[boardID]
	if !ok {
		close(ch)
		return
	}

	if _, ok := boardSubs[ch]; ok {
		delete(boardSubs, ch)
		close(ch)
	}

	if len(boardSubs) == 0 {
		delete(h.subscribers, boardID)
	}
}

func (h *BoardEventHub) Publish(event BoardEvent) {
	event.OccurredAt = time.Now().UnixMilli()

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers[event.BoardID] {
		select {
		case ch <- event:
		default:
		}
	}
}
