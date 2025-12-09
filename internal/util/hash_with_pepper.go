package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashWithPepper(pepper, value string) string {
	h := sha256.New()
	h.Write([]byte(pepper))
	h.Write([]byte("::"))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}
