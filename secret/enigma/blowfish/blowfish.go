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

package blowfish

import (
	"Carmel/secret"
	"crypto/subtle"
	"log"
)

const (
	blockSize    = 8 // in bytes (64-bit, two uint32 words)
	roundCount   = 16
	MinKeyLength = 4  // in bytes
	MaxKeyLength = 56 // in bytes
)

type Blowfish struct {
	p [roundCount + 2]uint32
	s [4][256]uint32
}

func New(key []byte) *Blowfish {
	keyLen := len(key)
	if keyLen < MinKeyLength || keyLen > MaxKeyLength {
		log.Printf("Blowfish error. Invalid key length. Is %d bit, should be 32..448 bit.\n", 8*keyLen)
		return nil
	}

	bf := new(Blowfish)

	// S - init
	for i := 0; i < 4; i++ {
		for j := 0; j < 256; j++ {
			bf.s[i][j] = orgS[i][j]
		}
	}

	// P - init
	k := 0
	for i := 0; i < (roundCount + 2); i++ {
		data := uint32(0)
		for j := 0; j < 4; j++ {
			data = (data << 8) | uint32(key[k])
			k += 1
			if k >= keyLen {
				k = 0
			}
		}
		bf.p[i] = orgP[i] ^ data
	}

	xl, xr := uint32(0), uint32(0)

	// P
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[0], bf.p[1] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[2], bf.p[3] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[4], bf.p[5] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[6], bf.p[7] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[8], bf.p[9] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[10], bf.p[11] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[12], bf.p[13] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[14], bf.p[15] = xl, xr
	xl, xr = bf.encryptBlock(xl, xr)
	bf.p[16], bf.p[17] = xl, xr

	// S
	for i := 0; i < 4; i++ {
		for j := 0; j < 256; j += 2 {
			xl, xr = bf.encryptBlock(xl, xr)
			bf.s[i][j] = xl
			bf.s[i][j+1] = xr
		}
	}

	return bf
}

// EncryptCBC - implementation with random IV described in (pages 66-67)
// "Cryptography Engineering - Design, Principles and Practical Applications"
// Niels Fergusson, Bruce Schneier, Tadayoshi Kohno
func (bf *Blowfish) EncryptCBC(plainText, iv []byte) []byte {
	nbytes := len(plainText)
	n := nbytes % blockSize
	if n != 0 {
		dn := blockSize - n
		plainText = append(plainText, secret.Padding(dn)...)
		nbytes = len(plainText)
	}
	if iv == nil {
		iv = secret.RandomBytes(blockSize)
		if iv == nil {
			return nil
		}
	}
	if len(iv) != blockSize {
		return nil
	}

	buffer := make([]byte, nbytes+blockSize)
	subtle.ConstantTimeCopy(1, buffer[:blockSize], iv)

	n1, n2 := bf.bytes2block(iv)
	for i := 0; i < nbytes; i += blockSize {
		t1, t2 := bf.bytes2block(plainText[i:])
		n1, n2 = bf.encryptBlock(t1^n1, t2^n2)
		bf.block2bytes(n1, n2, buffer[(i+blockSize):])
	}
	return buffer
}

func (bf *Blowfish) DecryptCBC(cipherText []byte) []byte {
	if cipherText == nil {
		return nil
	}

	nbytes := len(cipherText)
	buffer := make([]byte, nbytes-blockSize)

	p1, p2 := bf.bytes2block(cipherText)
	for i := blockSize; i < nbytes; i += blockSize {
		n1, n2 := bf.bytes2block(cipherText[i:])
		t1, t2 := n1, n2

		c1, c2 := bf.decryptBlock(n1, n2)

		bf.block2bytes(c1^p1, c2^p2, buffer[(i-blockSize):])
		p1, p2 = t1, t2
	}

	if padIdx := secret.PaddingIndex(buffer); padIdx != -1 {
		return buffer[:padIdx]
	}
	return buffer
}

func (bf *Blowfish) EncryptECB(plainText []byte) []byte {
	if plainText == nil {
		return nil
	}

	nbytes := len(plainText)
	n := nbytes % blockSize
	if n != 0 {
		plainText = append(plainText, secret.Padding(blockSize-n)...)
		nbytes = len(plainText)
	}

	buffer := make([]byte, nbytes)
	for i := 0; i < nbytes; i += blockSize {
		xl, xr := bf.bytes2block(plainText[i:])
		xl, xr = bf.encryptBlock(xl, xr)
		bf.block2bytes(xl, xr, buffer[i:])
	}
	return buffer
}

func (bf *Blowfish) DecryptECB(cipherText []byte) []byte {
	if cipherText == nil {
		return nil
	}
	nbytes := len(cipherText)

	buffer := make([]byte, nbytes)
	for i := 0; i < nbytes; i += blockSize {
		xl, xr := bf.bytes2block(cipherText[i:])
		xl, xr = bf.decryptBlock(xl, xr)
		bf.block2bytes(xl, xr, buffer[i:])
	}

	if idx := secret.PaddingIndex(buffer); idx != -1 {
		buffer = buffer[:idx]
	}
	return buffer
}

func (bf *Blowfish) f(x uint32) uint32 {
	d := x & 0xff
	x >>= 8
	c := x & 0xff
	x >>= 8
	b := x & 0xff
	x >>= 8
	a := x & 0xff
	return ((bf.s[0][a] + bf.s[1][b]) ^ bf.s[2][c]) + bf.s[3][d]
}

func (bf *Blowfish) encryptBlock(xl, xr uint32) (uint32, uint32) {
	xl = xl ^ bf.p[0]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[1]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[2]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[3]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[4]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[5]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[6]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[7]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[8]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[9]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[10]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[11]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[12]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[13]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[14]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[15]
	xl = bf.f(xr) ^ xl

	return xr ^ bf.p[17], xl ^ bf.p[16]
}

func (bf *Blowfish) decryptBlock(xl, xr uint32) (uint32, uint32) {
	xl = xl ^ bf.p[17]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[16]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[15]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[14]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[13]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[12]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[11]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[10]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[9]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[8]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[7]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[6]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[5]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[4]
	xl = bf.f(xr) ^ xl

	xl = xl ^ bf.p[3]
	xr = bf.f(xl) ^ xr
	xr = xr ^ bf.p[2]
	xl = bf.f(xr) ^ xl

	return xr ^ bf.p[0], xl ^ bf.p[1]
}

func (bf *Blowfish) bytes2block(data []byte) (uint32, uint32) {
	w0 := uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
	w1 := uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
	return w0, w1
}

func (bf *Blowfish) block2bytes(a0, a1 uint32, o []byte) {
	o[7] = byte(a1)
	o[6] = byte(a1 >> 8)
	o[5] = byte(a1 >> 16)
	o[4] = byte(a1 >> 24)
	o[3] = byte(a0)
	o[2] = byte(a0 >> 8)
	o[1] = byte(a0 >> 16)
	o[0] = byte(a0 >> 24)
}
