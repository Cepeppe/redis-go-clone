package main

import (
	"sync"
)

// KeyDataSpace is a thread-safe wrapper around a map[string]string.
// It uses a sync.RWMutex to manage concurrent access.
type KeyDataSpace struct {
	data map[string]string
	mu   sync.RWMutex // Read-Write Mutex to protect the map
}

// Here we store db data
var keyDataSpace *KeyDataSpace

// NewKeyDataSpace creates and returns a pointer to a new KeyDataSpace instance.
func NewKeyDataSpace() *KeyDataSpace {
	return &KeyDataSpace{
		data: make(map[string]string),
	}
}

// Add inserts or updates a key-value pair in a thread-safe manner.
// It requires an exclusive write lock.
func (s *KeyDataSpace) Add(key, value string) {
	// Acquire a write lock to ensure exclusive access for modification.
	s.mu.Lock()
	defer s.mu.Unlock() // Release the lock when the function returns.

	s.data[key] = value
}

// Remove deletes a key from the map in a thread-safe manner.
// It requires an exclusive write lock.
func (s *KeyDataSpace) Remove(key string) {
	// Acquire a write lock to ensure exclusive access for modification.
	s.mu.Lock()
	defer s.mu.Unlock() // Release the lock when the function returns.

	delete(s.data, key)
}

// Exists checks if a key is present in the map in a thread-safe manner.
// It only requires a read lock, allowing multiple concurrent reads.
func (s *KeyDataSpace) Exists(key string) bool {
	// Acquire a read lock. Multiple readers can hold this lock simultaneously.
	s.mu.RLock()
	defer s.mu.RUnlock() // Release the read lock when the function returns.

	_, found := s.data[key]
	return found
}

// --- Optional: A function to retrieve a value safely ---

// Get retrieves the value for a key in a thread-safe manner.
// It returns the value and a boolean indicating if the key was found.
func (s *KeyDataSpace) Get(key string) (string, bool) {
	// Acquire a read lock.
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, found := s.data[key]
	return value, found
}

// IsInitialized checks if the internal map has been initialized (i.e., is not nil).
func (s *KeyDataSpace) IsInitialized() bool {
	// Acquire a read lock.
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.data != nil
}

// Length returns the total number of key-value pairs stored in the map.
func (s *KeyDataSpace) Length() int {
	// Acquire a read lock.
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.data)
}

// Keys returns a slice containing all the keys present in the map.
func (s *KeyDataSpace) Keys() []string {
	// Acquire a read lock.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a slice with capacity equal to the current map size for efficiency.
	keys := make([]string, 0, len(s.data))

	// Iterate over the map and append each key to the slice.
	for key := range s.data {
		keys = append(keys, key)
	}

	return keys
}

// DeepCopy creates a complete, independent clone of the KeyDataSpace.
// It acquires a read lock on the original map to ensure a consistent snapshot.
func (s *KeyDataSpace) DeepCopy() *KeyDataSpace {
    s.mu.RLock() // Acquire read lock on the original map
    defer s.mu.RUnlock()

    // 1. Create a new map with the same capacity
    clonedData := make(map[string]string, len(s.data))

    // 2. Copy every key-value pair from the original map
    for key, value := range s.data {
        clonedData[key] = value
    }

    // 3. Create the new KeyDataSpace instance with its own fresh RWMutex
    return &KeyDataSpace{
        data: clonedData,
        // The mutex is zero-valued (fresh) and ready to use, ensuring
        // the snapshot is completely independent.
    }
}