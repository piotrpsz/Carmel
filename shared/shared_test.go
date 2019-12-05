package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_OnlyDigits(t *testing.T) {
	var tests = []struct {
		item string
		want bool
	}{
		{"", false},
		{"6", true},
		{"12345", true},
		{"98765543", true},
		{"987a5543", false},
		{"9875543x", false},
		{"w", false},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, OnlyDigits(test.item))
	}
}

func Test_OnlyHexDigits(t *testing.T) {
	var tests = []struct {
		item string
		want bool
	}{
		{"0", true},
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", true},
		{"5", true},
		{"6", true},
		{"7", true},
		{"8", true},
		{"9", true},
		{"a", true},
		{"A", true},
		{"b", true},
		{"B", true},
		{"c", true},
		{"d", true},
		{"e", true},
		{"f", true},
		{"g", false},
		{"x", false},
		{"Y", false},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, OnlyHexDigits(test.item))
	}
}

func Test_IsValidIPAddress(t *testing.T) {
	var tests = []struct {
		address string
		want    bool
	}{
		{"", false},
		{"1.2", false},
		{"1.2.3", false},
		{"1.2.3.4", true},
		{"1.2.3,4", false},
		{"1.2.3.4.5", false},
		{"123.456.56.128", false},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, IsValidIPAddress(test.address))
	}
}

func Test_IsValidIPName(t *testing.T) {
	var tests = []struct {
		name string
		want bool
	}{
		{"", false},
		{"piotr", true},
		{"artur", true},
		{"pio tr", false},
		{"Piotr", false},
		{"p2io8tr", true},
		{"7piotr", false},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, IsValidName(test.name))
	}
}
