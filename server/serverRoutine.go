package main

import (
	"fmt"
	"net"
)

func handleServerRoutine(conn net.Conn) {

	defer conn.Close() // Close the connection when the handler finishes

	for {
		//Read data from the client
		buffer := make([]byte, COMMAND_MAX_LEN)
		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Redis clone server: error reading from conn channel", err.Error())

			//send back error message
			conn.Write([]byte("Redis clone server: error in command reception"))
			continue
		}

		fmt.Printf("Redis clone server, received from %s: %s\n", conn.RemoteAddr(), string(buffer[:n]))

		//process command
		_, err = conn.Write([]byte("..."))

		if err != nil {
			fmt.Println("Redis clone server, error writing:", err)
			return
		}
	}
}
