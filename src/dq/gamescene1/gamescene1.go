package gamescene1

import (
	"fmt"
)
import (
	"dq/network"
)

type GameScene1 struct {

	// tcp
	TCPAddr string

	TcpClient *network.TCPClient
}

func (game5g *GameScene1) Run(closeSig chan bool) {

	var tcpClient *network.TCPClient
	if game5g.TCPAddr != "" {
		tcpClient = new(network.TCPClient)
		tcpClient.Addr = game5g.TCPAddr

		tcpClient.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &GameScene1Agent{conn: conn}
			a.RegisterToGate()
			return a
		}
	}

	if tcpClient != nil {
		game5g.TcpClient = tcpClient
		tcpClient.Start()
	}
	<-closeSig

	if tcpClient != nil {
		tcpClient.Close()
		game5g.TcpClient = nil
	}
	fmt.Println("GameScene1 over")
}
