package tcppool

import (
	"errors"
	//"io"
	"fmt"
	"net"
	"sync"
)

var (
	// ErrPoolClosed represent error when pool is closed
	ErrPoolClosed = errors.New("pool is closed")
	// ErrConnClosed represent error when connection closed
	ErrConnClosed = errors.New("connection already closed")
	// ErrInvalidConfig represnt error when config is not valid
	ErrInvalidConfig = errors.New("invalid config")
)

// Pool represent pool
// conns is bounded semaphore
type Pool struct {
	sync.Mutex
	min         uint
	max         uint
	currentSize uint
	conns       chan net.Conn
	factory     func() (net.Conn, error)
	heartbeat   func(conn net.Conn) error
}

// NewPool create pool .
// Factory is used to create connection
// min connection
// max connection allowed
func NewPool(min, max uint, factory func() (net.Conn, error), heartbeat func(conn net.Conn) error) (*Pool, error) {
	if err := validateConfig(min, max, factory); err != nil {
		return nil, err
	}

	pool := &Pool{
		min:       min,
		max:       max,
		factory:   factory,
		conns:     make(chan net.Conn, max),
		heartbeat: heartbeat,
	}

	if err := pool.initConn(); err != nil {
		return nil, err
	}

	return pool, nil
}

func validateConfig(min, max uint, factory func() (net.Conn, error)) error {
	if factory == nil || min > max {
		return ErrInvalidConfig
	}

	return nil
}

func (p *Pool) initConn() error {
	for i := uint(0); i < p.min; i++ {
		conn, err := p.factory()
		if err != nil {
			return err
		}
		p.conns <- conn
		p.currentSize++
	}

	return nil
}

// Get connection pool from conns
func (p *Pool) Get() (net.Conn, error) {
	p.Lock()
	defer p.Unlock()

	if p.conns == nil {
		return nil, ErrPoolClosed
	}

	var err error
	select {
	case conn := <-p.conns:
		if err = p.heartbeat(conn); err != nil {
			fmt.Println("detected connection closed")
			conn, err = p.factory()
			if err != nil {
				return nil, err
			}
		}

		if conn == nil {
			conn, err = p.factory()
			if err != nil {
				return nil, err
			}

		}

		return conn, nil
	default:
		return p.factory()
	}
}

// Put back conn to pool
func (p *Pool) Put(conn net.Conn) error {
	p.Lock()
	defer p.Unlock()
	if p.conns == nil {
		return ErrPoolClosed
	}

	if conn == nil {
		return ErrConnClosed
	}

	select {
	case p.conns <- conn:
		return nil
	default:
		// pool is full, closing connection
		conn.Close()
	}
	return nil
}

// Close this connection
func (p *Pool) Close(conn net.Conn) error {
	if conn == nil {
		return ErrConnClosed
	}
	return conn.Close()
}

// Destroy pool
// Closing connection
// Set all variable to nil
func (p *Pool) Destroy() {
	p.Lock()
	defer p.Unlock()

	p.factory = nil
	if p.conns == nil {
		return
	}

	for v := range p.conns {
		if v != nil {
			p.Close(v)
		}
	}
	p.conns = nil

}

// Len is returning size of conns created
func (p *Pool) Len() int {
	return len(p.conns)
}
