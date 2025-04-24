package ltcp

import (
	"fmt"
	"net"
)

type (
	OnData func(packet *Packet, addr *net.UDPAddr)
)

type UDPServer struct {
	Addr   string
	Conn   *net.UDPConn
	OnData OnData
}

func NewUDPServer(addr string, onData OnData) *UDPServer {
	return &UDPServer{
		Addr:   addr,
		OnData: onData,
	}
}

// 异步方法
func (s *UDPServer) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	s.Conn = conn
	fmt.Println("UDP Server started on ", s.Addr)

	go s.run()

	return nil
}

func (s *UDPServer) StartSync() {
	udpAddr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		fmt.Println("resolve addr error: ", err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("listen addr error: ", err)
		return
	}
	s.Conn = conn
	fmt.Println("UDP Server started on ", s.Addr)

	s.run()
}

func (s *UDPServer) run() {
	for {
		// 接受数据包
		packet, clientAddr, err := s.receivePacket(s.Conn)
		if err != nil {
			fmt.Println("receive packet error: ", err)
			continue
		}

		// 打印一下 packet
		if s.OnData != nil {
			s.OnData(packet, clientAddr)
		}

		// 发送响应包
		responsePacket := NewPacket(PacketAck, []byte("Hello"), DefaultOptions)
		err = s.sendPacket(s.Conn, clientAddr, responsePacket)
		if err != nil {
			fmt.Println("sending packet error: ", err)
		}
	}
}

func (s *UDPServer) sendPacket(conn *net.UDPConn, addr *net.UDPAddr, packet *Packet) error {
	data, err := packet.Serialize()
	if err != nil {
		return err
	}

	_, err = conn.WriteToUDP(data, addr)
	return err
}

func (s *UDPServer) receivePacket(conn *net.UDPConn) (*Packet, *net.UDPAddr, error) {
	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, err
	}

	// 反序列化为 Packet
	packet, err := DeserializePacket(buf[:n])
	if err != nil {
		return nil, nil, err
	}

	return packet, addr, nil
}

func (s *UDPServer) Close() {
	s.Conn.Close()
}
