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

	ltcp *Ltcp // 负责可靠传输
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
	AutoSend:          true,
	SendTick:          time.Millisecond * 20,
	MaxSendNumPerTick: 16,
}

func NewConn(conn *net.UDPConn, opts LtcpConnOptions) *LtcpConn {
	con := &LtcpConn{
		conn:     conn,
		recvChan: make(chan []byte, 1<<16),
		recvErr:  make(chan error, 2),
		sendChan: make(chan []byte, 1<<16),
		sendErr:  make(chan error, 2),
		sendTick: make(chan int, 2),
		in:       make(chan []byte, 1<<16),
		ltcp:     NewLtcp(),
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

// Read 只负责从 recvChan 中拿数据
func (c *LtcpConn) Read(bts []byte) (n int, err error) {
	select {
	case data := <-c.recvChan:
		copy(bts, data)
		return len(data), nil
	case err := <-c.recvErr:
		return 0, err
	}
}

// 从 ltcp 中的读取缓冲区读取数据
// 把 Packet 中的 Payload 读出来，发送到 recvChan
func (c *LtcpConn) ltcpRecv(data []byte) error {
	for {
		n, err := c.ltcp.Recv(data)
		fmt.Println(n)
		if err != nil {
			c.recvErr <- err
			return err
		} else if n == 0 {
			break
		}
		bts := make([]byte, n)
		copy(bts, data[:n])
		c.recvChan <- bts
	}
	return nil
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
		c.unconnectedRecvLoop()
		// if c.Connected() {
		// 	// 和某个客户端建立的连接
		// 	c.connectedRecvLoop()
		// } else {
		// 	// 监听者
		// 	c.unconnectedRecvLoop()
		// }
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
	// TODO: 需要定义出去
	// 这部分逻辑有问题，ltcp 只负责监听 c.in
	// 读取应该负责从接受缓冲区读取
	data := make([]byte, 0x7fff)
	for {
		select {
		case bts := <-c.in:
			if err := c.ltcp.Parse(bts); err != nil {
				fmt.Println(err)
				return
			}
			if err := c.ltcpRecv(data); err != nil {
				return
			}
		}
	}
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
	// TODO: 1. 整理所有准备好的数据
	if err := c.collectBufferData(); err != nil {
		c.sendErr <- err
		return
	}
	// TODO: 2. 封装所有的数据为 Packet
	c.ltcp.Package(tick)
	// TODO: 3. 通过网络发送所有 Packet，封装一个迭代器
	for p := c.ltcp.sendQueue.Head.Next; p != c.ltcp.sendQueue.Tail; p = p.Next {
		data, err := p.Packet.Serialize()
		if err != nil {
			// TODO: 处理错误
			fmt.Println(err)
		}
		if c.Connected() {
			c.conn.Write(data)
		} else {
			c.conn.WriteToUDP(data, c.remoteAddr)
		}
	}
}

func (c *LtcpConn) collectBufferData() error {
	sendNum := 0
	for sendNum < c.opts.MaxSendNumPerTick {
		select {
		case bts := <-c.sendChan:
			// 整理数据包
			c.ltcp.Collect(bts)
			sendNum++
		default:
			return nil
		}
	}
	return nil
}
