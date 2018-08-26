package tcppool

import (
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PoolTestSuite struct {
	suite.Suite
	ServerAddress string
}

func TestPoolTestSuite(t *testing.T) {
	suite.Run(t, new(PoolTestSuite))
}

func (p *PoolTestSuite) SetupTest() {
	p.ServerAddress = "127.0.0.1:6666"

	go initTCPServer(p.ServerAddress)
}

func (p *PoolTestSuite) TestCreateConnection() {
	factory := func() (net.Conn, error) { return net.Dial("tcp", p.ServerAddress) }

	pool, err := NewPool(2, 4, factory)
	p.Nil(err)
	p.NotNil(pool)
}

func (p *PoolTestSuite) TestGetThenPut() {

}

func initTCPServer(serverAddress string) {
	l, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Fatalf("Error listening server address %v", err)
	}
	defer l.Close()

	log.Println("Tcp server started")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting connection, err = %s", err)
			return
		}

		go handleTCPReq(conn)
	}
}

func handleTCPReq(conn net.Conn) {
	for {
		re := make([]byte, 4)
		n, err := conn.Read(re)
		if err == nil && n == 4 {
			conn.Write([]byte("PONG"))
		}
	}
}
