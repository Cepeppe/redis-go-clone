package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
)

var keyExpirations *KeyExpirationMinHeap

// Here we store db data
var keyDataSpace = map[string]string{}

var dataLock sync.Mutex

////////////

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

// tryLoadRdbFile attempts to load an RDB file located at the given path.
// Initializes keys expirations when exp_ts != NO_EXP_TS
// Behavior:
// - If the file does not exist: return nil (nothing to load).
// - If the path is not a regular file: return an error.
// - If the file is empty: return nil.
// - Otherwise: open the file, run the existing processing, and return any error produced.
func tryLoadRdbFile(path string) error {

	// Verify existence and type.
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// File missing: nothing to load, not an error.
			// Create file and return
			os.Create(path)
			return nil
		}
		// Other stat error: propagate.
		return err
	}
	if !fi.Mode().IsRegular() {
		return fmt.Errorf("path is not a regular file: %s", path)
	}
	if fi.Size() == 0 {
		// Empty file: nothing to load.
		return nil
	}

	// Open the file.
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for {
		key, value, key_exp_ts, err := readRdbEntry(f)

		if err != nil {
			if err == io.EOF {
				//end of file reached
				break
			}

			// unexpected error
			return err
		}

		keyDataSpace[key] = value

		//Aggiungi la chiave alla heap di scadenza solo se ha un timestamp valido
		if key_exp_ts != NO_EXP_TS {
			keyExpirations.PushItem(KeyExpiration{key: key, expire_timestamp: key_exp_ts})
		}
	}

	return nil
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
	if keyDataSpace == nil {
		b.WriteString("  state: nil\n")
	} else {
		b.WriteString(fmt.Sprintf("  size: %d\n", len(keyDataSpace)))
		if len(keyDataSpace) == 0 {
			b.WriteString("  entries: {}\n")
		} else {
			// Collect and sort keys for deterministic output.
			keys := make([]string, 0, len(keyDataSpace))
			for k := range keyDataSpace {
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
				v := keyDataSpace[k]
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
