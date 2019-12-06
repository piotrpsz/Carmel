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
