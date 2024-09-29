package doc

import (
	"crypto/sha256"
	"fmt"
)

func Sha256(content []byte) string {
	h := sha256.New()
	h.Write(content)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
