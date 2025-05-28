package utils

import "sync"

type PubSubChannel[T any] struct {
	mu     sync.Mutex
	subs   []chan T
	quit   chan struct{}
	closed bool
}

func NewPubSubChannel[T any]() *PubSubChannel[T] {
	return &PubSubChannel[T]{
		subs: make([]chan T, 0, 2),
		quit: make(chan struct{}),
	}
}

func (b *PubSubChannel[T]) Publish(msg T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	for _, ch := range b.subs {
		ch <- msg
	}
}

func (b *PubSubChannel[T]) Subscribe() <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	ch := make(chan T)
	b.subs = append(b.subs, ch)
	return ch
}

func (b *PubSubChannel[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
	close(b.quit)

	for _, ch := range b.subs {
		close(ch)
	}
}
