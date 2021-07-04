package pool

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// channelPool implements the Pool interface based on buffered channels
type channelPool struct {
	// storage for our net.Conn connections
	mutex sync.RWMutex
	conns chan net.Conn

	// net.Conn generator
	factory Factory
}

// Factory is a function to create new connections.
type Factory func() (net.Conn, error)

// NewChannelPool returns a new pool based on buffered channels with an initial
// capacity and maximum capacity. Factory is used when initial capacity is
// greater than zero to fill the pool. A zero initialCap doesn't fill the Pool
// until a new Get() is called. During a Get(), If there is no new connection
// available in the pool, a new connection will be created via the Factory()
// method.
func NewChannelPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		conns:   make(chan net.Conn, maxCap),
		factory: factory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.

	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}

	return c, nil
}

func (c *channelPool) getConnsAndFactory() (chan net.Conn, Factory) {
	c.mutex.RLock()
	conns := c.conns
	factory := c.factory
	c.mutex.RUnlock()
	return conns, factory
}

// Get If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *channelPool) Get() (net.Conn, error) {
	conns, factory := c.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}
	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}
		return c.wrapConn(conn), nil
	default:
		fmt.Println("something string happened")
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		return c.wrapConn(conn), nil
	}
}

// PUT puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *channelPool) put(conn net.Conn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.conns == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}

	select {
	case c.conns <- conn:
		return nil
	default:
		// pool is full, close pased connection
		return conn.Close()
	}
}

// Close the TCP conn
func (c *channelPool) Close() {
	c.mutex.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mutex.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		if pc, ok := conn.(*PoolConn); ok {
			pc.MarkUnusable()
			pc.Close()
		}
	}
}

// Len Get the current number of connection
func (c *channelPool) Len() int {
	conns, _ := c.getConnsAndFactory()
	return len(conns)
}
