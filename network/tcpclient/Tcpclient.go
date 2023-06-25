package tcpclient

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"
)

const limitMessageSize = 16384

type Client interface {
	// GetHostAddr is used to get address of TCP host.
	GetHostAddr() (hostAddr string)
	// Send is used to send and get TCP data via TCP connection.
	Send(input []byte) (output []byte, err error)
	// Close is used to close all connections in connection pool.
	Close() (err error)
}

type defaultClient struct {
	status          bool
	hostAddr        string
	minConns        int
	maxConns        int
	idleConnTimeout time.Duration
	waitConnTimeout time.Duration
	clearPeriod     time.Duration
	poolSize        int
	poolCounter     uint64
	poolLock        sync.Mutex
	connPool        chan *connection
	initDelay       time.Duration
	maxRetry        int
}

type connection struct {
	id         uint64
	tcpConn    net.Conn
	lastActive time.Time
}

func NewClient(hostAddr string, minConns, maxConns int, idleConnTimeout, waitConnTimeout, clearPeriod time.Duration) (client Client, err error) {
	c := &defaultClient{
		status:          true,
		hostAddr:        hostAddr,
		minConns:        minConns,
		maxConns:        maxConns,
		idleConnTimeout: idleConnTimeout,
		waitConnTimeout: waitConnTimeout,
		clearPeriod:     clearPeriod,
		poolSize:        0,
		poolLock:        sync.Mutex{},
		connPool:        make(chan *connection, maxConns),
		initDelay:       time.Millisecond * 10,
		maxRetry:        4,
	}
	for i := 0; i < c.minConns; i++ {
		if _, err = c.fillConnPool(false); err != nil {
			c.Close()
			return client, err
		}
	}
	go c.startPoolManager()
	return c, nil
}

// GetHostAddr is used to get address of TCP host.
func (c *defaultClient) GetHostAddr() (hostAddr string) {
	return c.hostAddr
}

// Send is used to send and get TCP data via TCP connection.
func (c *defaultClient) Send(input []byte) (output []byte, err error) {
	if !c.status {
		return nil, errors.New("all connections in connection pool are already closed")
	}
	conn := &connection{}
	select {
	case conn = <-c.connPool:
	case <-time.After(c.waitConnTimeout):
		conn, err = c.fillConnPool(true)
		if err != nil {
			return nil, err
		}
	}
	originConn := conn
	conn.lastActive = time.Now()
	defer func() {
		if conn == nil {
			c.drainConnPool(originConn, true)
		} else {
			c.connPool <- conn
		}
	}()
	retry(
		c.initDelay,
		c.maxRetry,
		func() {
			output, err = c.sendAndReceive(conn, input)
		},
		func() bool {
			if err != nil {
				_, err = c.drainConnPool(conn, true)
				if err != nil {
					return true
				}
				originConn = nil
				conn, err = c.fillConnPool(true)
				return true
			}
			return false
		},
	)
	return output, err
}

// Close is used to close all connections in connection pool.
func (c *defaultClient) Close() (err error) {
	for empty := false; !empty && c.poolSize > 0; {
		conn := <-c.connPool
		empty, err = c.drainConnPool(conn, true)
		if err != nil {
			return err
		}
	}
	c.status = false
	return nil
}

func (c *defaultClient) sendAndReceive(conn *connection, input []byte) (output []byte, err error) {
	if conn == nil {
		return nil, errors.New("connection is empty")
	}
	// send data length
	dataSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataSize, uint32(len(input)))
	_, err = conn.tcpConn.Write(dataSize)
	if err != nil {
		return nil, err
	}
	// send data
	for i := 0; i <= len(input)/limitMessageSize; i++ {
		start := i * limitMessageSize
		end := (i + 1) * limitMessageSize
		if end > len(input) {
			end = len(input)
		}
		_, err = conn.tcpConn.Write(input[start:end])
		if err != nil {
			return nil, err
		}
	}
	// set read timeout
	if err != nil {
		return nil, err
	}

	// receive data length
	_, err = conn.tcpConn.Read(dataSize)
	if err != nil {
		return nil, err
	}
	// receive data
	output = make([]byte, binary.LittleEndian.Uint32(dataSize))
	for i := 0; i <= len(output)/limitMessageSize; i++ {
		start := i * limitMessageSize
		end := (i + 1) * limitMessageSize
		if end > len(output) {
			end = len(output)
		}
		_, err = conn.tcpConn.Read(output[start:end])
		if err != nil {
			return nil, err
		}
	}
	return output, err
}

func (c *defaultClient) fillConnPool(getConn bool) (conn *connection, err error) {
	c.poolLock.Lock()
	defer c.poolLock.Unlock()
	if c.poolSize == c.maxConns {
		return nil, errors.New("connection pool is full")
	}
	tcpConn, err := net.Dial("tcp", c.hostAddr)
	if err != nil {
		return nil, err
	}
	c.poolCounter++
	conn = &connection{
		id:         c.poolCounter,
		tcpConn:    tcpConn,
		lastActive: time.Now(),
	}
	c.poolSize++
	if getConn {
		return conn, nil
	}
	c.connPool <- conn
	return nil, nil
}

func (c *defaultClient) startPoolManager() {
	for {
		if !c.status {
			return
		}
		poolLength := len(c.connPool)
		for i := 0; i < poolLength; i++ {
			conn := <-c.connPool
			if time.Since(conn.lastActive) > c.idleConnTimeout {
				_, _ = c.drainConnPool(conn, false)
			} else {
				c.connPool <- conn
			}
		}
		time.Sleep(c.clearPeriod)
	}
}

func (c *defaultClient) drainConnPool(conn *connection, forceMode bool) (empty bool, err error) {
	c.poolLock.Lock()
	defer c.poolLock.Unlock()
	if c.poolSize == 0 {
		return true, errors.New("connection pool is empty")
	}
	defer func() {
		if err != nil {
			c.connPool <- conn
		}
	}()
	if c.poolSize == c.minConns && !forceMode {
		err = errors.New("pool size cannot be lower than minimum number of connections")
		return false, err
	}
	c.poolSize--
	if conn != nil {
		err = conn.tcpConn.Close()
	}
	if err != nil {
		return false, err
	}
	empty = c.poolSize == 0
	return empty, nil
}

func retry(initDelay time.Duration, maxRetry int, processFn func(), retryCondFn func() bool) {
	processFn()
	delay := initDelay
	for i := 0; i < maxRetry && retryCondFn(); i++ {
		time.Sleep(delay)
		delay *= 2
		processFn()
	}
}
