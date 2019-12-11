package blowfish

import (
	"Carmel/secret"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_OneBlock(t *testing.T) {
	var tests = []struct {
		plain0, plain1   uint32
		cipher0, cipher1 uint32
	}{
		{1, 2, uint32(0xdf333fd2), uint32(0x30a71bb4)},
	}

	//key := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	key := []byte("TESTKEY")
	bf := New(key)

	for _, test := range tests {
		e0, e1 := bf.encryptBlock(test.plain0, test.plain1)
		assert.Equal(t, e0, test.cipher0)
		assert.Equal(t, e1, test.cipher1)

		a0, a1 := bf.decryptBlock(test.cipher0, test.cipher1)
		assert.Equal(t, a0, test.plain0)
		assert.Equal(t, a1, test.plain1)
	}
}

// Test vectors from https://www.schneier.com/code/vectors.txt
func Test_BlockECB(t *testing.T) {
	var tests = []struct {
		key    []byte
		plain  []byte
		cipher []byte
	}{
		{
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0x4E, 0xF9, 0x97, 0x45, 0x61, 0x98, 0xDD, 0x78},
		},
		{
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			[]byte{0x51, 0x86, 0x6F, 0xD5, 0xB8, 0x5E, 0xCB, 0x8A},
		},
		{
			[]byte{0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			[]byte{0x7D, 0x85, 0x6F, 0x9A, 0x61, 0x30, 0x63, 0xF2},
		},
		{
			[]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
			[]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
			[]byte{0x24, 0x66, 0xDD, 0x87, 0x8B, 0x96, 0x3C, 0x9D},
		},
		{
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			[]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
			[]byte{0x61, 0xF9, 0xC3, 0x80, 0x22, 0x81, 0xB0, 0x96},
		},
		{
			[]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			[]byte{0x7D, 0x0C, 0xC6, 0x30, 0xAF, 0xDA, 0x1E, 0xC7},
		},
		{
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0x4E, 0xF9, 0x97, 0x45, 0x61, 0x98, 0xDD, 0x78},
		},
		{
			[]byte{0xFE, 0xDC, 0xBA, 0x98, 0x76, 0x54, 0x32, 0x10},
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			[]byte{0x0A, 0xCE, 0xAB, 0x0F, 0xC6, 0xA0, 0xA2, 0x8D},
		},
		{
			[]byte{0x7C, 0xA1, 0x10, 0x45, 0x4A, 0x1A, 0x6E, 0x57},
			[]byte{0x01, 0xA1, 0xD6, 0xD0, 0x39, 0x77, 0x67, 0x42},
			[]byte{0x59, 0xC6, 0x82, 0x45, 0xEB, 0x05, 0x28, 0x2B},
		},
		{
			[]byte{0x01, 0x31, 0xD9, 0x61, 0x9D, 0xC1, 0x37, 0x6E},
			[]byte{0x5C, 0xD5, 0x4C, 0xA8, 0x3D, 0xEF, 0x57, 0xDA},
			[]byte{0xB1, 0xB8, 0xCC, 0x0B, 0x25, 0x0F, 0x09, 0xA0},
		},
		{
			[]byte{0x07, 0xA1, 0x13, 0x3E, 0x4A, 0x0B, 0x26, 0x86},
			[]byte{0x02, 0x48, 0xD4, 0x38, 0x06, 0xF6, 0x71, 0x72},
			[]byte{0x17, 0x30, 0xE5, 0x77, 0x8B, 0xEA, 0x1D, 0xA4},
		},
		{
			[]byte{0x38, 0x49, 0x67, 0x4C, 0x26, 0x02, 0x31, 0x9E},
			[]byte{0x51, 0x45, 0x4B, 0x58, 0x2D, 0xDF, 0x44, 0x0A},
			[]byte{0xA2, 0x5E, 0x78, 0x56, 0xCF, 0x26, 0x51, 0xEB},
		},
		{
			[]byte{0x04, 0xB9, 0x15, 0xBA, 0x43, 0xFE, 0xB5, 0xB6},
			[]byte{0x42, 0xFD, 0x44, 0x30, 0x59, 0x57, 0x7F, 0xA2},
			[]byte{0x35, 0x38, 0x82, 0xB1, 0x09, 0xCE, 0x8F, 0x1A},
		},
		{
			[]byte{0x01, 0x13, 0xB9, 0x70, 0xFD, 0x34, 0xF2, 0xCE},
			[]byte{0x05, 0x9B, 0x5E, 0x08, 0x51, 0xCF, 0x14, 0x3A},
			[]byte{0x48, 0xF4, 0xD0, 0x88, 0x4C, 0x37, 0x99, 0x18},
		},
		{
			[]byte{0x01, 0x70, 0xF1, 0x75, 0x46, 0x8F, 0xB5, 0xE6},
			[]byte{0x07, 0x56, 0xD8, 0xE0, 0x77, 0x47, 0x61, 0xD2},
			[]byte{0x43, 0x21, 0x93, 0xB7, 0x89, 0x51, 0xFC, 0x98},
		},
		{
			[]byte{0x43, 0x29, 0x7F, 0xAD, 0x38, 0xE3, 0x73, 0xFE},
			[]byte{0x76, 0x25, 0x14, 0xB8, 0x29, 0xBF, 0x48, 0x6A},
			[]byte{0x13, 0xF0, 0x41, 0x54, 0xD6, 0x9D, 0x1A, 0xE5},
		},
		{
			[]byte{0x07, 0xA7, 0x13, 0x70, 0x45, 0xDA, 0x2A, 0x16},
			[]byte{0x3B, 0xDD, 0x11, 0x90, 0x49, 0x37, 0x28, 0x02},
			[]byte{0x2E, 0xED, 0xDA, 0x93, 0xFF, 0xD3, 0x9C, 0x79},
		},
		{
			[]byte{0x04, 0x68, 0x91, 0x04, 0xC2, 0xFD, 0x3B, 0x2F},
			[]byte{0x26, 0x95, 0x5F, 0x68, 0x35, 0xAF, 0x60, 0x9A},
			[]byte{0xD8, 0x87, 0xE0, 0x39, 0x3C, 0x2D, 0xA6, 0xE3},
		},
		{
			[]byte{0x37, 0xD0, 0x6B, 0xB5, 0x16, 0xCB, 0x75, 0x46},
			[]byte{0x16, 0x4D, 0x5E, 0x40, 0x4F, 0x27, 0x52, 0x32},
			[]byte{0x5F, 0x99, 0xD0, 0x4F, 0x5B, 0x16, 0x39, 0x69},
		},
		{
			[]byte{0x1F, 0x08, 0x26, 0x0D, 0x1A, 0xC2, 0x46, 0x5E},
			[]byte{0x6B, 0x05, 0x6E, 0x18, 0x75, 0x9F, 0x5C, 0xCA},
			[]byte{0x4A, 0x05, 0x7A, 0x3B, 0x24, 0xD3, 0x97, 0x7B},
		},
		{
			[]byte{0x58, 0x40, 0x23, 0x64, 0x1A, 0xBA, 0x61, 0x76},
			[]byte{0x00, 0x4B, 0xD6, 0xEF, 0x09, 0x17, 0x60, 0x62},
			[]byte{0x45, 0x20, 0x31, 0xC1, 0xE4, 0xFA, 0xDA, 0x8E},
		},
		{
			[]byte{0x02, 0x58, 0x16, 0x16, 0x46, 0x29, 0xB0, 0x07},
			[]byte{0x48, 0x0D, 0x39, 0x00, 0x6E, 0xE7, 0x62, 0xF2},
			[]byte{0x75, 0x55, 0xAE, 0x39, 0xF5, 0x9B, 0x87, 0xBD},
		},
		{
			[]byte{0x49, 0x79, 0x3E, 0xBC, 0x79, 0xB3, 0x25, 0x8F},
			[]byte{0x43, 0x75, 0x40, 0xC8, 0x69, 0x8F, 0x3C, 0xFA},
			[]byte{0x53, 0xC5, 0x5F, 0x9C, 0xB4, 0x9F, 0xC0, 0x19},
		},
		{
			[]byte{0x4F, 0xB0, 0x5E, 0x15, 0x15, 0xAB, 0x73, 0xA7},
			[]byte{0x07, 0x2D, 0x43, 0xA0, 0x77, 0x07, 0x52, 0x92},
			[]byte{0x7A, 0x8E, 0x7B, 0xFA, 0x93, 0x7E, 0x89, 0xA3},
		},
		{
			[]byte{0x49, 0xE9, 0x5D, 0x6D, 0x4C, 0xA2, 0x29, 0xBF},
			[]byte{0x02, 0xFE, 0x55, 0x77, 0x81, 0x17, 0xF1, 0x2A},
			[]byte{0xCF, 0x9C, 0x5D, 0x7A, 0x49, 0x86, 0xAD, 0xB5},
		},
		{
			[]byte{0x01, 0x83, 0x10, 0xDC, 0x40, 0x9B, 0x26, 0xD6},
			[]byte{0x1D, 0x9D, 0x5C, 0x50, 0x18, 0xF7, 0x28, 0xC2},
			[]byte{0xD1, 0xAB, 0xB2, 0x90, 0x65, 0x8B, 0xC7, 0x78},
		},
		{
			[]byte{0x1C, 0x58, 0x7F, 0x1C, 0x13, 0x92, 0x4F, 0xEF},
			[]byte{0x30, 0x55, 0x32, 0x28, 0x6D, 0x6F, 0x29, 0x5A},
			[]byte{0x55, 0xCB, 0x37, 0x74, 0xD1, 0x3E, 0xF2, 0x01},
		},
		{
			[]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			[]byte{0xFA, 0x34, 0xEC, 0x48, 0x47, 0xB2, 0x68, 0xB2},
		},
		{
			[]byte{0x1F, 0x1F, 0x1F, 0x1F, 0x0E, 0x0E, 0x0E, 0x0E},
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			[]byte{0xA7, 0x90, 0x79, 0x51, 0x08, 0xEA, 0x3C, 0xAE},
		},
		{
			[]byte{0xE0, 0xFE, 0xE0, 0xFE, 0xF1, 0xFE, 0xF1, 0xFE},
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			[]byte{0xC3, 0x9E, 0x07, 0x2D, 0x9F, 0xAC, 0x63, 0x1D},
		},
		{
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			[]byte{0x01, 0x49, 0x33, 0xE0, 0xCD, 0xAF, 0xF6, 0xE4},
		},
		{
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0xF2, 0x1E, 0x9A, 0x77, 0xB7, 0x1C, 0x49, 0xBC},
		},
		{
			[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF},
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			[]byte{0x24, 0x59, 0x46, 0x88, 0x57, 0x54, 0x36, 0x9A},
		},
		{
			[]byte{0xFE, 0xDC, 0xBA, 0x98, 0x76, 0x54, 0x32, 0x10},
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			[]byte{0x6B, 0x5C, 0x5A, 0x9C, 0x5D, 0x9E, 0x0A, 0x5A},
		},
	}

	for _, test := range tests {
		bf := New(test.key)
		c := bf.EncryptECB(test.plain)
		ok := secret.AreSlicesEqual(test.cipher, c)
		assert.True(t, ok)

		p := bf.DecryptECB(c)
		ok = secret.AreSlicesEqual(test.plain, p)
		assert.True(t, ok)
	}
}

func Test_CBC(t *testing.T) {
	iv := secret.RandomBytes(blockSize)
	key := secret.RandomBytes(MaxKeyLength)

	bf := New(key)

	plain := []byte("Beesoft Software, Piotr Pszczółkowski")
	cipher := bf.EncryptCBC(plain, iv)
	ok := secret.AreSlicesEqual(plain, bf.DecryptCBC(cipher))
	assert.True(t, ok)
}
