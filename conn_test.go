package ltcp

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
	p := NewPacket(PacketData, []byte("Hello"))
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
