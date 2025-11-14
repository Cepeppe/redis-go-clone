package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const RDB_FILE_PATH = "rdb.bin"
const RDB_SNAPSHOT_INTERVAL = 3 * time.Second

var NATIVE_ENDIAN = binary.NativeEndian
var rdbFileMutex sync.RWMutex

// An entry is: key_len(uint_32) key(string) data_len(uint_32) data (string) expiration_timestamp_ms(int64)
// returns key and value and expiration_timestamp_ms error
func readRdbEntry(f *os.File) (string, string, int64, error) {

	// READ KEY LEN
	var keyLen uint32
	var err error

	// err is io.ErrUnexpectedEOF if file ended before N bytes,
	// or io.EOF if the file had 0 bytes left.
	if err = binary.Read(f, NATIVE_ENDIAN, &keyLen); err != nil {
		return "", "", -1, err
	}

	// READ KEY
	key_buf := make([]byte, keyLen)
	if err = binary.Read(f, NATIVE_ENDIAN, &key_buf); err != nil {
		return "", "", -1, err
	}

	// READ DATA LEN
	var data_len uint32
	if err = binary.Read(f, NATIVE_ENDIAN, &data_len); err != nil {
		return "", "", -1, err
	}

	// READ DATA
	data_buf := make([]byte, data_len)
	if err = binary.Read(f, NATIVE_ENDIAN, data_buf); err != nil {
		return "", "", -1, err
	}

	// READ EXPIRATION
	var expiration_timestamp_ms int64
	if err = binary.Read(f, NATIVE_ENDIAN, &expiration_timestamp_ms); err != nil {
		return "", "", -1, err
	}

	//We always returns err, cause if it equals io.EOF file is finished
	return string(key_buf), string(data_buf), expiration_timestamp_ms, err
}

// An entry is: key_len(uint_32) key(string) data_len(uint_32) data(string) expiration_timestamp_ms(int64)
// writes a single entry to the given writer.
// accepst an io.Writer (like *bufio.Writer) for performance.
func writeRdbEntry(w io.Writer, key string, data string, exp_ts_ms int64) error {
	// key_len(uint_32)
	var err = binary.Write(w, NATIVE_ENDIAN, uint32(len(key)))
	if err != nil {
		return err
	}

	// key(string)
	_, err = w.Write([]byte(key))
	if err != nil {
		return err
	}

	// data_len(uint_32)
	err = binary.Write(w, NATIVE_ENDIAN, uint32(len(data)))
	if err != nil {
		return err
	}

	// data(string)
	_, err = w.Write([]byte(data))
	if err != nil {
		return err
	}

	// expiration_timestamp_ms(int64)
	err = binary.Write(w, NATIVE_ENDIAN, exp_ts_ms)
	if err != nil {
		return err
	}

	return nil
}

// saveRDBFile performs the complete, atomic, and safe persistence routine.
// It opens the file, clears existing content, writes the snapshot, flushes, and syncs.
func saveRDBFile(rdbFileName string, dataSnapshot *KeyDataSpace, expSnapshot *KeyExpirationMinHeap) error {
	rdbFileMutex.RLock()
	defer rdbFileMutex.RUnlock()

	// Open/Create/Truncate File
	// O_WRONLY: Write-only access
	// O_CREATE: Create the file if it doesn't exist
	// O_TRUNC: TRUNCATE (WIPE) existing content to ensure a clean start
	file, err := os.OpenFile(rdbFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("RDB Snapshot: critical error opening/creating file: %w", err)
	}
	defer file.Close() // Ensure the file is closed

	// Initialize the Buffered Writer
	// This makes writes highly efficient by minimizing system calls.
	writer := bufio.NewWriter(file)

	// Write Data Snapshot
	log.Printf("RDB Snapshot: Writing %d entries\n", len(dataSnapshot.data))
	for key, value := range dataSnapshot.data {
		var exp_ts int64 = NO_EXP_TS
		if ts, exists := keyExpirations.FindExpiration(key); exists {
			exp_ts = ts
		}

		if err := writeRdbEntry(writer, key, value, exp_ts); err != nil {
			return fmt.Errorf("RDB Snapshot: error writing entry for key %s: %w", key, err)
		}
	}

	// Flush the Go buffer (moves data from writer buffer to OS kernel buffer)
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("RDB Snapshot: error during buffer flush: %w", err)
	}

	// Sync file descriptor to disk (forces OS to commit data to physical storage)
	// critical step for persistence guarantee.
	if err := file.Sync(); err != nil {
		return fmt.Errorf("RDB Snapshot: error during disk synchronization: %w", err)
	}

	log.Println("RDB Snapshot: completed successfully")

	return nil
}

// tryLoadRdbFile attempts to load an RDB file located at the given path.
// Initializes keys expirations when exp_ts != NO_EXP_TS
// Behavior:
// - If the file does not exist: return nil (nothing to load).
// - If the path is not a regular file: return an error.
// - If the file is empty: return nil.
// - Otherwise: open the file, run the existing processing, and return any error produced.
func tryLoadRdbFile(path string) error {
	rdbFileMutex.RLock()
	defer rdbFileMutex.RUnlock()
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

		keyDataSpace.Add(key, value)

		//Aggiungi la chiave alla heap di scadenza solo se ha un timestamp valido
		if key_exp_ts != NO_EXP_TS {
			keyExpirations.PushItem(KeyExpiration{key: key, expire_timestamp: key_exp_ts})
		}
	}

	return nil
}
