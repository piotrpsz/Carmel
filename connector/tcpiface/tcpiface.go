package tcpiface

import (
	"Carmel/shared/tr"
	"bufio"
	"net"
)

type TcpInterface struct {
	writer *net.TCPConn
	reader *bufio.Reader
}

func New(conn *net.TCPConn) *TcpInterface {
	if conn != nil {
		if reader := bufio.NewReader(conn); reader != nil {
			return &TcpInterface{writer: conn, reader: reader}
		}
	}
	return nil
}

func (iface *TcpInterface) Write(data []byte) bool {
	if _, err := iface.writer.Write(data); tr.IsOK(err) {
		return true
	}
	return false
}

func (iface *TcpInterface) Read(bytesNumber int) []byte {
	buffer := make([]byte, bytesNumber)
	if _, err := iface.reader.Read(buffer); tr.IsOK(err) {
		return buffer
	}
	return nil
}
