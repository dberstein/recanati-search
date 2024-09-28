package doc

import (
	"crypto/sha256"
	"fmt"
)

func GetID(content []byte) string {
	h := sha256.New()
	h.Write(content)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
