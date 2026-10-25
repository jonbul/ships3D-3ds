// Package transport opens the game's connections to the server on every platform the game runs
// on: through the console's socket service on the Nintendo 3DS (go-3ds's network), and through
// Go's own net and crypto/tls elsewhere. Either way a connection never makes a frame wait: Read
// returns what has arrived, Write queues, and Flush sends what can go now. That is what the
// websocket and web packages take.
package transport

// Connection is an open connection to the server.
type Connection interface {
	// Read copies into buffer what has arrived and returns at once: 0 bytes and no error when
	// nothing has.
	Read(buffer []byte) (int, error)
	// Write queues bytes and sends what can go now; it never waits.
	Write(bytes []byte) (int, error)
	// Flush sends what can go now of what Write queued; call it once a frame.
	Flush() error
	// Pending is how many written bytes still wait to be sent.
	Pending() int
	Close() error
}
