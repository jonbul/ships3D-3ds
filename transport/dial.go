//go:build !nintendo3ds

package transport

import (
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"
)

// How long Dial waits for the server, and how long bytes may wait to be sent before the
// connection counts as stuck: the same as the console's.
const (
	connectTimeout = 5 * time.Second
	stuckAfter     = 5 * time.Second
	readChunkBytes = 32 << 10
)

// The codes of this package's errors, for programs to check with errors.Is.
var (
	ErrClosed = errors.New("transport.closed")
	ErrStuck  = errors.New("transport.stuck")
)

// Dial connects to host on port, over TLS when secure, with Go's net and crypto/tls and the
// system's trusted roots.
func Dial(host string, port int, secure bool) (Connection, error) {
	return DialWith(host, port, secure, nil)
}

// DialWith is Dial with a TLS configuration of one's own (other trusted roots, for a test
// server); nil is the system's.
func DialWith(host string, port int, secure bool, configuration *tls.Config) (Connection, error) {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: connectTimeout}
	var connection net.Conn
	var err error
	if secure {
		if configuration == nil {
			configuration = &tls.Config{}
		}
		configuration = configuration.Clone()
		configuration.ServerName = host
		connection, err = tls.DialWithDialer(dialer, "tcp", address, configuration)
	} else {
		connection, err = dialer.Dial("tcp", address)
	}
	if err != nil {
		return nil, err
	}
	return start(connection), nil
}

// stream is a net.Conn read and written by goroutines of its own, so the game's frame never
// waits on the network: Read takes what the reader has gathered, Write hands bytes to the
// writer.
type stream struct {
	connection net.Conn
	lock       sync.Mutex
	arrived    []byte
	readErr    error
	queued     []byte
	writeErr   error
	waiting    time.Time // since when the oldest queued byte has waited
	wake       chan struct{}
	closed     bool
}

func start(connection net.Conn) *stream {
	opened := &stream{connection: connection, wake: make(chan struct{}, 1)}
	go opened.readLoop()
	go opened.writeLoop()
	return opened
}

func (opened *stream) readLoop() {
	chunk := make([]byte, readChunkBytes)
	for {
		count, err := opened.connection.Read(chunk)
		opened.lock.Lock()
		opened.arrived = append(opened.arrived, chunk[:count]...)
		if err != nil {
			opened.readErr = err
			opened.lock.Unlock()
			return
		}
		opened.lock.Unlock()
	}
}

func (opened *stream) writeLoop() {
	for range opened.wake {
		opened.lock.Lock()
		bytes := opened.queued
		opened.queued = nil
		opened.lock.Unlock()
		if len(bytes) == 0 {
			continue
		}
		if _, err := opened.connection.Write(bytes); err != nil {
			opened.lock.Lock()
			opened.writeErr = err
			opened.lock.Unlock()
			return
		}
	}
}

func (opened *stream) Read(buffer []byte) (int, error) {
	opened.lock.Lock()
	defer opened.lock.Unlock()
	if len(opened.arrived) == 0 && opened.readErr != nil {
		return 0, errors.Join(ErrClosed, opened.readErr)
	}
	count := copy(buffer, opened.arrived)
	opened.arrived = opened.arrived[:copy(opened.arrived, opened.arrived[count:])]
	return count, nil
}

func (opened *stream) Write(bytes []byte) (int, error) {
	opened.lock.Lock()
	if opened.closed || opened.writeErr != nil {
		opened.lock.Unlock()
		return 0, errors.Join(ErrClosed, opened.writeErr)
	}
	if len(opened.queued) == 0 {
		opened.waiting = time.Now()
	}
	opened.queued = append(opened.queued, bytes...)
	opened.lock.Unlock()
	opened.signal()
	return len(bytes), nil
}

func (opened *stream) Flush() error {
	opened.lock.Lock()
	defer opened.lock.Unlock()
	if opened.writeErr != nil {
		return errors.Join(ErrClosed, opened.writeErr)
	}
	if len(opened.queued) > 0 && time.Since(opened.waiting) > stuckAfter {
		return ErrStuck
	}
	return nil
}

func (opened *stream) Pending() int {
	opened.lock.Lock()
	defer opened.lock.Unlock()
	return len(opened.queued)
}

func (opened *stream) Close() error {
	opened.lock.Lock()
	if opened.closed {
		opened.lock.Unlock()
		return nil
	}
	opened.closed = true
	opened.lock.Unlock()
	close(opened.wake)
	return opened.connection.Close()
}

// signal wakes the writer, without waiting when it is already awake.
func (opened *stream) signal() {
	select {
	case opened.wake <- struct{}{}:
	default:
	}
}
