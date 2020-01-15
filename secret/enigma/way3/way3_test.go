package way3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_gamma(t *testing.T) {
	var tests = []struct {
		data []uint32
		want []uint32
	}{
		{[]uint32{0x00000000, 0x00000000, 0x00000000}, []uint32{0xffffffff, 0xffffffff, 0xffffffff}},
		{[]uint32{0x00000001, 0x00000002, 0x00000003}, []uint32{0xffffffff, 0xfffffffd, 0xfffffffe}},
		{[]uint32{0x00000004, 0x00000005, 0x00000006}, []uint32{0xfffffff9, 0xfffffffa, 0xfffffff8}},
		{[]uint32{0xffffffff, 0xffffffff, 0xffffffff}, []uint32{0x00000000, 0x00000000, 0x00000000}},
		{[]uint32{0x01010101, 0x02020202, 0x03030303}, []uint32{0xffffffff, 0xfdfdfdfd, 0xfefefefe}},
		{[]uint32{0x01234567, 0x89abcdef, 0xfedcba98}, []uint32{0x88888888, 0x77777777, 0x89abcdef}},
	}

	for _, test := range tests {
		res0, res1, res2 := gamma(test.data[0], test.data[1], test.data[2])
		assert.Equal(t, res0, test.want[0])
		assert.Equal(t, res1, test.want[1])
		assert.Equal(t, res2, test.want[2])
	}
}

func Test_mu(t *testing.T) {
	var tests = []struct {
		data []uint32
		want []uint32
	}{
		{[]uint32{0x00000000, 0x00000000, 0x00000000}, []uint32{0x00000000, 0x00000000, 0x00000000}},
		{[]uint32{0x00000001, 0x00000002, 0x00000003}, []uint32{0xc0000000, 0x40000000, 0x80000000}},
		{[]uint32{0x00000004, 0x00000005, 0x00000006}, []uint32{0x60000000, 0xa0000000, 0x20000000}},
		{[]uint32{0xffffffff, 0xffffffff, 0xffffffff}, []uint32{0xffffffff, 0xffffffff, 0xffffffff}},
		{[]uint32{0x01010101, 0x02020202, 0x03030303}, []uint32{0xc0c0c0c0, 0x40404040, 0x80808080}},
		{[]uint32{0x01234567, 0x89abcdef, 0xfedcba98}, []uint32{0x195d3b7f, 0xf7b3d591, 0xe6a2c480}},
	}

	for _, test := range tests {
		res0, res1, res2 := mu(test.data[0], test.data[1], test.data[2])
		assert.Equal(t, res0, test.want[0])
		assert.Equal(t, res1, test.want[1])
		assert.Equal(t, res2, test.want[2])
	}
}

func Test_theta(t *testing.T) {
	var tests = []struct {
		data []uint32
		want []uint32
	}{
		{[]uint32{0x00000000, 0x00000000, 0x00000000}, []uint32{0x00000000, 0x00000000, 0x00000000}},
		{[]uint32{0x00000001, 0x00000002, 0x00000003}, []uint32{0x01000201, 0x02000302, 0x03000103}},
		{[]uint32{0x00000004, 0x00000005, 0x00000006}, []uint32{0x04070204, 0x05070105, 0x06070306}},
		{[]uint32{0xffffffff, 0xffffffff, 0xffffffff}, []uint32{0xffffffff, 0xffffffff, 0xffffffff}},
		{[]uint32{0x01010101, 0x02020202, 0x03030303}, []uint32{0x02000003, 0x03000001, 0x01000002}},
		{[]uint32{0x01234567, 0x89abcdef, 0xfedcba98}, []uint32{0xab3210fe, 0xdc321001, 0x23321089}},
	}

	for _, test := range tests {
		res0, res1, res2 := theta(test.data[0], test.data[1], test.data[2])
		assert.Equal(t, res0, test.want[0])
		assert.Equal(t, res1, test.want[1])
		assert.Equal(t, res2, test.want[2])
	}
}

