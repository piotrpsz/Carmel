/*
 * BSD 2-Clause License
 *
 *	Copyright (c) 2019, Piotr Pszczółkowski
 *	All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 * list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 * this list of conditions and the following disclaimer in the documentation
 * and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

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
