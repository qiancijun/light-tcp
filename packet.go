package ltcp

import (
	"bytes"
	"encoding/gob"
)

type PacketType uint8

const (
	PacketTypeData PacketType = iota
	PacketTypeAck
	PacketTypePing
	PacketTypeClose
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