func Test_rho(t *testing.T) {
	var tests = []struct {
		data []uint32
		want []uint32
	}{
		{[]uint32{0x00000000, 0x00000000, 0x00000000}, []uint32{0xffffffff, 0xffffffff, 0xffffffff}},
		{[]uint32{0x00000001, 0x00000002, 0x00000003}, []uint32{0xf77f7ff6, 0x7dbfbcfd, 0xbefeffff}},
		{[]uint32{0x00000004, 0x00000005, 0x00000006}, []uint32{0xededf06e, 0x7bf9ff3a, 0x7dbdfdfe}},
		{[]uint32{0xffffffff, 0xffffffff, 0xffffffff}, []uint32{0x00000000, 0x00000000, 0x00000000}},
		{[]uint32{0x01010101, 0x02020202, 0x03030303}, []uint32{0xfe7efff7, 0xfc3f7ffe, 0xfebfbfff}},
		{[]uint32{0x01234567, 0x89abcdef, 0xfedcba98}, []uint32{0x842224d3, 0x1a47237a, 0xbb1e62f3}},
	}

	for _, test := range tests {
		res0, res1, res2 := rho(test.data[0], test.data[1], test.data[2])
		assert.Equal(t, res0, test.want[0])
		assert.Equal(t, res1, test.want[1])
		assert.Equal(t, res2, test.want[2])
	}
}

func Test_bytes3words(t *testing.T) {
	var tests = []struct {
		data  []uint32
		bytes []byte
	}{
		{
			[]uint32{0x1, 0x2, 0x3},
			[]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03},
		},
		{
			[]uint32{0x12345678, 0x90abcdef, 0x9876fede},
			[]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef, 0x98, 0x76, 0xfe, 0xde},
		},
	}

	for _, test := range tests {
		output := make([]byte, 12)
		block3bytes(test.data[0], test.data[1], test.data[2], output)
		assert.Equal(t, test.bytes, output)

		res0, res1, res2 := bytes3block(output)
		assert.Equal(t, test.data[0], res0)
		assert.Equal(t, test.data[1], res1)
		assert.Equal(t, test.data[2], res2)
	}
}

func Test_encrypt_decrypt(t *testing.T) {
	var tests = []struct {
		key    []uint32
		plain  []uint32
		cipher []uint32
	}{
		{
			[]uint32{0, 0, 0},
			[]uint32{1, 1, 1},
			[]uint32{0x4059c76e, 0x83ae9dc4, 0xad21ecf7},
		},
		{
			[]uint32{6, 5, 4},
			[]uint32{3, 2, 1},
			[]uint32{0xd2f05b5e, 0xd6144138, 0xcab920cd},
		},
		{
			[]uint32{0xdef01234, 0x456789ab, 0xbcdef012},
			[]uint32{0x23456789, 0x9abcdef0, 0x01234567},
			[]uint32{0x0aa55dbb, 0x9cdddb6d, 0x7cdb76b2},
		},
		{
			[]uint32{0xd2f05b5e, 0xd6144138, 0xcab920cd},
			[]uint32{0x4059c76e, 0x83ae9dc4, 0xad21ecf7},
			[]uint32{0x478ea871, 0x6b13f17c, 0x15b155ed},
		},
	}

	key := make([]byte, 12)
	for _, test := range tests {
		block3bytes(test.key[0], test.key[1], test.key[2], key)

		tw := New(key)
		assert.NotNil(t, tw)

		r0, r1, r2 := tw.encryptBlock(test.plain[0], test.plain[1], test.plain[2])
		assert.Equal(t, r0, test.cipher[0])
		assert.Equal(t, r1, test.cipher[1])
		assert.Equal(t, r2, test.cipher[2])

		q0, q1, q2 := tw.decryptBlock(r0, r1, r2)
		assert.Equal(t, q0, test.plain[0])
		assert.Equal(t, q1, test.plain[1])
		assert.Equal(t, q2, test.plain[2])
	}

}
