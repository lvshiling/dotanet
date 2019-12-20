// testclient project testclient.go

package main

import (
	//"encoding/json"
	"flag"
	//"net/url"
	"strconv"
	"sync"
	"time"

	"fmt"

	//"dq/datamsg"
	"dq/network"
	//"dq/protobuf"

	//"math/rand"

	//"github.com/gorilla/websocket"
)

//var addr = flag.String("addr", "www.game5868.top:443", "http service address")

var addr = flag.String("addr", "127.0.0.1:1117", "http service address")

func main() {

	fmt.Println("start!!")
	var waitg sync.WaitGroup
	for j := 0; j < 25; j++ {
		waitg.Add(1)

		go func() {

			client(strconv.Itoa(j))
			//fmt.Println("client count:" + strconv.Itoa(j))
			waitg.Done()
		}()

		time.Sleep(time.Millisecond * 200)
	}

	waitg.Wait()
	fmt.Println("over!!")

}

func client(id string) {

	closesig := make(chan bool, 1)

	var tcpClient *network.TCPClient
	tcpClient = new(network.TCPClient)
	tcpClient.Addr = "127.0.0.1:1119"
	tcpClient.NewAgent = func(conn *network.TCPConn) network.Agent {
		a := &GameAgent{conn: conn}

		a.QuickLogin(id, "name"+id)
		return a
	}
	if tcpClient != nil {
		tcpClient.Start()
	}
	<-closesig
}
