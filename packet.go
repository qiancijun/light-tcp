package ltcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

type PacketType uint8

const (
	MinPacketLength            = 17
	PacketTypeData  PacketType = iota
	PacketTypeAck
	PacketTypePing
	PacketTypeClose
)

var (
	ErrPacketNotEnouthLength = errors.New("packet has not enough length")
	ErrPacketCorrupted       = errors.New("packet has been corrupted")
)

type PacketNode struct {
	Prev, Next *PacketNode
	Packet     *Packet
}

type Packet struct {
	Type    PacketType // 包类型
	Seq     uint32     // 序列号
	Ack     uint32     // 确认号
	Tick    uint32
	Payload []byte // 载荷
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
func (p *Packet) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 写入 Byte
	if err := buf.WriteByte(byte(p.Type)); err != nil {
		return nil, err
	}

	// 写入 Seq，Ack，Tick
	fileds := []uint32{p.Seq, p.Ack, p.Tick}
	for _, f := range fileds {
		if err := binary.Write(buf, binary.BigEndian, f); err != nil {
			return nil, err
		}
	}

	// 写入 Payload
	if _, err := buf.Write(p.Payload); err != nil {
		return nil, err
	}

	// 计算 CRC
	crc := crc32.ChecksumIEEE(buf.Bytes())
	if err := binary.Write(buf, binary.BigEndian, crc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 从字节数组反序列化为 Packet
func DeserializePacket(data []byte) (*Packet, error) {
	if len(data) < MinPacketLength {
		return nil, ErrPacketNotEnouthLength
	}

	// 校验 CRC
	payloadWithoutCRC := data[:len(data)-4]
	expectCRC := binary.BigEndian.Uint32(data[len(data)-4:])
	calculatedCRC := crc32.ChecksumIEEE(payloadWithoutCRC)
	if calculatedCRC != expectCRC {
		return nil, ErrPacketCorrupted
	}

	p := &Packet{}
	buf := bytes.NewReader(payloadWithoutCRC)

	// 读 Type
	typeByte, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	p.Type = PacketType(typeByte)

	// 读 Seq，Ack，Tick
	fields := []*uint32{&p.Seq, &p.Ack, &p.Tick}
	for _, f := range fields {
		if err := binary.Read(buf, binary.BigEndian, f); err != nil {
			return nil, err
		}
	}

	// 填充 Payload
	p.Payload, err = io.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// 迭代器
type PacketIterator struct {
	current *PacketNode
	tail    *PacketNode
}

func (it *PacketIterator) HasNext() bool {
	return it.current != nil && it.current != it.tail
}

func (it *PacketIterator) Next() *Packet {
	if !it.HasNext() {
		return nil
	}
	data := it.current.Packet
	it.current = it.current.Next
	return data
}

type PacketQueue struct {
	Head, Tail *PacketNode
	cnt        int // 统计节点总个数
}

func NewPacketQueue() *PacketQueue {
	head, tail := &PacketNode{}, &PacketNode{}
	head.Next, tail.Prev = tail, head
	return &PacketQueue{
		Head: head,
		Tail: tail,
		cnt:  0,
	}
}

// 尾插，意味着尾部是最新的数据
func (q *PacketQueue) push(p *Packet) {
	tail := q.Tail.Prev
	newNode := &PacketNode{
		Packet: p,
	}
	tail.Next, newNode.Prev = newNode, tail
	newNode.Next, q.Tail.Prev = q.Tail, newNode
	// 单线程操作，不需要考虑并发问题
	// 一个连接会绑定各自的 ltcp
	q.cnt++
}

// 头删
func (q *PacketQueue) pop() *Packet {
	if q.Head.Next == q.Tail {
		return nil
	}
	head := q.Head
	popNode := head.Next
	head.Next = popNode.Next
	popNode.Next.Prev = head
	q.cnt--
	return popNode.Packet
}

func (q *PacketQueue) HasValue() bool {
	return q.Head.Next != q.Tail
}

// 迭代器
func (pq *PacketQueue) Iterator() *PacketIterator {
	return &PacketIterator{
		current: pq.Head.Next,
		tail:    pq.Tail,
	}
}
