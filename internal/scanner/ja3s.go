package scanner

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"time"

)

func fingerprintJA3S(ctx context.Context, host string, port int) (string, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "", fmt.Errorf("empty host")
	}

	dialer := net.Dialer{Timeout: 3 * time.Second}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return "", fmt.Errorf("ja3s dial: %w", err)
	}
	defer conn.Close()

	payload := buildJA3SClientHello(host)
	_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write(payload); err != nil {
		return "", fmt.Errorf("ja3s write: %w", err)
	}

	buff := make([]byte, 8192)
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := conn.Read(buff)
	if err != nil || n < 9 {
		return "", fmt.Errorf("ja3s read: %w", err)
	}

	hello, err := parseServerHelloBasic(buff[:n])
	if err != nil {
		return "", fmt.Errorf("ja3s parse: %w", err)
	}

	hash := digestJA3S(hello)
	if hash == "" {
		return "", fmt.Errorf("ja3s fingerprint unavailable")
	}
	return hash, nil
}

// buildJA3SClientHello returns a TLS 1.2 ClientHello with SNI for JA3S probing.
func buildJA3SClientHello(serverName string) []byte {
	var random [32]byte
	_, _ = rand.Read(random[:])

	hostname := []byte(serverName)
	sniList := make([]byte, 0, 5+len(hostname))
	sniList = append(sniList, 0x00) // name_type host_name
	sniList = append(sniList, byte(len(hostname)>>8), byte(len(hostname)))
	sniList = append(sniList, hostname...)

	sniExt := make([]byte, 0, 4+len(sniList))
	sniExt = append(sniExt, 0x00, 0x00) // server_name
	sniExt = append(sniExt, byte(len(sniList)>>8), byte(len(sniList)))
	sniExt = append(sniExt, sniList...)

	extensions := append([]byte{
		// supported_groups
		0x00, 0x0a, 0x00, 0x08, 0x00, 0x06, 0x00, 0x17, 0x00, 0x18, 0x00, 0x19,
		// ec_point_formats
		0x00, 0x0b, 0x00, 0x02, 0x01, 0x00,
		// signature_algorithms
		0x00, 0x0d, 0x00, 0x14, 0x00, 0x12, 0x04, 0x03, 0x08, 0x04, 0x04, 0x01, 0x05, 0x03, 0x08, 0x05, 0x05, 0x01, 0x08, 0x06, 0x06, 0x01, 0x02, 0x01,
	}, sniExt...)

	cipherSuites := []byte{
		0x13, 0x01, 0x13, 0x02, 0x13, 0x03, 0xc0, 0x2c, 0xc0, 0x30, 0x00, 0x9f,
		0xcc, 0xa9, 0xcc, 0xa8, 0xcc, 0xaa, 0xc0, 0x2b, 0xc0, 0x2f, 0x00, 0x9e,
		0xc0, 0x24, 0xc0, 0x28, 0x00, 0x6b, 0xc0, 0x23, 0xc0, 0x27, 0x00, 0x67,
	}

	handshake := make([]byte, 0, 256)
	handshake = append(handshake, 0x01) // ClientHello
	handshakeLenPos := len(handshake)
	handshake = append(handshake, 0, 0, 0) // placeholder uint24 length
	handshake = append(handshake, 0x03, 0x03) // TLS 1.2
	handshake = append(handshake, random[:]...)
	handshake = append(handshake, 0x00) // session id length
	handshake = append(handshake, byte(len(cipherSuites)>>8), byte(len(cipherSuites)))
	handshake = append(handshake, cipherSuites...)
	handshake = append(handshake, 0x01, 0x00) // compression: null
	handshake = append(handshake, byte(len(extensions)>>8), byte(len(extensions)))
	handshake = append(handshake, extensions...)

	bodyLen := len(handshake) - 4
	binary.BigEndian.PutUint32(handshake[handshakeLenPos:handshakeLenPos+4], uint32(bodyLen))
	handshake[handshakeLenPos] = byte(bodyLen >> 16)

	record := make([]byte, 0, 5+len(handshake))
	record = append(record, 0x16, 0x03, 0x01)
	record = append(record, byte(len(handshake)>>8), byte(len(handshake)))
	record = append(record, handshake...)
	return record
}