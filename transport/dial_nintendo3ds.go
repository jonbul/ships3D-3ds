//go:build nintendo3ds

package transport

import "github.com/akrck02/go-3ds/network"

// Dial connects to host on port, over TLS when secure, through the console's socket service.
func Dial(host string, port int, secure bool) (Connection, error) {
	if secure {
		return network.DialSecure(host, port)
	}
	return network.Dial(host, port)
}
