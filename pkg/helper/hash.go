package helper

import (
	"crypto/sha256"
	"fmt"
)

func Sha1Str(str string) string {
	data := []byte(str)
	hash := sha256.New()
	hash.Write(data)
	hashed := hash.Sum(nil)
	return fmt.Sprintf("%x", hashed)
}
