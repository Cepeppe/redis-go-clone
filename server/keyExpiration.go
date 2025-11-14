// File: expiration_heap.go
//
// Package note:
//   This file must reside in the same package as KeyExpiration,
//   because its fields (`key`, `expire_timestamp`) are unexported.
//
// Purpose:
//   Min-heap for KeyExpiration, ordered by `expire_timestamp`
//   so that the earliest deadline is always at the top.
//
//   **This implementation enforces key uniqueness.**
//   Pushing an item with an existing key will update its timestamp
//   and re-order the heap.
//
// Asymptotic costs:
//   - Peek (min):            O(1)
//   - Push (insert/update):  O(log n)
//   - PopMin (extract min):  O(log n)
//   - Remove (by key):       O(log n)
//   - Find (by key):         O(1) (via index)
//
// Concurrency:
//   Not thread-safe. Guard with a mutex or use a single-owner goroutine
//   in concurrent contexts.
//

package main

import "container/heap"

const NO_EXP_TS int64 = -1

type KeyExpiration struct {
	key              string
	expire_timestamp int64
}

// KeyExpirationMinHeap implements a min-heap of KeyExpiration ordered by
// expire_timestamp (earliest timestamp has highest priority).
// It ensures key uniqueness via an internal map.
type KeyExpirationMinHeap struct {
	items []KeyExpiration // The heap's underlying slice
	index map[string]int  // Map: key -> index in items
}

// NewKeyExpirationMinHeap creates a new, empty heap ready for use.
func NewKeyExpirationMinHeap() *KeyExpirationMinHeap {
	return &KeyExpirationMinHeap{
		items: make([]KeyExpiration, 0),
		index: make(map[string]int),
	}
}

// --- Inizio implementazione heap.Interface ---

// Len returns the number of items.
// Part of heap.Interface.
func (h *KeyExpirationMinHeap) Len() int { return len(h.items) }

// Less reports whether element i should sort before element j.
// For a min-heap, the smallest expire_timestamp must come first.
// Part of heap.Interface.
func (h *KeyExpirationMinHeap) Less(i, j int) bool {
	return h.items[i].expire_timestamp < h.items[j].expire_timestamp
}

// Swap exchanges elements i and j, AND updates the index map.
// Part of heap.Interface.
func (h *KeyExpirationMinHeap) Swap(i, j int) {
	// 1. Swap items in the slice
	h.items[i], h.items[j] = h.items[j], h.items[i]

	// 2. Update the index map to reflect the swap
	h.index[h.items[i].key] = i
	h.index[h.items[j].key] = j
}

// Push appends a new value to the underlying slice.
// Intended to be used by container/heap.
// Part of heap.Interface.
func (h *KeyExpirationMinHeap) Push(x any) {
	item := x.(KeyExpiration)
	n := len(h.items)

	// 3. Add the key to the index map, pointing to the *end* of the slice
	//    (The heap logic will call Swap to bubble it up, fixing the index)
	h.index[item.key] = n

	// 4. Append the item to the slice
	h.items = append(h.items, item)
}

// Pop removes and returns the last element from the underlying slice.
// Intended to be used by container/heap.
// Part of heap.Interface.
func (h *KeyExpirationMinHeap) Pop() any {
	old := h.items
	n := len(old)
	item := old[n-1]           // Save the item to return
	old[n-1] = KeyExpiration{} // Zero out the cell in the slice (crucial for GC)
	h.items = old[:n-1]        // Shorten the slice

	// 5. Remove the item from the index map
	delete(h.index, item.key)

	return item // Return the item
}

// --- Fine implementazione heap.Interface ---

// --- Inizio API Pubblica ---

// Peek returns the smallest (earliest-expiring) element without removing it.
// Returns (zeroValue, false) if the heap is empty.
func (h *KeyExpirationMinHeap) Peek() (KeyExpiration, bool) {
	if len(h.items) == 0 {
		var zero KeyExpiration
		return zero, false
	}
	return h.items[0], true
}

// PushItem adds an item or updates the timestamp of an existing item.
// If the key already exists, its timestamp is updated, and the heap
// is adjusted to maintain the heap property.
func (h *KeyExpirationMinHeap) PushItem(item KeyExpiration) {
	if existingIdx, ok := h.index[item.key]; ok {
		// Key exists. Update timestamp and fix heap.
		h.items[existingIdx].expire_timestamp = item.expire_timestamp
		heap.Fix(h, existingIdx) // Re-establish heap invariant
	} else {
		// Key doesn't exist. Push new item onto the heap.
		heap.Push(h, item)
	}
}

// PopMin removes and returns the smallest (earliest-expiring) element.
// Returns (zeroValue, false) if the heap is empty.
func (h *KeyExpirationMinHeap) PopMin() (KeyExpiration, bool) {
	if h.Len() == 0 {
		var zero KeyExpiration
		return zero, false
	}
	// heap.Pop calls our internal Pop() method, which correctly
	// removes from both the slice and the map.
	return heap.Pop(h).(KeyExpiration), true
}

// Remove removes the item associated with the given key, regardless of its
// position in the heap.
// Returns the removed item and true if found, or (zero, false) otherwise.
// This is an O(log n) operation.
func (h *KeyExpirationMinHeap) Remove(key string) (KeyExpiration, bool) {
	if idx, ok := h.index[key]; ok {
		// heap.Remove swaps the element with the last, then Pops.
		// Our Swap and Pop methods will handle updating the index map.
		item := heap.Remove(h, idx)
		return item.(KeyExpiration), true
	}
	var zero KeyExpiration
	return zero, false
}

// --- Fine API Pubblica ---
