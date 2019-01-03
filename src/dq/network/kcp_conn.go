package network

import (
	"dq/conf"
	"dq/kcp"
	"dq/log"
	"net"
	"sync"
	"time"
)

type KcpConnSet map[*kcp.UDPSession]struct{}

type KCPConn struct {
	sync.Mutex
	conn      *kcp.UDPSession
	writeChan chan []byte
	closeFlag bool
	msgParser *KcpMsgParser
}

func newKCPConn(conn *kcp.UDPSession, pendingWriteNum int, msgParser *KcpMsgParser) *KCPConn {

	conn.SetWindowSize(128, 128)
	conn.SetMinRto(10)

	tcpConn := new(KCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser

	go func() {
		for b := range tcpConn.writeChan {
			if b == nil {
				break
			}

			_, err := conn.Write(b)
			if err != nil {
				break
			}
		}

		conn.Close()
		tcpConn.Lock()
		tcpConn.closeFlag = true
		tcpConn.Unlock()
	}()

	return tcpConn
}

func (tcpConn *KCPConn) doDestroy() {
	//tcpConn.conn.(*kcp.UDPSession).SetLinger(0)
	tcpConn.conn.Close()

	if !tcpConn.closeFlag {
		close(tcpConn.writeChan)
		tcpConn.closeFlag = true
	}
}

func (tcpConn *KCPConn) Destroy() {
	tcpConn.Lock()
	defer tcpConn.Unlock()

	tcpConn.doDestroy()
}

func (tcpConn *KCPConn) Close() {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return
	}

	tcpConn.doWrite(nil)
	tcpConn.closeFlag = true
}

func (tcpConn *KCPConn) doWrite(b []byte) {
	for len(tcpConn.writeChan) >= cap(tcpConn.writeChan) {
		log.Debug("conn: channel full %d  %d", len(tcpConn.writeChan), cap(tcpConn.writeChan))
		time.Sleep(time.Millisecond * 2)
	}

	tcpConn.writeChan <- b
}

// b must not be modified by the others goroutines
func (tcpConn *KCPConn) Write(b []byte) {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag || b == nil {
		return
	}

	tcpConn.doWrite(b)
}

//func (tcpConn *KCPConn) Read(b []byte) (int, error) {
//	return tcpConn.conn.Read(b)
//}
func (tcpConn *KCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

func (tcpConn *KCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

func (tcpConn *KCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

func (tcpConn *KCPConn) ReadMsg() ([]byte, error) {
	return tcpConn.msgParser.Read(tcpConn)
}

func (tcpConn *KCPConn) WriteMsg(args []byte) error {
	return tcpConn.msgParser.Write(tcpConn, args)
}
func (tcpConn *KCPConn) ReadSucc() {
	tcpConn.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(conf.Conf.GateInfo.TimeOut)))
}
