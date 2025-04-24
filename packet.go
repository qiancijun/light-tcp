package ltcp

import (
	"bytes"
	"encoding/gob"
)

type PacketType uint8

const (
	PackTypeData PacketType = iota
	PacketData
	PacketAck
)

type PacketList struct {
	Next *PacketList
	Data []byte
}

type Packet struct {
	Type    PacketType // 包类型
	Seq     uint32     // 序列号
	Ack     uint32     // 确认号
	Payload []byte     // 载荷
}

func NewPacket(typ PacketType, data []byte) *Packet {
	packet := &Packet{
		Type:    typ,
		Payload: data,
	}
	// TODO: 根据 typ 封装信息
	return packet
}

// 将 Packet 序列化为字节数组
// TODO: 目前使用 gob 实现序列化，后续优化实现自定义序列化方式
func (p *Packet) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 从字节数组反序列化为 Packet
func DeserializePacket(data []byte) (*Packet, error) {
	var p Packet
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}
	// TODO: 做一些校验工作
	return &p, nil
}
