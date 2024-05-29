package packets

import (
	"encoding/binary"
	"errors"
	"github.com/qdmc/go_stream_readWriter"
	"io"
)

type ReadRtuDataHandle func(fid uint8, reader io.Reader) ([]byte, error)

type RtuPacket struct {
	SlaveId      uint8
	FunctionCode uint8
	Data         []byte
	CRC          uint16
	dataHandle   ReadRtuDataHandle
}

func (p *RtuPacket) SetHandle(h ReadRtuDataHandle) {
	p.dataHandle = h
}
func (p *RtuPacket) GetFunctionCode() uint8 {
	return p.FunctionCode
}

func (p *RtuPacket) SetFunctionCode(id uint8) {
	p.FunctionCode = id
}

func (p *RtuPacket) GetData() []byte {
	return p.Data
}

func (p *RtuPacket) SetData(d []byte) {
	p.Data = d
}

func (p *RtuPacket) GetMod() PacketMod {
	return RTU
}
func (p *RtuPacket) ToBytes() ([]byte, error) {
	var err error
	w := go_stream_readWriter.NewWriter()
	_, err = w.WriteUint8(p.SlaveId)
	if err != nil {
		return nil, err
	}
	_, err = w.WriteUint8(p.FunctionCode)
	if err != nil {
		return nil, err
	}
	if p.Data != nil && len(p.Data) > 0 {
		crcCode := crc16(p.Data)
		_, err = w.WriteBytes(p.Data)
		if err != nil {
			return nil, err
		}
		_, err = w.WriteUint16(crcCode)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = w.WriteUint16(uint16(0xFFFF))
		if err != nil {
			return nil, err
		}
	}
	return w.BufBytes(), nil
}
func (p *RtuPacket) CheckCrc() error {
	if p.CRC != crc16(p.Data) {
		return errors.New("crc code is error")
	}
	return nil
}

func (p *RtuPacket) Read(r io.Reader) error {
	var err error
	if p.dataHandle == nil {
		return errors.New("data handle is not set")
	}
	reader := go_stream_readWriter.NewReader()
	p.SlaveId, err = reader.ReadUint8(r)
	if err != nil {
		return err
	}
	p.FunctionCode, err = reader.ReadUint8(r)
	if err != nil {
		return err
	}
	p.Data, err = p.dataHandle(p.FunctionCode, r)
	if err != nil {
		return err
	}
	p.CRC, err = reader.ReadUint16(r)
	if err != nil {
		return err
	}
	err = p.CheckCrc()
	if err != nil {
		return err
	}
	return nil
}
func crc16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	if data == nil || len(data) < 1 {
		return crc
	}
	for _, b := range data {
		// 将当前字节与CRC值进行异或操作
		crc ^= uint16(b)
		// 对CRC值进行8次右移操作
		for i := 0; i < 8; i++ {
			// 如果最低位为1，则进行异或运算
			if crc&0x0001 != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				// 如果最低位为0，则直接右移
				crc >>= 1
			}
		}
	}
	crcBytes := make([]byte, 2)
	crcBytes[0] = byte(crc & 0xFF)
	crcBytes[1] = byte((crc >> 8) & 0xFF)
	return binary.BigEndian.Uint16(crcBytes)
}
