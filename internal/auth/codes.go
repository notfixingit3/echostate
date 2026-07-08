package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"
)

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func CodePepper() string {
	if v := strings.TrimSpace(os.Getenv("ECHOSTATE_AUTH_PEPPER")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("ECHOSTATE_BREAK_GLASS_SECRET")); v != "" {
		return v
	}
	return "echostate-dev-pepper"
}

func HashCode(code string) string {
	mac := hmac.New(sha256.New, []byte(CodePepper()))
	_, _ = mac.Write([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(mac.Sum(nil))
}

func GenerateNumericCode(length int) (string, error) {
	if length <= 0 {
		length = 8
	}
	var b strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	return b.String(), nil
}

func GenerateRecoveryCode(length int) (string, error) {
	if length < 8 {
		length = 12
	}
	var b strings.Builder
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b.WriteByte(codeAlphabet[n.Int64()])
	}
	return b.String(), nil
}

func NormalizeEnrollmentInput(code string) string {
	return strings.TrimSpace(code)
}

func FormatCodeHint(length int, recovery bool) string {
	if recovery {
		return fmt.Sprintf("%d-character recovery code", length)
	}
	return fmt.Sprintf("%d-digit enrollment code", length)
}
