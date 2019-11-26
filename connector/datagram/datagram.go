package datagram

import (
	"Carmel/connector/tcpiface"
	"Carmel/secret"
)

func Send(iface *tcpiface.TCPInterface, data []byte) bool {
	n := uint32(len(data))
	if n != 0 {
		bytesNumber := secret.Uint32ToBytes(n)
		return iface.Write(bytesNumber) && iface.Write(data)
	}
	return false
}

func Read(iface *tcpiface.TCPInterface) []byte {
	if bytesNumber := iface.Read(4); bytesNumber != nil {
		n := int(secret.BytesToUint32(bytesNumber))
		if data := iface.Read(n); data != nil {
			return data
		}
	}
	return nil
}
