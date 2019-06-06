// testclient project testclient.go

package main

import (
	"dq/app"
	"dq/log"
	//"dq/vec2d"
)

func main() {

	//	count := 20.0
	//	for i := 1.0; i <= count; i++ {
	//		dir := vec2d.Vec2{0, 1}
	//		dir.Rotate(float64(i * (180.0 / count)))
	//		log.Info("111=====%f %f", dir.X, dir.Y)

	//		dir2 := vec2d.Vec2{0, 1}
	//		dir2.Rotate(0.0 - float64(i*(180.0/count)))
	//		log.Info("222=====%f %f", dir2.X, dir2.Y)
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
