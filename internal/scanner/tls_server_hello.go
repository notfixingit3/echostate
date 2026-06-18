package scanner

import (
	"errors"

	"golang.org/x/crypto/cryptobyte"
)

type serverHelloBasic struct {
	Vers        uint16
	CipherSuite uint16
	Extensions  []uint16
}

func parseServerHelloBasic(data []byte) (*serverHelloBasic, error) {
	if len(data) < 9 {
		return nil, errors.New("server returned short message")
	}

	serverHelloLen := int(data[6])<<16 | int(data[7])<<8 | int(data[8])
	if serverHelloLen >= len(data) || len(data) < 9+serverHelloLen {
		return nil, errors.New("invalid server hello length")
	}
	data = data[5 : 9+serverHelloLen]

	hello := &serverHelloBasic{}
	s := cryptobyte.String(data)

	var (
		vers              uint16
		random            []byte
		sessionID         []byte
		cipherSuite       uint16
		compressionMethod uint8
	)
	if !s.Skip(4) ||
		!s.ReadUint16(&vers) ||
		!readUint8LengthPrefixed(&s, &random) ||
		!readUint8LengthPrefixed(&s, &sessionID) ||
		!s.ReadUint16(&cipherSuite) ||
		!s.ReadUint8(&compressionMethod) {
		return nil, errors.New("invalid server hello message")
	}

	hello.Vers = vers
	hello.CipherSuite = cipherSuite

	if s.Empty() {
		return hello, nil
	}

	var extensions cryptobyte.String
	if !s.ReadUint16LengthPrefixed(&extensions) || !s.Empty() {
		return nil, errors.New("failed to read server hello extensions")
	}

	for !extensions.Empty() {
		var extension uint16
		var extData cryptobyte.String
		if !extensions.ReadUint16(&extension) ||
			!extensions.ReadUint16LengthPrefixed(&extData) {
			return nil, errors.New("failed to read server hello extension")
		}
		hello.Extensions = append(hello.Extensions, extension)
	}

	return hello, nil
}

func readUint8LengthPrefixed(s *cryptobyte.String, out *[]byte) bool {
	return s.ReadUint8LengthPrefixed((*cryptobyte.String)(out))
}