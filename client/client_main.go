package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	SERVER_HOST     = "127.0.0.1:6378"
	CONNECT_TIMEOUT = 3 * time.Second  // timeout for the initial connection
	IO_TIMEOUT      = 10 * time.Second // timeout for each write/read to the server
)

func main() {
	// Connect with timeout.
	dialer := &net.Dialer{Timeout: CONNECT_TIMEOUT}
	conn, err := dialer.Dial("tcp", SERVER_HOST)
	if err != nil {
		log.Fatalf("dial error: %v", err)
	}
	defer conn.Close()
	log.Println("connected to", SERVER_HOST)

	// Buffered reader/writer for a line-based protocol ('\n' terminated).
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	// Scanner on STDIN: reads one line at a time (newline excluded).
	sc := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ") // prompt
		if !sc.Scan() {
			// EOF on stdin or read error → exit.
			if err := sc.Err(); err != nil {
				log.Printf("stdin error: %v", err)
			}
			return
		}
		line := sc.Text() // already without '\n'; ScanLines also removes trailing '\r' if present

		// Ignore empty lines.
		if len(line) == 0 {
			continue
		}

		// WRITE: send the line + '\n' with timeout
		if err := conn.SetWriteDeadline(time.Now().Add(IO_TIMEOUT)); err != nil {
			log.Printf("set write deadline error: %v", err)
			return
		}
		if _, err := w.WriteString(line + "\n"); err != nil {
			log.Printf("write error: %v", err)
			return
		}
		if err := w.Flush(); err != nil {
			log.Printf("flush error: %v", err)
			return
		}

		// READ: read one response line (terminated by '\n') with timeout
		if err := conn.SetReadDeadline(time.Now().Add(IO_TIMEOUT)); err != nil {
			log.Printf("set read deadline error: %v", err)
			return
		}
		resp, err := r.ReadString('\n')
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				log.Println("read timeout (10s) waiting for server response")
				continue
			}
			if err == io.EOF {
				log.Println("server closed the connection")
			} else {
				log.Printf("read error: %v", err)
			}
			return
		}

		// Print the response (without trailing CR/LF).
		respLine := strings.TrimRight(resp, "\r\n")
		fmt.Println(respLine)

		// If the command was ESC and the response is not an error, terminate the client.
		// Treat any line prefixed with "ERR" (case-insensitive) as an error.
		cmdTrim := strings.TrimSpace(line)
		if strings.EqualFold(cmdTrim, "ESC") && !strings.HasPrefix(strings.ToUpper(respLine), "ERR") {
			return
		}
	}
}
