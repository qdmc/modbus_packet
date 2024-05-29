package modbus_tcp_packet

import (
	"errors"
	"github.com/qdmc/modbus_packet/packets"
	"io"
)

type ReadRtuDataHandle packets.ReadRtuDataHandle

type PacketInterface packets.PacketInterface

type PacketHandle func(PacketInterface, error)

type CodecInterface interface {
	ReadOnce(r io.Reader) (PacketInterface, error)
	Read(reader io.Reader, handle PacketHandle) error
	WritePacket(w io.Writer, p PacketInterface) (int, error)
	GetPacketMod() packets.PacketMod
	SetPacketMod(m uint8)
	SetRtdHandler(h ReadRtuDataHandle)
}

func NewCodec(mod ...uint8) CodecInterface {
	c := &defaultCodec{
		rtuHandle: nil,
		mod:       packets.TCP,
	}
	if mod != nil && len(mod) >= 1 {
		c.SetPacketMod(mod[0])
	}
	return c
}

type defaultCodec struct {
	rtuHandle ReadRtuDataHandle
	mod       packets.PacketMod
}

func (c *defaultCodec) ReadOnce(r io.Reader) (PacketInterface, error) {
	if c.mod == packets.RTU {
		return c.readRtu(r)
	} else if c.mod == packets.TCP {
		return c.readTcp(r)
	} else {
		return c.readAscii(r)
	}
}

func (c *defaultCodec) Read(r io.Reader, handle PacketHandle) error {
	if handle == nil {
		return errors.New("callback Handle is not set")
	}
	if c.mod == packets.RTU {
		if c.rtuHandle == nil {
			return errors.New("ptuPacket readDataHandle is not set")
		}
		go func() {
			for {
				p, err := c.readRtu(r)
				if err != nil {
					handle(nil, err)
					return
				}
				go handle(p, nil)
			}
		}()
	} else if c.mod == packets.TCP {
		go func() {
			for {
				p, err := c.readTcp(r)
				if err != nil {
					handle(nil, err)
					return
				}
				go handle(p, nil)
			}
		}()

	} else {
		go func() {
			for {
				p, err := c.readAscii(r)
				if err != nil {
					handle(nil, err)
					return
				}
				go handle(p, nil)
			}
		}()
	}
	return nil
}

func (c *defaultCodec) WritePacket(w io.Writer, p PacketInterface) (int, error) {
	bs, err := p.ToBytes()
	if err != nil {
		return 0, err
	}
	return w.Write(bs)
}

func (c *defaultCodec) GetPacketMod() packets.PacketMod {
	return c.mod
}
func (c *defaultCodec) SetPacketMod(m uint8) {
	switch m {
	case 0:
		c.mod = packets.RTU
		break
	case 1:
		c.mod = packets.TCP
		break
	case 2:
		c.mod = packets.ASCII
		break
	default:
		return
	}
}
func (c *defaultCodec) SetRtdHandler(h ReadRtuDataHandle) {
	c.rtuHandle = h
}

func (c *defaultCodec) readRtu(r io.Reader) (PacketInterface, error) {
	p := new(packets.RtuPacket)
	err := p.Read(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *defaultCodec) readTcp(r io.Reader) (PacketInterface, error) {
	p := new(packets.TcpPacket)
	err := p.Read(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (c *defaultCodec) readAscii(r io.Reader) (PacketInterface, error) {
	p := new(packets.AsciiPacket)
	err := p.Read(r)
	if err != nil {
		return nil, err
	}
	return p, nil
}
