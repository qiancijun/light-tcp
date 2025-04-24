package ltcp

import (
	"fmt"
	"net"
	"sync"
)

const (
	MAX_PACKAGE = 0x7fff
)

// 负责对连接的监听与管理
type LtcpListener struct {
	conn *net.UDPConn
	lock sync.RWMutex

	newLtcpConnChan chan *LtcpConn
	newLtcpErr      chan error

	ltcpConnMap map[string]*LtcpConn
}

func NewLtcpListener(conn *net.UDPConn) *LtcpListener {
	// TODO: 提供 option 选项
	listen := &LtcpListener{
		conn:            conn,
		newLtcpConnChan: make(chan *LtcpConn, 1024),
		newLtcpErr:      make(chan error, 12),
		ltcpConnMap:     make(map[string]*LtcpConn),
	}
	return listen
}

func (l *LtcpListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.newLtcpConnChan:
		return c, nil
	case e := <-l.newLtcpErr:
		return nil, e
	}
}

func (l *LtcpListener) Close() error {
	// 先断开所有维护的连接
	l.closeAllUdpConn()
	l.ltcpConnMap = make(map[string]*LtcpConn)
	return l.conn.Close()
}

// 关闭回调函数，交给具体的虚拟链接调用，用于清楚链接信息的记录
func (l *LtcpListener) CloseLtcp(addr string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.ltcpConnMap, addr)
}

func (l *LtcpListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}

func (l *LtcpListener) closeAllUdpConn() {
	l.lock.Lock()
	defer l.lock.Unlock()
	for name, conn := range l.ltcpConnMap {
		if err := conn.Close(); err != nil {
			fmt.Printf("close conn %s error: %s\n", name, err)
		}
	}
}

func (l *LtcpListener) Run() {
	data := make([]byte, MAX_PACKAGE)
	for {
		n, remoteAddr, err := l.conn.ReadFromUDP(data)
		if err != nil {
			// TODO: 处理错误
			l.newLtcpErr <- err
			continue
		}
		l.lock.RLock()
		ltcpConn, ok := l.ltcpConnMap[remoteAddr.String()]
		l.lock.RUnlock()

		// 没有这个链接的管理，创建一个虚拟的链接
		if !ok {
			ltcpConn = NewUnConn(l.conn, remoteAddr, l.CloseLtcp, DefaultLtcpConnOptions)
			l.lock.Lock()
			l.ltcpConnMap[remoteAddr.String()] = ltcpConn
			l.lock.Unlock()
			l.newLtcpConnChan <- ltcpConn
		}
		bts := make([]byte, n)
		copy(bts, data[:n])
		ltcpConn.in <- bts
	}
}
