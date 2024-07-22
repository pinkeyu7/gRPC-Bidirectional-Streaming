package helper

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
)

func Sha1Str(str string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(str))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func Md5Str(str string) (string, error) {
	m := md5.New()
	_, err := io.WriteString(m, str)
	hash := fmt.Sprintf("%x", m.Sum(nil))
	return hash, err
}
