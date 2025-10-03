package balancer

import "time"

type Item[T any] struct {
	item        *T
	maxRequests int
	numRequests int
	timestamp   int64
}

func NewItem[T any](item *T, maxRequests int) *Item[T] {
	return &Item[T]{item: item, maxRequests: maxRequests}
}

func (i *Item[T]) tooManyRequests() bool {
	if time.Now().Unix() == i.timestamp {
		return i.numRequests >= i.maxRequests
	}
	return false
}

func (i *Item[T]) addRequest() int {
	t := time.Now().Unix()
	if t != i.timestamp {
		i.timestamp = t
		i.numRequests = 0
	}
	i.numRequests++
	return i.numRequests
}
