package ltcp

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createLtcpConn(t *testing.T) *LtcpConn {
	udpAddr, err := net.ResolveUDPAddr("udp", ":8080")
	assert.NoError(t, err)
	udpConn, err := net.ListenUDP("udp", udpAddr)
	assert.NoError(t, err)

	// conn := NewConn(udpConn, DefaultLtcpConnOptions)
	// assert.NotNil(t, conn)
	rUdpAddr, err := net.ResolveUDPAddr("udp", ":8888")
	assert.NoError(t, err)
	conn := NewUnConn(udpConn, rUdpAddr, nil, DefaultLtcpConnOptions)
	assert.NotNil(t, conn)
	return conn
}

func TestConn(t *testing.T) {
	server := NewUDPServer(":8080", func(p *Packet, addr *net.UDPAddr) {
		fmt.Printf("packet type: %d, packet data: %s\n", p.Type, string(p.Payload))
	})

	err := server.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// 构造一个客户端
	severAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	assert.NoError(t, err)
	conn, err := net.DialUDP("udp", nil, severAddr)
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()

	// 构建一个 Packet
	p := NewPacket(PacketTypeData, []byte("Hello"))
	data, err := p.Serialize()
	assert.NoError(t, err)
	_, err = conn.Write(data)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// 读取来自服务端的消息
	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(buf)
	assert.NoError(t, err)
	p, err = DeserializePacket(buf[:n])
	assert.NoError(t, err)
	fmt.Printf("client receive packet from %s, packet type: %d, packet data: %s\n", addr, p.Type, string(p.Payload))
}

func TestConnWrite(t *testing.T) {
	udpAddr, err := net.ResolveUDPAddr("udp", ":8080")
	assert.NoError(t, err)
	udpConn, err := net.ListenUDP("udp", udpAddr)
	assert.NoError(t, err)

	rUdpAddr, err := net.ResolveUDPAddr("udp", ":8888")
	assert.NoError(t, err)
	conn := NewUnConn(udpConn, rUdpAddr, nil, DefaultLtcpConnOptions)
	assert.NotNil(t, conn)

	datas := [][]byte{
		[]byte("Hello"),
		[]byte(","),
		[]byte("World"),
	}

	for _, data := range datas {
		n, err := conn.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
	}

	time.Sleep(1 * time.Second)

	assert.Equal(t, len(datas), conn.ltcp.sendQueue.cnt)

	idx := 0
	for it := conn.ltcp.sendQueue.Iterator(); it.HasNext(); it.Next() {
		// fmt.Println(string(it.current.Packet.Payload))
		assert.Equal(t, datas[idx], it.current.Packet.Payload)
		idx++
	}
}

func TestConnRead(t *testing.T) {
	conn := createLtcpConn(t)

	datas := [][]byte{
		[]byte("Hello"),
		[]byte(","),
		[]byte("World"),
	}

	for _, data := range datas {
		p := NewPacket(PacketTypeData, data)
		bts, err := p.Serialize()
		assert.NoError(t, err)
		conn.in <- bts
	}

	

	idx := 0
	for {
		data := make([]byte, 32767)
		n, err := conn.Read(data)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
		fmt.Println(string(data[:n]))
		assert.Equal(t, len(datas[idx]), n)
		assert.Equal(t, datas[idx], data[:n])
		idx++
	}
	assert.Equal(t, 3, idx)
}

// func TestWriteSeq(t *testing.T) {
// 	conn := createLtcpConn(t)	
// }