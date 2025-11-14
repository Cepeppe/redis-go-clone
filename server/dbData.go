package main

import (
	"fmt"
	"sort"
	"strings"
)


// EnsureKeyExpirationMinHeap allocates and initializes the heap if nil.
// If *dst is already non-nil, it leaves it unchanged.
func initKeyExpirationMinHeap(dst **KeyExpirationMinHeap) {
	if dst == nil {
		// Defensive check: a nil destination pointer is a programmer error.
		panic("EnsureKeyExpirationMinHeap: nil destination pointer")
	}
	if *dst == nil {
		// Il costruttore NewKeyExpirationMinHeap() gestisce
		// sia l'allocazione che l'inizializzazione.
		*dst = NewKeyExpirationMinHeap()
	}
}

// initKeyDataSpace safely initializes a *KeyDataSpace pointer if it is currently nil.
//
// The function takes a pointer to a pointer to KeyDataSpace (**KeyDataSpace)
// which allows it to modify the pointer variable in the calling scope.
func initKeyDataSpace(dst **KeyDataSpace) {
	if dst == nil {
		// Defensive check: a nil destination pointer means the caller passed 'nil' 
		// where an address (e.g., &myKeyDataSpace) was expected. This is a programmer error.
		panic("KeyDataSpace: nil destination pointer")
	}
	// Check if the actual *KeyDataSpace pointer (the value pointed to by dst) is nil.
	if *dst == nil {
		// Initialize the pointer in the calling scope with a new instance.
		*dst = NewKeyDataSpace()
	}
}

// printMemoryStatus builds a human-readable snapshot of in-memory structures,
// prints it to stdout, and returns the same string.
// Arguments:
//   - keyExpirations: pointer to min-heap of KeyExpiration (may be nil)
//   - keyDataSpace: key/value space (may be nil)
//
// Display limits:
//   - up to heapShowLimit items from the heap (sorted by expire_timestamp asc)
//   - up to mapShowLimit keys from the map (sorted ascending)
func printMemoryStatus() {
	const (
		heapShowLimit = 16
		mapShowLimit  = 16
	)

	var b strings.Builder
	b.WriteString("=== Memory Status ===\n")

	// --- Heap section ---
	b.WriteString("Heap (KeyExpirationMinHeap):\n")
	if keyExpirations == nil {
		b.WriteString("  state: nil\n")
	} else {
		n := keyExpirations.Len()
		b.WriteString(fmt.Sprintf("  size: %d\n", n))
		if n == 0 {
			b.WriteString("  entries: []\n")
		} else {
			// Copy and sort by expire_timestamp
			copySlice := make([]KeyExpiration, n)
			copy(copySlice, keyExpirations.items)
			sort.Slice(copySlice, func(i, j int) bool {
				return copySlice[i].expire_timestamp < copySlice[j].expire_timestamp
			})

			limit := n
			if limit > heapShowLimit {
				limit = heapShowLimit
			}
			b.WriteString("  entries (earliest first):\n")
			for i := 0; i < limit; i++ {
				ke := copySlice[i]
				b.WriteString(fmt.Sprintf("    - key=%q expire_ms=%d\n", ke.key, ke.expire_timestamp))
			}
			if n > limit {
				b.WriteString(fmt.Sprintf("    ... (%d more)\n", n-limit))
			}
		}
	}

	// --- Map section ---
	b.WriteString("KeyDataSpace (map[string]string):\n")
	if !keyDataSpace.IsInitialized() {
		b.WriteString("  state: nil\n")
	} else {
		b.WriteString(fmt.Sprintf("  size: %d\n", keyDataSpace.Length()))
		if keyDataSpace.Length() == 0 {
			b.WriteString("  entries: {}\n")
		} else {
			// Collect and sort keys for deterministic output.
			keys := make([]string, 0, keyDataSpace.Length())
			for k := range keyDataSpace.data {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			limit := len(keys)
			if limit > mapShowLimit {
				limit = mapShowLimit
			}
			b.WriteString("  entries (sorted by key):\n")
			for i := 0; i < limit; i++ {
				k := keys[i]
				v, _ := keyDataSpace.Get(k)
				b.WriteString(fmt.Sprintf("    - %q: %q\n", k, v))
			}
			if len(keys) > limit {
				b.WriteString(fmt.Sprintf("    ... (%d more)\n", len(keys)-limit))
			}
		}
	}

	fmt.Println()
	out := b.String()
	fmt.Println(out)
}
