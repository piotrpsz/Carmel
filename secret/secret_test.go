package secret

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_AreSlicesEqual(t *testing.T) {
	var tests = []struct {
		set0 []byte
		set1 []byte
		want bool
	}{
		{[]byte("12345"), []byte("12345"), true},
		{[]byte("12045"), []byte("12345"), false},
		{[]byte("1234"), []byte("12345"), false},
		{[]byte("12345"), []byte("1234"), false},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, AreSlicesEqual(test.set0, test.set1))
	}
}

func Test_Padding(t *testing.T) {
	var tests = []struct {
		n    int
		want []byte
	}{
		{1, []byte{128}},
		{3, []byte{128, 0, 0}},
		{5, []byte{128, 0, 0, 0, 0}},
		{10, []byte{128, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	}

	for _, test := range tests {
		result := Padding(test.n)
		assert.True(t, AreSlicesEqual(result, test.want))
	}
}
