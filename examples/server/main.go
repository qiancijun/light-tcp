package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/qiancijun/ltcp"
)

// Server 服务端

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	listener := ltcp.NewLtcpListener(conn)
	fmt.Println("server run on port 8080")

	go func() {
		for {
			rconn, err := listener.AcceptLtcpConn()
			if err != nil {
				fmt.Println("accept conn error: ", err)
				continue
			}
			log.Println("server accept a new connection: ", rconn.RemoteAddr().String())
			go read(rconn)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT)
	select {
	case <-signalChan:
	}
}

func read(conn *ltcp.LtcpConn) {
	log.Println("read data from conn...")
	for {
		data := make([]byte, 32767)
		n, err := conn.Read(data)
		if err != nil {
			fmt.Println("read error: ", err)
			break
		}
		fmt.Println("server receive data: ", string(data[:n]))
	}
}
