package main

import (
	"fmt"
	"net"
	"time"

	"github.com/qiancijun/ltcp"
)

// Client 客户端

func main() {
	raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	laddr, err := net.ResolveUDPAddr("udp", ":")
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		panic(err)
	}

	// p := ltcp.NewPacket(ltcp.PacketTypeData, []byte("Hello"))
	// data, err := p.Serialize()
	// if err != nil {
	// 	panic(err)
	// }
	// conn.Write(data)
	// return

	ltcpConn := ltcp.NewConn(conn, ltcp.DefaultLtcpConnOptions)
	fmt.Printf("client %s connect to server %s\n", laddr.String(), raddr.String())

	datas := [][]byte{
		[]byte("Hello"),
		[]byte(","),
		[]byte("World"),
	}

	for _, data := range datas {
		fmt.Println("ready to send data to server, data len: ", len(data))
		_, err := ltcpConn.Write(data)
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	time.Sleep(1 * time.Second)

	ltcpConn.Close()
}
