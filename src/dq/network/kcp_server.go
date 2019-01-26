package network

import (
	"dq/log"
	"net"
	//"sync"
	"time"
	//"fmt"
	"dq/kcp"
	"dq/utils"
)

//type ServerData struct {
//	Addr            string
//	MaxConnNum      int
//	PendingWriteNum int
//	NewAgent        func(Conn) Agent
//	//ln              net.Listener

//	//Agents			map[int]interface{}
//	Agents         *utils.BeeMap
//	LoginedConnect *utils.BeeMap
//	mutexConns     sync.Mutex
//	wgLn           sync.WaitGroup
//	wgConns        sync.WaitGroup

//}

type KCPServer struct {
	ServerData
	ln    *kcp.Listener
	conns KcpConnSet
	// msg parser
	msgParser *KcpMsgParser
}

func (server *KCPServer) GetLoginedConnect() *utils.BeeMap {
	return server.LoginedConnect
}
func (server *KCPServer) GetAgents() *utils.BeeMap {

	return server.Agents
}

func (server *KCPServer) Start() {
	server.init()
	go server.run()
}

func (server *KCPServer) init() {
	//ln, err := net.Listen("tcp", server.Addr)
	ln, err := kcp.Listen(kcp.MODE_FAST, server.Addr)
	if err != nil {
		log.Error("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 10000
		log.Debug("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Debug("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.NewAgent == nil {
		log.Error("NewAgent must not be nil")
	}
	server.ln = ln
	server.conns = make(KcpConnSet)
	server.Agents = utils.NewBeeMap()
	server.LoginedConnect = utils.NewBeeMap()

	// msg parser
	msgParser := NewKcpMsgParser()
	server.msgParser = msgParser

	log.Info("------Listen:" + server.Addr)
}

func (server *KCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Debug("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Debug("too many connections")
			continue
		}
		server.conns[conn] = struct{}{}

		server.mutexConns.Unlock()

		server.wgConns.Add(1)

		tcpConn := newKCPConn(conn, server.PendingWriteNum, server.msgParser)
		agent := server.NewAgent(tcpConn)

		server.Agents.Set(agent.GetConnectId(), agent)

		go func() {
			agent.Run()

			// cleanup
			tcpConn.Close()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()
			server.Agents.Delete(agent.GetConnectId())
			agent.OnClose()

			server.wgConns.Done()

			log.Debug("conn close :%d ---conn size:%d", agent.GetConnectId(), server.Agents.Size())
		}()
	}
}

func (server *KCPServer) Close() {
	server.ln.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()

	server.wgConns.Wait()
	log.Info("tcp Close :%s", server.Addr)
}
