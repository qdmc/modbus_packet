package packets

import "io"

type PacketMod uint8

const (
	RTU   PacketMod = 0
	TCP   PacketMod = 1
	ASCII PacketMod = 2
)

type PacketInterface interface {
	GetFunctionCode() uint8
	SetFunctionCode(id uint8)
	GetData() []byte
	SetData(d []byte)
	GetMod() PacketMod
	ToBytes() ([]byte, error)
	Read(r io.Reader) error
}
