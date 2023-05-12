package chanhub

import (
	"context"
	"sync"
)

type Hub struct {
	sync.Mutex
	m map[any][]chan any
}

func New() *Hub {
	return &Hub{
		m: make(map[any][]chan any),
	}
}

func (h *Hub) AddAndWait(ctx context.Context, key any) (any, error) {
	h.Lock()
	ch := make(chan any)
	defer close(ch)
	h.m[key] = append(h.m[key], ch)
	id := len(h.m[key]) - 1
	h.Unlock()

	defer h.deleteChan(key, id)

	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (h *Hub) deleteChan(key any, id int) {
	h.Lock()
	defer h.Unlock()

	v, ok := h.m[key]
	if !ok {
		return
	}

	v = append(v[:id], v[id+1:]...)
	if len(v) == 0 {
		delete(h.m, key)
		return
	}

	h.m[key] = v
}

func (h *Hub) Broadcast(key any, result any) {
	h.Lock()
	defer h.Unlock()

	if len(h.m[key]) == 0 {
		return
	}

	for _, ch := range h.m[key] {
		ch <- result
	}
}
