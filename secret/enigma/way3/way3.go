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

package way3

import (
	"Carmel/secret"
	"log"
)

const (
	nmbr      = 11 // number of rounds
	blockSize = 12 // in bytes
	KeySize   = 12 // in bytes
)

var (
	ercon = [12]uint32{0x0b0b, 0x1616, 0x2c2c, 0x5858, 0xb0b0, 0x7171, 0xe2e2, 0xd5d5, 0xbbbb, 0x6767, 0xcece, 0x8d8d}
	drcon = [12]uint32{0xb1b1, 0x7373, 0xe6e6, 0xdddd, 0xabab, 0x4747, 0x8e8e, 0x0d0d, 0x1a1a, 0x3434, 0x6868, 0xd0d0}
)

type Way3 struct {
	k  [3]uint32
	ki [3]uint32
}

// New - creates 3-Way object and initiates its members (keys).
// As parameter user must pass a key.
// The key size not equal 96 bit is treated as an error.
func New(key []byte) *Way3 {
	if len(key) != KeySize {
		log.Printf("Invalid key length. Is %d bit. Should be 96 bits (12 bytes).\n", len(key)*8)
		return nil
	}
	tw := new(Way3)
	k0, k1, k2 := tw.bytes2block(key)
	tw.keyGenerator(k0, k1, k2)
	return tw
}

func (tw *Way3) encryptBlock(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	for i := 0; i < nmbr; i++ {
		a0 ^= tw.k[0] ^ (ercon[i] << 16)
		a1 ^= tw.k[1]
		a2 ^= tw.k[2] ^ ercon[i]
		a0, a1, a2 = rho(a0, a1, a2)
	}
	a0 ^= tw.k[0] ^ (ercon[nmbr] << 16)
	a1 ^= tw.k[1]
	a2 ^= tw.k[2] ^ ercon[nmbr]

	return theta(a0, a1, a2)
}

func (tw *Way3) decryptBlock(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	a0, a1, a2 = mu(a0, a1, a2)

	for i := 0; i < nmbr; i++ {
		a0 ^= tw.ki[0] ^ (drcon[i] << 16)
		a1 ^= tw.ki[1]
		a2 ^= tw.ki[2] ^ drcon[i]
		a0, a1, a2 = rho(a0, a1, a2)
	}
	a0 ^= tw.ki[0] ^ (drcon[nmbr] << 16)
	a1 ^= tw.ki[1]
	a2 ^= tw.ki[2] ^ drcon[nmbr]

	return mu(theta(a0, a1, a2))
}

func (tw *Way3) EncryptECB(plainText []byte) []byte {
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
		a0, a1, a2 := tw.bytes2block(plainText[i:])
		c0, c1, c2 := tw.encryptBlock(a0, a1, a2)
		tw.block2bytes(c0, c1, c2, buffer[i:])
	}
	return buffer
}

func (tw *Way3) DecryptECB(cipherText []byte) []byte {
	if cipherText == nil {
		return nil
	}
	nbytes := len(cipherText)
	buffer := make([]byte, nbytes)

	for i := 0; i < nbytes; i += blockSize {
		c0, c1, c2 := tw.bytes2block(cipherText[i:])
		a0, a1, a2 := tw.decryptBlock(c0, c1, c2)
		tw.block2bytes(a0, a1, a2, buffer[i:])
	}

	if idx := secret.PaddingIndex(buffer); idx != -1 {
		return buffer[:idx]
	}
	return buffer
}

func (tw *Way3) EncryptCBC(plainText, iv []byte) []byte {
	if plainText == nil || len(plainText) == 0 {
		return nil
	}
	nbytes := len(plainText)
	n := nbytes % blockSize
	if n != 0 {
		dn := blockSize - n
		plainText = append(plainText, secret.Padding(dn)...)
		nbytes += dn
	}
	if iv == nil {
		tiv := secret.RandomBytes(blockSize)
		if tiv == nil {
			return nil
		}
		iv = tiv
	}

	buffer := make([]byte, nbytes+blockSize)
	for i := 0; i < blockSize; i++ {
		buffer[i] = iv[i]
	}

	a0, a1, a2 := tw.bytes2block(iv)
	for i := 0; i < nbytes; i += blockSize {
		t0, t1, t2 := tw.bytes2block(plainText[i:])
		a0, a1, a2 = tw.encryptBlock(t0^a0, t1^a1, t2^a2)
		tw.block2bytes(a0, a1, a2, buffer[(i+blockSize):])
	}
	return buffer
}

