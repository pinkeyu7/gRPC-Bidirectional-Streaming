package helper

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

func Sha1Str(str string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(str))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func ScryptStr(str string) string {
	salt := os.Getenv("SCRYPT_SALT")
	secret := []byte(salt)
	dk, _ := scrypt.Key([]byte(str), secret, 16384, 8, 1, 32)
	return fmt.Sprintf("%x", dk)
}

func Md5Str(str string) (string, error) {
	m := md5.New()
	_, err := io.WriteString(m, str)
	hash := fmt.Sprintf("%x", m.Sum(nil))
	return hash, err
}

func TaxBillStatementHash(merId int, ym string) string {
	return strings.ToUpper(Sha1Str(fmt.Sprintf("taxbi-%d-%s-statement", merId, ym)))
}

func InvoiceFeeBillHash(areaType string, areaCode int, ym string) string {
	return strings.ToUpper(Sha1Str(fmt.Sprintf("taxbi-%s-%d-%s-invoice-fee-bill", areaType, areaCode, ym)))
}

func InvoiceFeePaymentHash(paymentId int64, areaType string, areaCode int) string {
	return strings.ToUpper(Sha1Str(fmt.Sprintf("taxbi-%d-%s-%d-invoice-fee-payment", paymentId, areaType, areaCode)))
}
