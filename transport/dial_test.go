//go:build !nintendo3ds

package transport

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net"
	"testing"
	"time"
)

// echoServer answers every byte it gets with the same byte, until the test ends.
func echoServer(test *testing.T, listener net.Listener) (host string, port int) {
	test.Cleanup(func() { listener.Close() }) //nolint:errcheck // closing a test's listener
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go io.Copy(connection, connection) //nolint:errcheck // an echo ends when the client goes
		}
	}()
	address := listener.Addr().(*net.TCPAddr)
	return "127.0.0.1", address.Port
}

// echoes writes a message and reads it back, a frame at a time, as the game does.
func echoes(test *testing.T, connection Connection) {
	test.Helper()
	message := []byte("hello, arena")
	if _, err := connection.Write(message); err != nil {
		test.Fatal(err)
	}
	got := []byte{}
	buffer := make([]byte, 64)
	for deadline := time.Now().Add(3 * time.Second); len(got) < len(message) && time.Now().Before(deadline); {
		if err := connection.Flush(); err != nil {
			test.Fatal(err)
		}
		count, err := connection.Read(buffer)
		if err != nil {
			test.Fatal(err)
		}
		got = append(got, buffer[:count]...)
		time.Sleep(5 * time.Millisecond) // a frame
	}
	if string(got) != string(message) {
		test.Fatalf("got %q back, want %q", got, message)
	}
}

func TestAMessageGoesToTheServerAndBack(test *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		test.Fatal(err)
	}
	host, port := echoServer(test, listener)
	connection, err := Dial(host, port, false)
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()
	echoes(test, connection)
}

func TestAMessageGoesToTheServerAndBackOverTLS(test *testing.T) {
	certificate, roots := selfSigned(test)
	listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{certificate}})
	if err != nil {
		test.Fatal(err)
	}
	host, port := echoServer(test, listener)
	connection, err := DialWith(host, port, true, &tls.Config{RootCAs: roots})
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()
	echoes(test, connection)
}

func TestReadingNothingReturnsAtOnce(test *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		test.Fatal(err)
	}
	host, port := echoServer(test, listener)
	connection, err := Dial(host, port, false)
	if err != nil {
		test.Fatal(err)
	}
	defer connection.Close()
	started := time.Now()
	count, err := connection.Read(make([]byte, 16))
	if count != 0 || err != nil || time.Since(started) > 50*time.Millisecond {
		test.Fatalf("read %d bytes, %v, after %v: want nothing, at once", count, err, time.Since(started))
	}
}

// selfSigned is a certificate for 127.0.0.1, and the roots that trust it.
func selfSigned(test *testing.T) (tls.Certificate, *x509.CertPool) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		test.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "127.0.0.1"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, IsCA: true, BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		test.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		test.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(parsed)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, roots
}
