package ghost

import (
	"Carmel/secret"
	"log"
)

const (
	KeySize   = 32 // in bytes
	blockSize = 8  // in bytes (2 x uint32, 8 bytes, 64 bit, )
)

type Gost struct {
	k   [8]uint32 // key - 256 bit
	k87 [256]byte
	k65 [256]byte
	k43 [256]byte
	k21 [256]byte
}

// New - creates Gost object and initiates its members.
// As parameter user must pass a key.
// The key size not equal 256 bit is treated as an error.
func New(key []byte) *Gost {
	if (len(key)) != KeySize {
		log.Printf("Ghost error. Invalid key length. Is %d bit, should be 256 bit.\n", 8*len(key))
		return nil
	}

	k8 := [16]byte{14, 4, 13, 1, 2, 15, 11, 8, 3, 10, 6, 12, 5, 9, 0, 7}
	k7 := [16]byte{15, 1, 8, 14, 6, 11, 3, 4, 9, 7, 2, 13, 12, 0, 5, 10}
	k6 := [16]byte{10, 0, 9, 14, 6, 3, 15, 5, 1, 13, 12, 7, 11, 4, 2, 8}
	k5 := [16]byte{7, 13, 14, 3, 0, 6, 9, 10, 1, 2, 8, 5, 11, 12, 4, 15}
	k4 := [16]byte{2, 12, 4, 1, 7, 10, 11, 6, 8, 5, 3, 15, 13, 0, 14, 9}
	k3 := [16]byte{12, 1, 10, 15, 9, 2, 6, 8, 0, 13, 3, 4, 14, 7, 5, 11}
	k2 := [16]byte{4, 11, 2, 14, 15, 0, 8, 13, 3, 12, 9, 7, 5, 10, 6, 1}
	k1 := [16]byte{13, 2, 8, 4, 6, 15, 11, 1, 10, 9, 3, 14, 5, 0, 12, 7}

	gost := new(Gost)

	for i := 0; i < 256; i++ {
		p1 := i >> 4
		p2 := i & 15
		gost.k87[i] = (k8[p1] << 4) | k7[p2]
		gost.k65[i] = (k6[p1] << 4) | k5[p2]
		gost.k43[i] = (k4[p1] << 4) | k3[p2]
		gost.k21[i] = (k2[p1] << 4) | k1[p2]
	}
	var v uint32
	var idx int
	for i := 0; i < 8; i++ {
		idx = (i * 4) + 3
		v = 0
		v = (v << 8) + uint32(key[idx])
		idx--
		v = (v << 8) + uint32(key[idx])
		idx--
		v = (v << 8) + uint32(key[idx])
		idx--
		v = (v << 8) + uint32(key[idx])
		gost.k[i] = v
	}

	return gost
}

func (gost *Gost) encryptBlock(n1, n2 uint32) (uint32, uint32) {
	n2 ^= gost.f(n1 + gost.k[0])
	n1 ^= gost.f(n2 + gost.k[1])
	n2 ^= gost.f(n1 + gost.k[2])
	n1 ^= gost.f(n2 + gost.k[3])
	n2 ^= gost.f(n1 + gost.k[4])
	n1 ^= gost.f(n2 + gost.k[5])
	n2 ^= gost.f(n1 + gost.k[6])
	n1 ^= gost.f(n2 + gost.k[7])

	n2 ^= gost.f(n1 + gost.k[0])
	n1 ^= gost.f(n2 + gost.k[1])
	n2 ^= gost.f(n1 + gost.k[2])
	n1 ^= gost.f(n2 + gost.k[3])
	n2 ^= gost.f(n1 + gost.k[4])
	n1 ^= gost.f(n2 + gost.k[5])
	n2 ^= gost.f(n1 + gost.k[6])
	n1 ^= gost.f(n2 + gost.k[7])

	n2 ^= gost.f(n1 + gost.k[0])
	n1 ^= gost.f(n2 + gost.k[1])
	n2 ^= gost.f(n1 + gost.k[2])
	n1 ^= gost.f(n2 + gost.k[3])
	n2 ^= gost.f(n1 + gost.k[4])
	n1 ^= gost.f(n2 + gost.k[5])
	n2 ^= gost.f(n1 + gost.k[6])
	n1 ^= gost.f(n2 + gost.k[7])

	n2 ^= gost.f(n1 + gost.k[7])
	n1 ^= gost.f(n2 + gost.k[6])
	n2 ^= gost.f(n1 + gost.k[5])
	n1 ^= gost.f(n2 + gost.k[4])
	n2 ^= gost.f(n1 + gost.k[3])
	n1 ^= gost.f(n2 + gost.k[2])
	n2 ^= gost.f(n1 + gost.k[1])
	n1 ^= gost.f(n2 + gost.k[0])

	return n2, n1
}

