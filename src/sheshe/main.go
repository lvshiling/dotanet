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

//type test struct {
//	x int32
//	y int32
//}

func main() {

	//	t1 := make(map[int32][]test)
	//	for i := 0; i < 10; i++ {

	//		p := test{}
	//		p.x = int32(i)
	//		p.y = int32(i)
	//		t1[int32(i%2)] = append(t1[int32(i%2)], p)
	//	}

	//	t1 = make(map[int32][]test)
	//	_, ok1 := t1[0]
	//	t1[100] = make([]test, 0)
	//	_, ok2 := t1[100]

	//	log.Info("-:%b---%b", ok1, ok2)

	//	for k, v := range t1 {
	//		for _, v1 := range v {
	//			log.Info("k:%d---%v", k, v1)
	//		}

	//	}

	app := new(app.DefaultApp)
	app.Run()
	log.Info("dq over!")
	//	log.Info("111!")
	//	core := cyward.CreateWardCore()
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
	//	var points []vec2d.Vec2
	//	points = append(points, vec2d.Vec2{-10, 0})
	//	points = append(points, vec2d.Vec2{0, 10})
	//	points = append(points, vec2d.Vec2{10, 0})
	//	points = append(points, vec2d.Vec2{0, -10})
	//	//	points[0] = vec2d.Vec2{-10, 0}
	//	//	points[1] = vec2d.Vec2{0, 10}
	//	//	points[2] = vec2d.Vec2{10, 0}
	//	//	points[3] = vec2d.Vec2{0, -10}
	//	core.CreateBodyPolygon(vec2d.Vec2{400, 200}, points, 30.0)

	//	t1 := time.Now().UnixNano()

	//	test[0].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[1].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[2].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[3].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[4].SetTarget(vec2d.Vec2{float64(500), float64(300)})
	//	test[5].SetTarget(vec2d.Vec2{float64(500), float64(300)})

	//	t2 := time.Now().UnixNano()
	//	log.Info("main time:%d", (t2-t1)/1e6)

}
