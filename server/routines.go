package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// handleClientServerRoutine processes one client connection using a simple line-based protocol.
// Each client message is one line terminated by '\n'; the server replies with exactly one line.
func handleClientServerRoutine(conn net.Conn) {
	defer conn.Close()

	r := bufio.NewReader(conn) // line reader for the socket
	w := bufio.NewWriter(conn) // buffered writer for replies

	for {
		// Read exactly one line (blocks until '\n' or error).
		line, err := r.ReadString('\n')
		if err != nil {
			// Remote closed or transport error; terminate the handler.
			if err == io.EOF {
				log.Println("Redis clone server: connection interrupted from", conn.RemoteAddr())
			} else {
				log.Println("Redis clone server: read error from", conn.RemoteAddr(), ":", err)
			}
			return
		}

		// Strip only the trailing line terminators; preserve internal whitespace.
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			// Ignore empty lines and continue.
			continue
		}

		log.Printf("Redis clone server, received from %s: %s", conn.RemoteAddr(), line)

		// Extract command token and arguments (separators: space or tab).
		cmdTok, _, ok := cutFirstTokenSpaceTab(line)
		if ok != nil {
			// Malformed input; return a single-line error and continue.
			_, _ = w.WriteString("ERR: empty command\n")
			_ = w.Flush()
			continue
		}

		// Handle explicit connection close request (case-insensitive).
		if strings.EqualFold(cmdTok, "ESC") {
			_, _ = w.WriteString("closing connection.\n")
			_ = w.Flush()
			return
		}

		// Canonicalize command to upper-case for map lookup; arguments are kept as-is.

		// Execute handler; always reply with exactly one line.
		res, execErr := tryParseExecuteCommand(line)
		printMemoryStatus()

		if execErr != nil {
			_, _ = w.WriteString("ERR: " + execErr.Error() + "\n")
		} else {
			if res == "" {
				// Provide a minimal positive acknowledgment when handler returns empty output.
				res = "OK"
			}
			_, _ = w.WriteString(res + "\n")
		}
		// Flush the buffered writer to ensure the line is sent immediately.
		if err := w.Flush(); err != nil {
			log.Println("Redis clone server: write/flush error to", conn.RemoteAddr(), ":", err)
			return
		}
	}
}

func handleKeysExpirationGoRoutine() {
	for {
		keyExp, has_elements := keyExpirations.Peek()

		if !has_elements {
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("keysExpirationGoRoutine running...")

		if time.Now().UnixMilli() >= keyExp.expire_timestamp {
			log.Println("Removing expired key:", keyExp.key)
			keyDataSpace.Remove(keyExp.key)
			keyExpirations.PopMin()
		} else {
			time.Sleep(3 * time.Second)
			continue
		}
	}
}

// rdbSnapshotGoRoutine periodically takes a consistent snapshot of the in-memory
// key/data and key/expiration spaces and persists them to disk.
func rdbSnapshotGoRoutine() {
	// Create a Ticker that sends a signal to its channel every RDB_SNAPSHOT_INTERVAL.
	// (idiomatic Go way to handle fixed-interval periodic tasks, ensuring efficient, non-blocking waits)
	ticker := time.NewTicker(RDB_SNAPSHOT_INTERVAL)

	// Ensure the ticker is stopped when the function exits (e.g., if the application shuts down).
	defer ticker.Stop()

	// The loop blocks on 'ticker.C', waiting for the next interval to elapse.
	for range ticker.C {
		log.Println("Starting RDB snaphsot execution..")

		// Create consistent deep copies of the in-memory data structures.
		// DeepCopy methods should acquire read locks (RLock) on the original structures
		// to ensure a consistent, crash-safe snapshot of the data at that moment.
		copyOfKeyDataSpace := keyDataSpace.DeepCopy()
		copyOfKeyExpirationSpace := keyExpirations.DeepCopy()

		// Save the copied data to the RDB file (this is the I/O operation).
		// This function should handle serialization and file writing.
		saveRDBFile(RDB_FILE_PATH, copyOfKeyDataSpace, copyOfKeyExpirationSpace)

		// Record the time when the snapshot finished.
		// This timestamp reflects the moment the persistent state was established.
		last_rdb_snapshot_ts = time.Now().UnixMilli()
	}
}
