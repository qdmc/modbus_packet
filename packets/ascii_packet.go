package packets

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/qdmc/go_stream_readWriter"
	"io"
)

const (
	AsciiStartCode     = ":"
	AsciiEndCode       = "\r\n"
	AsciiChars         = "0123456789ABCDEF"
	AsciiDataLengthMax = 503
)

var charsMap = map[uint8]struct{}{
	'0':  {},
	'1':  {},
	'2':  {},
	'3':  {},
	'4':  {},
	'5':  {},
	'6':  {},
	'7':  {},
	'8':  {},
	'9':  {},
	'A':  {},
	'B':  {},
	'C':  {},
	'D':  {},
	'E':  {},
	'F':  {},
	'\r': {},
	'\n': {},
}

type AsciiPacket struct {
	//startCode    string
	SlaveId      uint8
	FunctionCode uint8
	Data         []byte
	LRC          uint16
	//endCode      string
}

func (p *AsciiPacket) GetFunctionCode() uint8 {
	return p.FunctionCode
}

func (p *AsciiPacket) SetFunctionCode(id uint8) {
	p.FunctionCode = id
}

func (p *AsciiPacket) GetData() []byte {
	return p.Data
}

func (p *AsciiPacket) SetData(d []byte) {
	p.Data = d
}

func (p *AsciiPacket) GetMod() PacketMod {
	return ASCII
}

func (p *AsciiPacket) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	var err error
	lrcBs := []byte{p.SlaveId, p.FunctionCode}
	lrcBs = append(lrcBs, p.Data...)
	// 计算LRC8校验码
	lrc := lrc8(lrcBs)
	lrcBs = append(lrcBs, lrc)
	// 打码
	pduBs := asciiEncode(lrcBs)
	_, err = buf.WriteString(AsciiStartCode)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(pduBs)
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteString(AsciiEndCode)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *AsciiPacket) Read(r io.Reader) error {
	var err error
	var parentByte byte
	w := go_stream_readWriter.NewReader()
	startCode, err := w.ReadUint8(r)
	if err != nil {
		return err
	}
	if string(startCode) != AsciiStartCode {
		return errors.New("start code is error")
	}
	p.SlaveId, err = w.ReadUint8(r)
	if err != nil {
		return err
	}
	p.FunctionCode, err = w.ReadUint8(r)
	if err != nil {
		return err
	}
	var dataBytes []byte
	// 读取data与结束符
	for index := 0; ; index++ {
		readByte, err := w.ReadUint8(r)
		if err != nil {
			return err
		}
		isOver, err := checkDataByte(readByte, parentByte)
		if err != nil {
			return err
		}
		dataBytes = append(dataBytes, readByte)
		if isOver {
			break
		}
		if index >= AsciiDataLengthMax {
			return errors.New("packet data length is out")
		}
		parentByte = readByte
	}
	if len(dataBytes) > 2 {
		// 解码
		p.Data, err = asciiDecode(dataBytes[0 : len(dataBytes)-2])
		if err != nil {
			return err
		}
	}
	return nil
}

func lrc8(data []byte) uint8 {
	lrc := uint8(0)
	if data == nil && len(data) < 1 {
		return uint8(^lrc) + 1
	}
	for _, b := range data {
		lrc += b
	}
	return uint8(^lrc) + 1
}

// asciiEncode 打码
func asciiEncode(bs []byte) []byte {
	var res []byte
	if bs == nil || len(bs) < 1 {
		return res
	}
	for _, b := range bs {
		res = append(res, AsciiChars[b>>4], AsciiChars[b&0x0f])
	}
	return res
}

// asciiDecode  解码
func asciiDecode(bs []byte) ([]byte, error) {
	var res []byte
	if bs == nil || len(bs) < 1 {
		return res, nil
	}
	bsLen := len(bs)
	if bsLen%2 != 0 {
		return res, errors.New("bytes length  must even")
	}
	index := 0
	for bsLen > 0 {
		hexBs := make([]byte, 1)
		_, err := hex.Decode(hexBs[:], []byte{bs[index], bs[index+1]})
		if err != nil {
			return res, err
		}
		res = append(res, hexBs[0])
		index += 2
		bsLen -= 2
	}
	return res, nil
}

func checkDataByte(b, parent byte) (bool, error) {
	if _, ok := charsMap[b]; !ok {
		return false, errors.New(" ascii chars must be in \"0,1,2,3,4,5,6,7,8,9,A,B,C,D,E,F,\\r,\\n\"")
	}
	if b == '\n' {
		if parent == '\r' {
			return true, nil
		} else {
			return false, errors.New("packet endCode is error")
		}
	}
	return false, nil
}