func (gost *Gost) decryptBlock(n1, n2 uint32) (uint32, uint32) {
	n2 ^= gost.f(n1 + gost.k[0])
	n1 ^= gost.f(n2 + gost.k[1])
	n2 ^= gost.f(n1 + gost.k[2])
	n1 ^= gost.f(n2 + gost.k[3])
	n2 ^= gost.f(n1 + gost.k[4])
	n1 ^= gost.f(n2 + gost.k[5])
	n2 ^= gost.f(n1 + gost.k[6])
	n1 ^= gost.f(n2 + gost.k[7])

	n2 ^= gost.f(n1 + gost.k[7])
	n1 ^= gost.f(n2 + gost.k[6])
	n2 ^= gost.f(n1 + gost.k[5])
	n1 ^= gost.f(n2 + gost.k[4])
	n2 ^= gost.f(n1 + gost.k[3])
	n1 ^= gost.f(n2 + gost.k[2])
	n2 ^= gost.f(n1 + gost.k[1])
	n1 ^= gost.f(n2 + gost.k[0])

	n2 ^= gost.f(n1 + gost.k[7])
	n1 ^= gost.f(n2 + gost.k[6])
	n2 ^= gost.f(n1 + gost.k[5])
	n1 ^= gost.f(n2 + gost.k[4])
	n2 ^= gost.f(n1 + gost.k[3])
	n1 ^= gost.f(n2 + gost.k[2])
	n2 ^= gost.f(n1 + gost.k[1])
	n1 ^= gost.f(n2 + gost.k[0])

	n2 ^= gost.f(n1 + gost.k[7])
	n1 ^= gost.f(n2 + gost.k[6])
	n2 ^= gost.f(n1 + gost.k[5])
	n1 ^= gost.f(n2 + gost.k[4])
	n2 ^= gost.f(n1 + gost.k[3])
	n1 ^= gost.f(n2 + gost.k[2])
	n2 ^= gost.f(n1 + gost.k[1])
	n1 ^= gost.f(n2 + gost.k[0])

	return n2, n1
}

func (gost *Gost) EncryptECB(plainText []byte) []byte {
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
		n1, n2 := gost.bytes2block(plainText[i:])
		n1, n2 = gost.encryptBlock(n1, n2)
		gost.block2bytes(n1, n2, buffer[i:])
	}

	return buffer
}

func (gost *Gost) DecryptECB(cipherText []byte) []byte {
	if cipherText == nil {
		return nil
	}
	nbytes := len(cipherText)

	buffer := make([]byte, nbytes)
	for i := 0; i < nbytes; i += blockSize {
		n1, n2 := gost.bytes2block(cipherText[i:])
		n1, n2 = gost.decryptBlock(n1, n2)
		gost.block2bytes(n1, n2, buffer[i:])
	}

	if idx := secret.PaddingIndex(buffer); idx != -1 {
		return buffer[:idx]
	}
	return buffer
}

func (gost *Gost) EncryptCBC(plainText, iv []byte) []byte {
	if plainText == nil {
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

	n1, n2 := gost.bytes2block(iv)
	for i := 0; i < nbytes; i += blockSize {
		t1, t2 := gost.bytes2block(plainText[i:])
		n1, n2 = gost.encryptBlock(t1^n1, t2^n2)
		gost.block2bytes(n1, n2, buffer[(i+blockSize):])
	}
	return buffer
}

func (gost *Gost) DecryptCBC(cipherText []byte) []byte {
	if cipherText == nil {
		return nil
	}

	nbytes := len(cipherText)
	buffer := make([]byte, nbytes-blockSize)

	p1, p2 := gost.bytes2block(cipherText)
	for i := blockSize; i < nbytes; i += blockSize {
		n1, n2 := gost.bytes2block(cipherText[i:])
		t1, t2 := n1, n2
		c1, c2 := gost.decryptBlock(n1, n2)
		gost.block2bytes(c1^p1, c2^p2, buffer[(i-blockSize):])
		p1, p2 = t1, t2
	}

	if padIdx := secret.PaddingIndex(buffer); padIdx != -1 {
		return buffer[:padIdx]
	}
	return buffer
}

func (gost *Gost) Clean() {
	gost.k[0] = 0
	gost.k[1] = 0
	gost.k[2] = 0
	gost.k[3] = 0
	gost.k[4] = 0
	gost.k[5] = 0
	gost.k[6] = 0
	gost.k[7] = 0

	for i := 0; i < 256; i++ {
		gost.k87[i] = 0
		gost.k65[i] = 0
		gost.k43[i] = 0
		gost.k21[i] = 0
	}
}

func (gost *Gost) f(x uint32) uint32 {
	w0 := uint32(gost.k87[(x>>24)&0xff]) << 24
	w1 := uint32(gost.k65[(x>>16)&0xff]) << 16
	w2 := uint32(gost.k43[(x>>8)&0xff]) << 8
	w3 := uint32(gost.k21[x&255])
	x = w0 | w1 | w2 | w3
	return (x << 11) | (x >> (32 - 11))
}

func (gost *Gost) bytes2block(data []byte) (uint32, uint32) {
	w0 := (uint32(data[3]) << 24) | (uint32(data[2]) << 16) | (uint32(data[1]) << 8) | uint32(data[0])
	w1 := (uint32(data[7]) << 24) | (uint32(data[6]) << 16) | (uint32(data[5]) << 8) | uint32(data[4])
	return w0, w1
}

func (gost *Gost) block2bytes(a0, a1 uint32, output []byte) {
	output[7] = byte((a1 >> 24) & 0xff)
	output[6] = byte((a1 >> 16) & 0xff)
	output[5] = byte((a1 >> 8) & 0xff)
	output[4] = byte(a1 & 0xff)
	output[3] = byte((a0 >> 24) & 0xff)
	output[2] = byte((a0 >> 16) & 0xff)
	output[1] = byte((a0 >> 8) & 0xff)
	output[0] = byte(a0 & 0xff)
}
