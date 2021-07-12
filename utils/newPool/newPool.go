package newPool

import (
	"errors"
	"io"
	"net"
	"oliver/entry/utils/logger"
	"sync"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid pool config")
	ErrPoolClosed    = errors.New("pool closed")
)

type Poolable interface {
	io.Closer
	GetActiveTime() time.Time
}

type factory func() (net.Conn, error)

type Pool interface {
	Acquire() (net.Conn, error) // 获取资源
	Release(net.Conn) error     // 释放资源
	Close(net.Conn) error       // 关闭资源
	Shutdown() error            // 关闭池
}

type GenericPool struct {
	sync.Mutex
	pool        chan net.Conn
	maxOpen     int  // 池中最大资源数
	numOpen     int  // 当前池中资源数
	minOpen     int  // 池中最少资源数
	closed      bool // 池是否已关闭
	maxLifetime time.Duration
	factory     factory // 创建连接的方法
}

func NewGenericPool(minOpen, maxOpen int, maxLifetime time.Duration, factory factory) (*GenericPool, error) {
	if maxOpen <= 0 || minOpen > maxOpen {
		return nil, ErrInvalidConfig
	}
	p := &GenericPool{
		maxOpen:     maxOpen,
		minOpen:     minOpen,
		maxLifetime: maxLifetime,
		factory:     factory,
		pool:        make(chan net.Conn, maxOpen),
	}

	for i := 0; i < minOpen; i++ {
		Conn, err := factory()
		if err != nil {
			continue
		}
		p.numOpen++
		p.pool <- Conn
	}
	return p, nil
}

func (p *GenericPool) Acquire() (net.Conn, error) {
	if p.closed {
		logger.Error.Println("这个连接已经关闭了")
		return nil, ErrPoolClosed
	}
	for {
		closer, err := p.getOrCreate()
		if err != nil {
			return nil, err
		}
		// 如果设置了超时且当前连接的活跃时间+超时时间早于现在，则当前连接已过期
		//if p.maxLifetime > 0 && closer.GetActiveTime().Add(p.maxLifetime).Before(time.Now()) {
		//	p.Close(closer)
		//	continue
		//}
		return closer, nil
	}
}

func (p *GenericPool) getOrCreate() (net.Conn, error) {
	select {
	case Conn := <-p.pool:
		return Conn, nil
	default:
	}
	p.Lock()
	if p.numOpen >= p.maxOpen {
		closer := <-p.pool
		p.Unlock()
		return closer, nil
	}
	// 新建连接
	closer, err := p.factory()
	if err != nil {
		p.Unlock()
		return nil, err
	}
	p.numOpen++
	p.Unlock()
	return closer, nil
}

// Release 释放单个资源到连接池
func (p *GenericPool) Release(Conn net.Conn) error {
	if p.closed {
		return ErrPoolClosed
	}
	p.pool <- Conn
	return nil
}

// Close 关闭单个资源
func (p *GenericPool) Close(Conn net.Conn) error {
	p.Lock()
	Conn.Close()
	p.numOpen--
	p.Unlock()
	return nil
}

// Shutdown 关闭连接池，释放所有资源
func (p *GenericPool) Shutdown() error {
	if p.closed {
		return ErrPoolClosed
	}
	p.Lock()
	close(p.pool)
	for closer := range p.pool {
		closer.Close()
		p.numOpen--
	}
	p.closed = true
	p.Unlock()
	return nil
}
