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

package vtc

type (
	RoleType            uint8
	OperationStatusType uint8
	MessageType         uint8
)

const (
	_ MessageType = iota
	Request
	Answer
)

const (
	_ uint32 = iota
	Login
)

const (
	_ RoleType = iota
	Server
	Client
)

const (
	_ OperationStatusType = iota
	Ok
	Timeout
	Error
	Cancel
	SecurityBreach
)

const (
	SignatureSize  int     = 512
	MessageTimeout float64 = 60 // w sekundach (1 min)
)

var RandomBytes = []byte{
	// 64 bytes
	0xa0, 0x80, 0x32, 0xff, 0xee, 0x3c, 0x6, 0x8f, 0xec, 0x24, 0x32, 0x1d,
	0x95, 0x90, 0xe6, 0x3e, 0xd4, 0x9b, 0xd6, 0x5, 0xe6, 0x21, 0x51, 0xd0,
	0xb0, 0x3c, 0x99, 0x5d, 0x93, 0xc1, 0x4a, 0x7a, 0xae, 0xa, 0x99, 0x74,
	0x3d, 0x4c, 0xb0, 0x3e, 0xa8, 0x82, 0xb5, 0xcf, 0x72, 0x64, 0x8d, 0x59,
	0xfe, 0x50, 0xce, 0x82, 0x6e, 0xa6, 0x7b, 0x98, 0x5e, 0xef, 0x57, 0x29,
	0xa7, 0xa9, 0x7a, 0xcb,
	// 128 bytes
	0x47, 0xde, 0x4a, 0x71, 0xa, 0x8b, 0x9, 0x3e, 0x37, 0x48, 0xc2, 0xf4,
	0xff, 0xbd, 0xc4, 0x35, 0x9f, 0x5f, 0x49, 0xba, 0x42, 0xa1, 0x25, 0x22,
	0x63, 0x1f, 0x41, 0x9d, 0x0, 0xbd, 0x2f, 0x60, 0xfe, 0x19, 0xa5, 0xc6,
	0xba, 0x1b, 0x5d, 0x11, 0x1a, 0xd0, 0xb5, 0xf0, 0x73, 0x42, 0xa7, 0x64,
	0xe8, 0xfb, 0x24, 0xb4, 0xb7, 0x81, 0xde, 0x43, 0x2f, 0xb6, 0x66, 0x80,
	0xab, 0xab, 0x96, 0x44, 0x11, 0x5, 0xfa, 0xc1, 0x56, 0xa, 0x41, 0x7d,
	0xb9, 0x34, 0xe2, 0x11, 0xd5, 0x1f, 0x54, 0x4c, 0xda, 0xab, 0x51, 0x8f,
	0x49, 0x6a, 0x9e, 0x22, 0x98, 0x6f, 0x4e, 0x90, 0xd3, 0x8a, 0xcb, 0xc0,
	0x64, 0xcc, 0x11, 0x87, 0xa7, 0x97, 0xab, 0xe8, 0x2a, 0xbf, 0x1e, 0xc7,
	0xb9, 0xfd, 0x80, 0x17, 0xb5, 0x81, 0xc2, 0x35, 0xf5, 0x2e, 0x91, 0xaf,
	0xd8, 0xbd, 0x79, 0xce, 0x21, 0x5b, 0x31, 0x1e,
	// 32 bytes
	0xde, 0xa9, 0x38, 0x3a, 0x85, 0x44, 0x43, 0x2e, 0x10, 0xb5, 0x66, 0x80,
	0xff, 0x47, 0x6a, 0x25, 0x48, 0xd7, 0x90, 0xfe, 0x59, 0xce, 0x15, 0x5,
	0xf3, 0xd9, 0xfa, 0xf8, 0xf1, 0x8f, 0xe7, 0x76,
	// 128 bytes
	0xbb, 0x36, 0x2a, 0xb4, 0x84, 0x7f, 0xbd, 0xc6, 0x2a, 0x1f, 0xbd, 0xf3,
	0x9f, 0x4c, 0xc, 0xa6, 0xb5, 0xc3, 0xbb, 0xf3, 0x50, 0xb5, 0x2c, 0x27,
	0x29, 0xd1, 0x5a, 0x88, 0x4c, 0x21, 0x5f, 0x88, 0xd4, 0xbb, 0xe9, 0x7,
	0xb6, 0x7a, 0x49, 0xca, 0x79, 0xd9, 0x4c, 0xa2, 0x7d, 0x4d, 0x62, 0x3e,
	0xef, 0x31, 0x9c, 0x3, 0x4f, 0xc5, 0x51, 0x64, 0x1f, 0x20, 0x2c, 0x44,
	0xd, 0x76, 0x6c, 0x87, 0x2b, 0xa0, 0x91, 0x33, 0xae, 0x4a, 0x15, 0xbc,
	0xd7, 0xda, 0x82, 0xbf, 0xc8, 0xbc, 0x5a, 0x3b, 0x24, 0xdc, 0xaf, 0xfe,
	0xe6, 0x16, 0x5, 0x33, 0x14, 0x71, 0xbc, 0xab, 0x9c, 0x5b, 0xbd, 0xfd,
	0x2d, 0xcf, 0x3a, 0xcf, 0xdf, 0x84, 0xe2, 0x8e, 0x2d, 0x74, 0x29, 0xf6,
	0x78, 0x6c, 0x84, 0xab, 0x6e, 0xa6, 0x32, 0x17, 0xab, 0xbd, 0x33, 0xdf,
	0x21, 0x23, 0xb4, 0xdc, 0xd1, 0xd0, 0x90, 0x8d,
	// 32 bytes
	0xe, 0xfe, 0xb8, 0xb, 0x41, 0x56, 0x45, 0xfe, 0x6c, 0xa8, 0x3f, 0x60,
	0xa3, 0x4, 0x4, 0x60, 0x3f, 0xb5, 0x85, 0xf4, 0x78, 0xc3, 0x3d, 0x12,
	0x4e, 0xc7, 0x2f, 0x5, 0xfe, 0xd6, 0x74, 0xfe,
}

type Keys struct {
	Blowfish []byte `json:"blowfish,omitempty"`
	Ghost    []byte `json:"ghost,omitempty"`
	Way3     []byte `json:"way3, omitempty"`
}
