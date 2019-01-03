package network

import (
	//"encoding/binary"
	//"errors"
	//"io"
	"math"
	//"dq/log"
)

// --------------
// | len | data |
// --------------
type KcpMsgParser struct {
	lenMsgLen    int
	minMsgLen    uint32
	maxMsgLen    uint32
	littleEndian bool
}

func NewKcpMsgParser() *KcpMsgParser {
	p := new(KcpMsgParser)
	p.lenMsgLen = 4
	p.minMsgLen = 1
	p.maxMsgLen = 4096 * 20
	p.littleEndian = false

	return p
}

// It's dangerous to call the method on reading or writing
func (p *KcpMsgParser) SetMsgLen(lenMsgLen int, minMsgLen uint32, maxMsgLen uint32) {
	if lenMsgLen == 1 || lenMsgLen == 2 || lenMsgLen == 4 {
		p.lenMsgLen = lenMsgLen
	}
	if minMsgLen != 0 {
		p.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		p.maxMsgLen = maxMsgLen
	}

	var max uint32
	switch p.lenMsgLen {
	case 1:
		max = math.MaxUint8
	case 2:
		max = math.MaxUint16
	case 4:
		max = math.MaxUint32
	}
	if p.minMsgLen > max {
		p.minMsgLen = max
	}
	if p.maxMsgLen > max {
		p.maxMsgLen = max
	}
}

// It's dangerous to call the method on reading or writing
func (p *KcpMsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

// goroutine safe
func (p *KcpMsgParser) Read(conn *KCPConn) ([]byte, error) {

	//	var b [4]byte
	//	bufMsgLen := b[:p.lenMsgLen]

	//	// read len
	//	if _, err := conn.ReadFull(bufMsgLen); err != nil {
	//		return nil, err
	//	}
	//	//conn.ReadFull(bufMsgLen)

	//	// parse len
	//	var msgLen uint32
	//	switch p.lenMsgLen {
	//	case 1:
	//		msgLen = uint32(bufMsgLen[0])
	//	case 2:
	//		if p.littleEndian {
	//			msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
	//		} else {
	//			msgLen = uint32(binary.BigEndian.Uint16(bufMsgLen))
	//		}
	//	case 4:
	//		if p.littleEndian {
	//			msgLen = binary.LittleEndian.Uint32(bufMsgLen)
	//		} else {
	//			msgLen = binary.BigEndian.Uint32(bufMsgLen)
	//		}
	//	}

	//	if msgLen > p.maxMsgLen {
	//		return nil, errors.New("message too long")
	//	} else if msgLen < p.minMsgLen {
	//		return nil, errors.New("message too short")
	//	}

	//	msgData := make([]byte, msgLen)
	//	if n, err := conn.ReadFull(msgData); err != nil {
	//		return nil, err
	//	}
	msgData := make([]byte, p.maxMsgLen)
	n, err := conn.Read(msgData)
	if err != nil {
		return nil, err
	}
	msgData = msgData[:n]

	return msgData, nil
}

// goroutine safe
func (p *KcpMsgParser) Write(conn *KCPConn, args ...[]byte) error {
	// get len
	//	var msgLen uint32
	//	for i := 0; i < len(args); i++ {
	//		msgLen += uint32(len(args[i]))
	//	}

	//	// check len
	//	if msgLen > p.maxMsgLen {
	//		return errors.New("message too long")
	//	} else if msgLen < p.minMsgLen {
	//		return errors.New("message too short")
	//	}

	//	msg := make([]byte, uint32(p.lenMsgLen)+msgLen)

	//	// write len
	//	switch p.lenMsgLen {
	//	case 1:
	//		msg[0] = byte(msgLen)
	//	case 2:
	//		if p.littleEndian {
	//			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
	//		} else {
	//			binary.BigEndian.PutUint16(msg, uint16(msgLen))
	//		}
	//	case 4:
	//		if p.littleEndian {
	//			binary.LittleEndian.PutUint32(msg, msgLen)
	//		} else {
	//			binary.BigEndian.PutUint32(msg, msgLen)
	//		}
	//	}

	//	// write data
	//	l := p.lenMsgLen
	//	for i := 0; i < len(args); i++ {
	//		copy(msg[l:], args[i])
	//		l += len(args[i])
	//	}
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}
	msg := make([]byte, msgLen)
	l := 0
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	conn.Write(msg)

	return nil
}