func (tw *Way3) DecryptCBC(cipherText []byte) []byte {
	if cipherText == nil || len(cipherText) < blockSize {
		return nil
	}
	nbytes := len(cipherText)

	buffer := make([]byte, nbytes-blockSize)
	p0, p1, p2 := tw.bytes2block(cipherText)
	for i := blockSize; i < nbytes; i += blockSize {
		a0, a1, a2 := tw.bytes2block(cipherText[i:])
		t0, t1, t2 := a0, a1, a2
		c0, c1, c2 := tw.decryptBlock(a0, a1, a2)
		tw.block2bytes(c0^p0, c1^p1, c2^p2, buffer[(i-blockSize):])
		p0, p1, p2 = t0, t1, t2
	}

	if idx := secret.PaddingIndex(buffer); idx != -1 {
		return buffer[:idx]
	}
	return buffer
}

func (tw *Way3) keyGenerator(k0, k1, k2 uint32) {
	// key
	tw.k[0], tw.k[1], tw.k[2] = k0, k1, k2
	// inverse key
	tw.ki[0], tw.ki[1], tw.ki[2] = mu(theta(k0, k1, k2))
}

//
// mu - inverts the order of the bits of a
//
func mu(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	w0, w1, w2 := uint32(0), uint32(0), uint32(0)

	for i := 0; i < 32; i++ {
		w0 <<= 1
		w1 <<= 1
		w2 <<= 1
		w0 |= a2 & 1
		w1 |= a1 & 1
		w2 |= a0 & 1

		a0 >>= 1
		a1 >>= 1
		a2 >>= 1
	}
	return w0, w1, w2
}

// gamma - the nonlinear step
func gamma(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	w0 := (^a0) ^ ((^a1) & a2)
	w1 := (^a1) ^ ((^a2) & a0)
	w2 := (^a2) ^ ((^a0) & a1)
	return w0, w1, w2
}

// theta - the linear step
func theta(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	w0 := a0 ^
		(a0 >> 16) ^ (a1 << 16) ^
		(a1 >> 16) ^ (a2 << 16) ^
		(a1 >> 24) ^ (a2 << 8) ^
		(a2 >> 8) ^ (a0 << 24) ^
		(a2 >> 16) ^ (a0 << 16) ^
		(a2 >> 24) ^ (a0 << 8)

	w1 := a1 ^
		(a1 >> 16) ^ (a2 << 16) ^
		(a2 >> 16) ^ (a0 << 16) ^
		(a2 >> 24) ^ (a0 << 8) ^
		(a0 >> 8) ^ (a1 << 24) ^
		(a0 >> 16) ^ (a1 << 16) ^
		(a0 >> 24) ^ (a1 << 8)

	w2 := a2 ^
		(a2 >> 16) ^ (a0 << 16) ^
		(a0 >> 16) ^ (a1 << 16) ^
		(a0 >> 24) ^ (a1 << 8) ^
		(a1 >> 8) ^ (a2 << 24) ^
		(a1 >> 16) ^ (a2 << 16) ^
		(a1 >> 24) ^ (a2 << 8)

	return w0, w1, w2
}

func pi1(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	w0 := (a0 >> 10) ^ (a0 << 22)
	w1 := a1
	w2 := (a2 << 1) ^ (a2 >> 31)

	return w0, w1, w2
}

func pi2(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	w0 := (a0 << 1) ^ (a0 >> 31)
	w1 := a1
	w2 := (a2 >> 10) ^ (a2 << 22)

	return w0, w1, w2
}

// rho - the round function
func rho(a0, a1, a2 uint32) (uint32, uint32, uint32) {
	return pi2(gamma(pi1(theta(a0, a1, a2))))
}

func (tw *Way3) bytes2block(data []byte) (uint32, uint32, uint32) {
	w0 := (uint32(data[3]) << 24) | (uint32(data[2]) << 16) | (uint32(data[1]) << 8) | uint32(data[0])
	w1 := (uint32(data[7]) << 24) | (uint32(data[6]) << 16) | (uint32(data[5]) << 8) | uint32(data[4])
	w2 := (uint32(data[11]) << 24) | (uint32(data[10]) << 16) | (uint32(data[9]) << 8) | uint32(data[8])
	return w0, w1, w2
}

func (tw *Way3) block2bytes(a0, a1, a2 uint32, output []byte) {
	output[3] = byte((a0 >> 24) & 0xff)
	output[2] = byte((a0 >> 16) & 0xff)
	output[1] = byte((a0 >> 8) & 0xff)
	output[0] = byte(a0 & 0xff)

	output[7] = byte((a1 >> 24) & 0xff)
	output[6] = byte((a1 >> 16) & 0xff)
	output[5] = byte((a1 >> 8) & 0xff)
	output[4] = byte(a1 & 0xff)

	output[11] = byte((a2 >> 24) & 0xff)
	output[10] = byte((a2 >> 16) & 0xff)
	output[9] = byte((a2 >> 8) & 0xff)
	output[8] = byte(a2 & 0xff)
}
