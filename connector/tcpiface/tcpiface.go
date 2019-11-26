package tcpiface

import (
	"Carmel/shared/tr"
	"bufio"
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
