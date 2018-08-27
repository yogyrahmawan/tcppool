package main

import (
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp4", "127.0.0.1:6666")
	if err != nil {
		log.Fatalf("Error listening server address %v", err)

	}
	defer l.Close()

	log.Println("Tcp server started")

	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		go handleTCPReq(conn)

	}
}

func handleTCPReq(conn net.Conn) {
	for {

		re := make([]byte, 4)
		n, err := conn.Read(re)
		if err == nil {
			if n == 4 {
				conn.Write([]byte("PONG"))
			} else {
				conn.Write([]byte("CHECK"))
			}
		}

	}
}
