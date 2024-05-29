package packets

import (
	"errors"
	"github.com/qdmc/go_stream_readWriter"
	"io"
)

type TcpPacket struct {
	TransactionId uint16
	ProtocolId    uint16
	Length        uint16
	UnitId        uint8
	FunctionCode  uint8
	Data          []byte
}

func (p *TcpPacket) GetFunctionCode() uint8 {
	return p.FunctionCode
}

func (p *TcpPacket) SetFunctionCode(id uint8) {
	p.FunctionCode = id
}

func (p *TcpPacket) GetData() []byte {
	return p.Data
}

func (p *TcpPacket) SetData(d []byte) {
	p.Data = d
}

func (p *TcpPacket) GetMod() PacketMod {
	return TCP
}
func (p *TcpPacket) ToBytes() ([]byte, error) {
	w := go_stream_readWriter.NewWriter()
	if p.Data != nil && len(p.Data) > 0 {
		if len(p.Data) > 253 {
			return nil, errors.New("packet data is too long")
		}
		p.Length = 2 + uint16(len(p.Data))
	} else {
		p.Length = 2
	}
	var err error
	_, err = w.WriteUint16(p.TransactionId)
	if err != nil {
		return nil, err
	}
	_, err = w.WriteUint16(p.ProtocolId)
	if err != nil {
		return nil, err
	}
	_, err = w.WriteUint16(p.Length)
	if err != nil {
		return nil, err
	}
	_, err = w.WriteUint8(p.UnitId)
	if err != nil {
		return nil, err
	}
	_, err = w.WriteUint8(p.FunctionCode)
	if err != nil {
		return nil, err
	}
	if p.Length > 2 {
		_, err = w.WriteBytes(p.Data)
		if err != nil {
			return nil, err
		}
	}
	return w.BufBytes(), nil
}
func (p *TcpPacket) Read(r io.Reader) error {
	var err error
	reader := go_stream_readWriter.NewReader()
	p.TransactionId, err = reader.ReadUint16(r)
	if err != nil {
		return err
	}
	p.ProtocolId, err = reader.ReadUint16(r)
	if err != nil {
		return err
	}
	p.Length, err = reader.ReadUint16(r)
	if err != nil {
		return err
	}
	p.UnitId, err = reader.ReadUint8(r)
	if err != nil {
		return err
	}
	p.FunctionCode, err = reader.ReadUint8(r)
	if err != nil {
		return err
	}
	p.Data, err = reader.ReadBytes(r, uint(p.Length))
	if err != nil {
		return err
	}
	return nil
}
