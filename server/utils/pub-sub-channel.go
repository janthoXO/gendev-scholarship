package utils

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type PubSubChannel[T any] struct {
	mu        sync.Mutex
	subs      []chan T
	closed    bool
	publishWg sync.WaitGroup // Track ongoing publish operations
}

func NewPubSubChannel[T any]() *PubSubChannel[T] {
	return &PubSubChannel[T]{
		subs: make([]chan T, 0, 2),
	}
}

func (b *PubSubChannel[T]) Publish(msg T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		log.Debug("Channel is closed, cannot publish message")
		return
	}

	// Increment wait group to track this publish operation
	b.publishWg.Add(1)

	// use goroutine to make the pub asynchronous 
	go func() {
		// Create a wait group to track all the goroutines
		var wg sync.WaitGroup
		for _, ch := range b.subs {
			wg.Add(1)
			go func(c chan T) {
				defer wg.Done()
				c <- msg
			}(ch)
		}

		// wait for all sends to complete
		wg.Wait()
		b.publishWg.Done()
	}()
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
	// Wait for all ongoing publish operations to complete before closing their channels
	b.publishWg.Wait()

	for _, ch := range b.subs {
		close(ch)
	}
}
