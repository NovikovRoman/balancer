package balancer

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

type Balancer[T any] struct {
	sync.RWMutex
	items   []*Item[T]
	shuffle bool
}

func New[T any](items []*Item[T]) *Balancer[T] {
	return &Balancer[T]{
		items: items,
	}
}

func (b *Balancer[T]) SetShuffle(shuffle bool) {
	b.shuffle = shuffle
}

func (b *Balancer[T]) NumItems() int {
	b.RLock()
	defer b.RUnlock()
	return len(b.items)
}

func (b *Balancer[T]) TotalMaxRequests() int {
	b.RLock()
	defer b.RUnlock()
	numReq := 0
	for _, c := range b.items {
		numReq += c.maxRequests
	}
	return numReq
}

func (b *Balancer[T]) TotalFreeRequests() int {
	b.RLock()
	defer b.RUnlock()
	numReq := 0
	for _, c := range b.items {
		numReq += c.maxRequests - c.numRequests
	}
	return numReq
}

func (b *Balancer[T]) Acquire() *T {
	b.Lock()
	defer b.Unlock()

	if b.shuffle {
		rand.Shuffle(len(b.items), func(i, j int) {
			b.items[i], b.items[j] = b.items[j], b.items[i]
		})
	}

	for _, c := range b.items {
		if c.tooManyRequests() {
			continue
		}
		c.addRequest()
		return c.item
	}
	return nil
}

func (b *Balancer[T]) AcquireWait(ctx context.Context, attempts int, pause time.Duration) *T {
	client := b.Acquire()
	if client != nil {
		return client
	}

	if pause <= 0 {
		pause = time.Second
	}
	t := time.NewTicker(pause)
	defer t.Stop()
	for range t.C {
		attempts--
		if attempts == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if client = b.Acquire(); client != nil {
				return client
			}
		}
	}
	return nil
}
