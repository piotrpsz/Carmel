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

package tcpiface

import (
	"Carmel/shared/tr"
	"bufio"
	"fmt"
	"net"
)

type TCPInterface struct {
	writer *net.TCPConn
	reader *bufio.Reader
}

func New(conn *net.TCPConn) *TCPInterface {
	if conn != nil {
		if reader := bufio.NewReader(conn); reader != nil {
			return &TCPInterface{writer: conn, reader: reader}
		}
	}
	return nil
}

func (iface *TCPInterface) Close() {
	defer func() {
		iface.writer = nil
		iface.reader = nil
	}()

	if iface.writer != nil {
		iface.writer.Close()
	}
}

func (iface *TCPInterface) Write(data []byte) bool {
	if _, err := iface.writer.Write(data); tr.IsOK(err) {
		return true
	}
	return false
}

func (iface *TCPInterface) Read(bytesNumber int) []byte {
	buffer := make([]byte, bytesNumber)
	if _, err := iface.reader.Read(buffer); tr.IsOK(err) {
		return buffer
	}
	return nil
}

func (iface *TCPInterface) Address() string {
	return fmt.Sprintf("local: %s -- remote: %s", iface.writer.LocalAddr(), iface.writer.RemoteAddr())
}
