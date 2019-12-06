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
