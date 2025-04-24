package ltcp

import (
	"fmt"
	"net"
	"time"
)

type LtcpConn struct {
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
	closeFn    func(addr string)

	recvChan chan []byte
	recvErr  chan error

	sendChan chan []byte
	sendErr  chan error
	sendTick chan int

	in chan []byte

	opts LtcpConnOptions
}

type LtcpConnOptions struct {
	// 是否启动自动发送，默认为 true
	// 设置为 false 需要手动调用发送，才会出发网络传输
	AutoSend bool
	// 自动发送的时间间隔
	SendTick time.Duration
	// 每一批的最大发送数据包个数
	MaxSendNumPerTick int
}

var DefaultLtcpConnOptions = LtcpConnOptions{
	AutoSend: true,
	SendTick: time.Millisecond * 20,
}

func NewConn(conn *net.UDPConn, opts LtcpConnOptions) *LtcpConn {
	con := &LtcpConn{
		conn:     conn,
		recvChan: make(chan []byte, 1<<16),
		recvErr:  make(chan error, 2),
		sendChan: make(chan []byte, 1<<16),
		sendErr:  make(chan error, 2),
		sendTick: make(chan int, 2),
		opts:     opts,
	}
	go con.run()
	return con
}

func NewUnConn(conn *net.UDPConn,
	remoteAddr *net.UDPAddr,
	closeFn func(string),
	opts LtcpConnOptions) *LtcpConn {
	con := NewConn(conn, opts)
	con.remoteAddr = remoteAddr
	con.closeFn = closeFn
	return con
}

// 实现一下 net.Conn 接口
func (c *LtcpConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *LtcpConn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return c.remoteAddr
	}
	return c.conn.RemoteAddr()
}

func (c *LtcpConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *LtcpConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *LtcpConn) SetWriteDeadline(t time.Time) error {
	return nil
}

var _ net.Conn = (*LtcpConn)(nil)

// 判断是不是服务端
func (c *LtcpConn) Connected() bool {
	// return c.remoteAddr != nil
	return c.remoteAddr == nil
}

func (c *LtcpConn) Close() error {
	if c.remoteAddr != nil {
		if c.closeFn != nil {
			c.closeFn(c.remoteAddr.String())
		}
		// TODO: 发送中断链接请求
	}
	return nil
}

func (c *LtcpConn) send(bts []byte) error {
	select {
	case c.sendChan <- bts:
		return nil
	case err := <-c.sendErr:
		return err
	}
}

func (c *LtcpConn) Write(bts []byte) (n int, err error) {
	if err := c.send(bts); err != nil {
		return 0, err
	}
	return len(bts), nil
}

func (c *LtcpConn) Read(bts []byte) (n int, err error) {
	select {
	case data := <-c.recvChan:
		copy(bts, data)
		return len(data), nil
	case err := <-c.recvErr:
		return 0, err
	}
}

func (c *LtcpConn) run() {
	if c.opts.AutoSend && c.opts.SendTick > 0 {
		// 自动发送数据包
		go func() {
			ticker := time.NewTicker(c.opts.SendTick)
			defer ticker.Stop()

			for range ticker.C {
				c.sendTick <- 1
			}
		}()
	}
	go func() {
		if c.Connected() {
			// 和某个客户端建立的连接
			c.connectedRecvLoop()
		} else {
			// 监听者
			c.unconnectedRecvLoop()
		}
	}()
	
	c.sendLoop()
}

func (c *LtcpConn) connectedRecvLoop() {
	// data := make([]byte, MAX_PACKAGE)
	// for {
	// 	// 从连接中读取字节流
	// 	_, err := c.conn.Read(data)
	// 	if err != nil {
	// 		// TODO: 错误处理
	// 		c.recvErr <- err
	// 		return
	// 	}
	// }
}

func (c *LtcpConn) unconnectedRecvLoop() {

}

func (c *LtcpConn) sendLoop() {
	for {
		select {
		case tick := <-c.sendTick:
			c.handleSendTick(tick)
		}
	}
}

func (c *LtcpConn) handleSendTick(tick int) {
	// TODO: 1. 整理所有准备好的数据，封装成 packet
	if err := c.collectBufferData(); err != nil {
		c.sendErr <- err
		return
	}
	// TODO: 2. 发送所有的 packets
}

func (c *LtcpConn) collectBufferData() error {
	sendNum := 0
	for sendNum < c.opts.MaxSendNumPerTick {
		select {
		case bts := <- c.sendChan:
			// TODO: 整理数据包
			fmt.Println(string(bts))
			sendNum++
		default:
			return nil
		}
	}
	return nil
}