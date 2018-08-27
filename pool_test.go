package tcppool

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PoolTestSuite struct {
	suite.Suite
	ServerAddress string
	Pool          *Pool
	TCPConn       net.Listener
}

func (p *PoolTestSuite) SetupSuite() {
	fmt.Println("Setup Suite")
	p.ServerAddress = "127.0.0.1:6666"

	go p.initTCPServer(p.ServerAddress)

	// give chance tcp server to up
	time.Sleep(1 * time.Second)

	p.createConnection()
}

func (p *PoolTestSuite) TearDownSuite() {
	p.TCPConn.Close()
}

func (p *PoolTestSuite) createConnection() {
	factory := func() (net.Conn, error) { return net.Dial("tcp", p.ServerAddress) }
	heartbeat := func(conn net.Conn) error {
		fmt.Println("Calling heartbeat")
		_, err := conn.Write([]byte("CHECK"))
		return err
	}

	pool, err := NewPool(2, 4, factory, heartbeat)
	p.Nil(err)
	p.NotNil(pool)

	p.Equal(2, pool.Len())
	p.Pool = pool
}

func (p *PoolTestSuite) TestGetThenPut() {
	c, err := p.Pool.Get()
	p.Nil(err)

	// send write
	n, err := c.Write([]byte("PING"))
	p.Nil(err)
	p.Equal(4, n)

	// read
	re := make([]byte, 4)
	n, err = c.Read(re)
	p.Nil(err)
	p.Equal(4, n)
	p.Equal(string(re), "PONG")

	p.Equal(1, p.Pool.Len())
	p.Pool.Put(c)
	p.Equal(2, p.Pool.Len())
}

func (p *PoolTestSuite) TestCloseTCPServerThenUp() {
	c, err := p.Pool.Get()
	p.Nil(err)

	// send write
	n, err := c.Write([]byte("PING"))
	p.Nil(err)
	p.Equal(4, n)

	// read
	re := make([]byte, 4)
	n, err = c.Read(re)
	p.Nil(err)
	p.Equal(4, n)
	p.Equal(string(re), "PONG")

	p.Nil(p.TCPConn.Close())

	// give time to teardown
	time.Sleep(4 * time.Second)

	go p.initTCPServer(p.ServerAddress)
	time.Sleep(1 * time.Second)

	c2, err := p.Pool.Get()
	// send write
	n, err = c2.Write([]byte("PING"))
	p.Nil(err)
	p.Equal(4, n)

	// read
	re = make([]byte, 4)
	n, err = c2.Read(re)
	p.Nil(err)
	p.Equal(4, n)
	p.Equal(string(re), "PONG")

	// run using c
	n, err = c.Write([]byte("PING"))
	p.Nil(err)
	p.Equal(4, n)

	// read
	re = make([]byte, 4)
	n, err = c.Read(re)
	p.Nil(err)
	p.Equal(4, n)
	p.Equal(string(re), "PONG")

	fmt.Println("Putting again")
	p.Pool.Put(c2)
	p.Pool.Put(c)
}

func (p *PoolTestSuite) initTCPServer(serverAddress string) {
	l, err := net.Listen("tcp4", serverAddress)
	if err != nil {
		log.Fatalf("Error listening server address %v", err)

	}
	defer l.Close()
	p.TCPConn = l

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
		if err == nil && n == 4 {
			conn.Write([]byte("PONG"))
		}

	}
}

func TestPoolTestSuite(t *testing.T) {
	suite.Run(t, new(PoolTestSuite))
}
