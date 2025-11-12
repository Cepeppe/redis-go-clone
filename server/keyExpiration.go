// File: expiration_heap.go
//
// Package note:
//   This file must reside in the same package as KeyExpiration,
//   because its fields (`key`, `expire_timestamp`) are unexported.
//
// Purpose:
//   Min-heap for KeyExpiration, ordered by `expire_timestamp`
//   so that the earliest deadline is always at the top.
//
// Asymptotic costs:
//   - Peek (min):            O(1)
//   - Push (insert):         O(log n)
//   - PopMin (extract min):  O(log n)
//   - Init (heapify slice):  O(n)
//
// Concurrency:
//   Not thread-safe. Guard with a mutex or use a single-owner goroutine
//   in concurrent contexts.
//

package main

import "container/heap"

const NO_EXP_TS int64 = -1

type KeyExpiration struct {
	key string
	expire_timestamp int64
}

// KeyExpirationMinHeap implements a min-heap of KeyExpiration ordered by
// expire_timestamp (earliest timestamp has highest priority).
type KeyExpirationMinHeap []KeyExpiration

// Len returns the number of items.
// Part of heap.Interface.
func (h KeyExpirationMinHeap) Len() int { return len(h) }

// Less reports whether element i should sort before element j.
// For a min-heap, the smallest expire_timestamp must come first.
// Part of heap.Interface.
func (h KeyExpirationMinHeap) Less(i, j int) bool {
	return h[i].expire_timestamp < h[j].expire_timestamp
}

// Swap exchanges elements i and j.
// Part of heap.Interface.
func (h KeyExpirationMinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

// Push appends a new value to the underlying slice.
// Intended to be used by container/heap.
// Part of heap.Interface.
func (h *KeyExpirationMinHeap) Push(x any) {
	*h = append(*h, x.(KeyExpiration))
}

// Pop removes and returns the last element from the underlying slice.
// Intended to be used by container/heap.
// Part of heap.Interface.
func (h *KeyExpirationMinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// Init transforms the underlying slice into a valid heap in O(n).
func (h *KeyExpirationMinHeap) Init() { heap.Init(h) }

// Peek returns the smallest (earliest-expiring) element without removing it.
// Returns (zeroValue, false) if the heap is empty.
func (h KeyExpirationMinHeap) Peek() (KeyExpiration, bool) {
	if len(h) == 0 {
		var zero KeyExpiration
		return zero, false
	}
	return h[0], true
}

// PushItem inserts a KeyExpiration while maintaining the heap property.
func (h *KeyExpirationMinHeap) PushItem(item KeyExpiration) { heap.Push(h, item) }

// PopMin removes and returns the smallest (earliest-expiring) element.
// Returns (zeroValue, false) if the heap is empty.
func (h *KeyExpirationMinHeap) PopMin() (KeyExpiration, bool) {
	if h.Len() == 0 {
		var zero KeyExpiration
		return zero, false
	}
	return heap.Pop(h).(KeyExpiration), true
}
