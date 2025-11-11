package main

import (
	"log"
	"net"
)

const SERVER_HOST = "127.0.0.1:6378" //port is 6378 because is one less of redis used port (6379)
const COMMAND_MAX_LEN = 2048 //max number of runes for each command (and args)

/*
	A SINGLE CLIENT INSTANCE FOR EACH SERVER INSTANCE
*/

func main() {

	log.Println("Redis clone server startup..")

	tcp_listener, err := net.Listen("tcp", SERVER_HOST)
	if err != nil {
		log.Println("Socket listening error:", err)
		return
	}

	//before returning close connection
	defer tcp_listener.Close()

	var conn net.Conn = nil

	log.Println("Redis clone server listening on " + SERVER_HOST)
	// Accept incoming connection
	for {
		var err error = nil
		conn, err = tcp_listener.Accept()
		if err != nil {
			log.Println("Redis clone server, error accepting:", err.Error())
			continue
		}
		break
	}

	log.Println("Redis clone server, accepted connection from:", conn.RemoteAddr())

	// Handle the connection in a goroutine
	go handleServerRoutine(conn)

	// fmt.Println("Redis clone server stopping...")
}
