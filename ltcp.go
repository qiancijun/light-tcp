package ltcp

import "fmt"

/**
* LTCP 在 UDP 上实现了以下功能
* 	1. 数据包编号
* 	2. ACK
*	3. 超时重传
*	4. 流量控制
 */

type Ltcp struct {
	// 发送缓冲区
	sendQueue *PacketQueue
	// 接收缓冲区
	recvQueue *PacketQueue

	sendId       uint32
	currentTick  uint32
	lastRecvTick uint32
}

func NewLtcp() *Ltcp {
	ltcp := &Ltcp{
		sendQueue:   NewPacketQueue(),
		recvQueue:   NewPacketQueue(),
		sendId:      0,
		currentTick: 0,
	}
	return ltcp
}

// 整理数据，为数据添加序列号，tick 等信息
func (ltcp *Ltcp) Collect(bts []byte) error {
	packet := NewPacket(PacketTypeData, bts)
	packet.Seq = ltcp.sendId
	ltcp.sendId++
	packet.Tick = ltcp.currentTick
	ltcp.sendQueue.push(packet)
	return nil
}

// 将数据封装成 Packet
func (ltcp *Ltcp) Package(tick int) {
	if ltcp.sendQueue.cnt == 0 {
		// 没有数据要发送，就包装一个 PING 请求发送
		pingPacket := NewPacket(PacketTypePing, []byte{})
		ltcp.sendQueue.push(pingPacket)
	}
	ltcp.currentTick += uint32(tick)
}

// TODO: 解析数据包
// 一些可靠性的逻辑
func (ltcp *Ltcp) Parse(bts []byte) error {
	packet, err := DeserializePacket(bts)
	if err != nil {
		// 接受到了一个不可解析的包
		return err
	}
	switch packet.Type {
	case PacketTypeData:
		// 接受到的包是一个数据包
		// 更新一下最后收到数据的 Tick
		ltcp.lastRecvTick = ltcp.currentTick
		// 加入到接收缓冲区
		fmt.Println("收到一个 Data 包")
		ltcp.recvQueue.push(packet)
	case PacketTypePing:
		// 心跳包
	default:
		// TODO: 接受到了一个未知类型的包
		return nil
	}
	return nil
}

// 从缓冲区中读取数据
func (ltcp *Ltcp) Recv(buf []byte) (int, error) {
	// 从读取缓冲区中读取一个 Packet
	// TODO: 目前没有处理序号的问题
	packet := ltcp.recvQueue.pop()
	if packet == nil {
		return 0, nil
	}
	copy(buf, packet.Payload)
	return len(packet.Payload), nil
}
