package main

import (
	"log"
	"net"
	"time"
)

const (
	SERVER_HOST     = "127.0.0.1:6378" //port is 6378 because is one less of redis used port (6379)
	COMMAND_MAX_LEN = 2048             //max number of runes for each command (and args)
)

var last_rdb_snapshot_ts int64 // Last RDB snapshot timestamp in millis

func initDataStructures() {
	log.Println("Initializing memorization data structures..")
	initKeyExpirationMinHeap(&keyExpirations)
	log.Println("Initialized key expiration data structure")
	initKeyDataSpace(&keyDataSpace)
	log.Println("Initialized key data space")
	tryLoadRdbFile(RDB_FILE_PATH)
	last_rdb_snapshot_ts = time.Now().UnixMilli()
	log.Println("Loaded key-value data structure and keys expirations data structure")
	log.Println("Completed data structures initializations")
}

func main() {

	log.Println("Redis clone server startup..")

	initDataStructures()
	printMemoryStatus()

	tcp_listener, err := net.Listen("tcp", SERVER_HOST)
	if err != nil {
		log.Println("Socket listening error:", err)
		return
	}

	//before returning close connection
	defer tcp_listener.Close()

	// Run keys expiration process
	go handleKeysExpirationGoRoutine()

	// Run rdb napshot process
	go rdbSnapshotGoRoutine()

	var conn net.Conn
	log.Println("Redis clone server listening on " + SERVER_HOST)
	// Accept incoming connection
	for {
		var err error = nil
		conn, err = tcp_listener.Accept()
		if err != nil {
			log.Println("Redis clone server, error accepting:", err.Error())
			continue
		}

		log.Println("Redis clone server, accepted connection from:", conn.RemoteAddr())

		// Handle the connection in a goroutine
		go handleClientServerRoutine(conn)
	}
}
