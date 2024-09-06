package helper

import (
	"crypto/rand"
	"math/big"
)

func RandomInt(minNum, maxNum int) int {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(maxNum)))
	if err != nil {
		panic(err)
	}
	n := int(nBig.Int64())
	return n + minNum
}
