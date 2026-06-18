package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
)

var ja3sGreaseValues = map[uint16]bool{
	0x0a0a: true, 0x1a1a: true, 0x2a2a: true, 0x3a3a: true,
	0x4a4a: true, 0x5a5a: true, 0x6a6a: true, 0x7a7a: true,
	0x8a8a: true, 0x9a9a: true, 0xaaaa: true, 0xbaba: true,
	0xcaca: true, 0xdada: true, 0xeaea: true, 0xfafa: true,
}

func digestJA3S(hello *serverHelloBasic) string {
	if hello == nil {
		return ""
	}
	sum := md5.Sum(ja3sBare(hello))
	return hex.EncodeToString(sum[:])
}

func ja3sBare(hello *serverHelloBasic) []byte {
	buffer := make([]byte, 0, 64)
	buffer = strconv.AppendInt(buffer, int64(hello.Vers), 10)
	buffer = append(buffer, ',')
	buffer = strconv.AppendInt(buffer, int64(hello.CipherSuite), 10)
	buffer = append(buffer, ',')

	if len(hello.Extensions) > 0 {
		last := len(hello.Extensions) - 1
		for _, ext := range hello.Extensions[:last] {
			if ja3sGreaseValues[ext] {
				continue
			}
			buffer = strconv.AppendInt(buffer, int64(ext), 10)
			buffer = append(buffer, '-')
		}
		if !ja3sGreaseValues[hello.Extensions[last]] {
			buffer = strconv.AppendInt(buffer, int64(hello.Extensions[last]), 10)
		}
	}

	return buffer
}