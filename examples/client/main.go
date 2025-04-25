package main

import (
	"fmt"
	"net"

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
	ltcpConn := ltcp.NewConn(conn, ltcp.DefaultLtcpConnOptions)

	datas := [][]byte{
		[]byte("Hello"),
		[]byte(","),
		[]byte("World"),
	}

	for _, data := range datas {
		_, err := ltcpConn.Write(data)
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	ltcpConn.Close()
}
