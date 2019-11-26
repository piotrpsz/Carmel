package secret

import (
	"Carmel/shared/tr"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

func ClearSlice(buffer *[]byte) {
	if buffer != nil && len(*buffer) > 0 {
		subtle.ConstantTimeCopy(1, *buffer, make([]byte, len(*buffer)))
		*buffer = nil
	}
}

func AreSlicesEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

func Padding(nbytes int) []byte {
	buffer := make([]byte, nbytes, nbytes)
	buffer[0] = 128
	return buffer
}

func PaddingIndex(data []byte) int {
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] != 0 {
			if data[i] == 128 {
				return i
			}
			break
		}
	}
	return -1
}

func RandomBytes(size int) []byte {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); tr.IsOK(err) {
		return buffer
	}
	return nil
}

func SliceToHex(data []byte) string {
	ndigits := hex.EncodedLen(len(data))
	buffer := make([]byte, ndigits, ndigits)
	hex.Encode(buffer, data)
	return string(buffer)
}

func BytesAsString(a []byte) string {
	var b strings.Builder

	fmt.Fprintf(&b, "{")
	for i := 0; i < (len(a) - 1); i++ {
		fmt.Fprintf(&b, "0x%x, ", a[i])
	}
	fmt.Fprintf(&b, "0x%x}", a[len(a)-1])
	return b.String()
}

func BytesToUint32(data []byte) uint32 {
	return (uint32(data[3]) << 24) |
		(uint32(data[2]) << 16) |
		(uint32(data[1]) << 8) |
		(uint32(data[0]))
}

func Uint32ToBytes(v uint32) []byte {
	data := [4]byte{}

	data[3] = byte(v>>24) & 0xff
	data[2] = byte(v>>16) & 0xff
	data[1] = byte(v>>8) & 0xff
	data[0] = byte(v & 0xff)

	return data[:]
}
