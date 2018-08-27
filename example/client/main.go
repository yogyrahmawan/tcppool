package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yogyrahmawan/tcppool"
)

func secs(d time.Duration) int {
	d += (time.Second - time.Nanosecond)
	return int(d.Seconds())
}

func main() {
	factory := func() (net.Conn, error) {
		conn, err := net.Dial("tcp", "127.0.0.1:6666")
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	ping := func(conn net.Conn) error {
		fmt.Println("Calling heartbeat")
		_, err := conn.Write([]byte("CHECK"))
		return err
	}

	fmt.Println("Initiating pool")
	pool, err := tcppool.NewPool(2, 4, factory, ping)
	if err != nil {
		log.Fatalf("error init pool, err=%s", err)
	}

	for i := 0; i < 20; i++ {
		time.Sleep(2 * time.Second)
		c, err := pool.Get()
		if err != nil {
			log.Printf("error getting connection, err = %s", err)
			continue

		}

		fmt.Println("Sending Ping")
		_, err = c.Write([]byte("PING"))
		if err != nil {
			log.Printf("error writing ping, err = %s", err)
		}

		// read
		fmt.Println("Reading pong")
		re := make([]byte, 4)
		_, err = c.Read(re)
		if err != nil {
			log.Printf("error reading pong, err = %s", err)
		}
		log.Println(string(re))

		pool.Put(c)

	}

}
