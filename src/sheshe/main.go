// testclient project testclient.go

package main

import (
	//"io/ioutil"
	//"net/http"
	"dq/app"
	"dq/log"
	//"os"
	//"os/signal"
	//"io"
	//	"fmt"
	//	"net"
	//"dq/rpc"
	//"time"
	//"net/rpc/jsonrpc"
	//"dq/cyward"
	//"dq/vec2d"
)

func main() {

	app := new(app.DefaultApp)
	app.Run()
	log.Info("dq over!")

	//	t1 := time.Now().UnixNano()

	//	addnum := 0
	//	for i := 0; i < 1000000000; i++ {
	//		addnum++
	//	}
	//	t2 := time.Now().UnixNano()
	//	log.Info("main time:%d", (t2-t1)/1e6)

	//	core := &cyward.WardCore{}
	//	var test []*cyward.Body

	//	for i := 0; i < 10; i++ {
	//		for j := 0; j < 10; j++ {
	//			pos := vec2d.Vec2{float64(100 + i*20), float64(100 + j*15)}
	//			r := vec2d.Vec2{float64(3 + i/3), float64(3 + j/3)}
	//			t := core.CreateBody(pos, r, 30.0)
	//			t.SetTag(i*10 + j)
	//			test = append(test, t)
	//		}

	//	}

	//	test[22].SetTarget(vec2d.Vec2{float64(50), float64(300)})

}
