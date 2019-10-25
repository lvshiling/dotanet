package utils

import (
	"dq/vec2d"
)

var width = 12.0
var height = 12.0

//标准矩形
type Rect struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

//检查两个矩形是否相交
func CheckRectCollision(r1 Rect, r2 Rect) bool {
	if r1.MaxX < r2.MinX || r2.MaxX < r1.MinX || r1.MaxY < r2.MinY || r2.MaxY < r1.MinX {
		return false
	}
	return true
}

func CreateRectFromCYWardR(center vec2d.Vec2, r vec2d.Vec2) Rect {
	p1 := vec2d.Vec2{X: center.X - r.X, Y: center.Y - r.Y}
	p2 := vec2d.Vec2{X: center.X + r.X, Y: center.Y + r.Y}

	return CreateRectFrom2(p1, p2)
}
func CreateRectFromCYWardP(center vec2d.Vec2, p []vec2d.Vec2) Rect {

	re := Rect{}
	for _, v := range p {
		v.X += center.X
		v.Y += center.Y
		if re.MinX > v.X {
			re.MinX = v.X
		}
		if re.MinY > v.Y {
			re.MinY = v.Y
		}
		if re.MaxX < v.X {
			re.MaxX = v.X
		}
		if re.MaxY < v.Y {
			re.MaxY = v.Y
		}
	}
	return re
}

func CreateRectFrom2(p1 vec2d.Vec2, p2 vec2d.Vec2) Rect {
	re := Rect{}

	re.MinX = p1.X
	if re.MinX > p2.X {
		re.MinX = p2.X
	}
	re.MinY = p1.Y
	if re.MinY > p2.Y {
		re.MinY = p2.Y
	}
	re.MaxX = p1.X
	if re.MaxX < p2.X {
		re.MaxX = p2.X
	}
	re.MaxY = p1.Y
	if re.MaxY < p2.Y {
		re.MaxY = p2.Y
	}

	return re
}
func CreateRect(p []vec2d.Vec2) Rect {
	re := Rect{}
	for _, v := range p {
		if re.MinX > v.X {
			re.MinX = v.X
		}
		if re.MinY > v.Y {
			re.MinY = v.Y
		}
		if re.MaxX < v.X {
			re.MaxX = v.X
		}
		if re.MaxY < v.Y {
			re.MaxY = v.Y
		}
	}
	return re
}

type SceneZone struct {
	ZoneX int32
	ZoneY int32
}

//通过坐标位置获取分区 来自宽高
func GetSceneZoneFromWH(x float64, y float64, w float64, h float64) SceneZone {
	re := SceneZone{}
	re.ZoneX = int32(x / w)
	re.ZoneY = int32(y / h)

	return re
}

//通过坐标位置获取 可视分区(自己分区的周围1圈分区)
func GetVisibleZonesFromWH(x float64, y float64, w float64, h float64) []SceneZone {
	my := SceneZone{}
	my.ZoneX = int32(x / w)
	my.ZoneY = int32(y / h)

	re := make([]SceneZone, 9)
	//从左下到右上
	re[0] = SceneZone{ZoneX: my.ZoneX - 1, ZoneY: my.ZoneY - 1}
	re[1] = SceneZone{ZoneX: my.ZoneX, ZoneY: my.ZoneY - 1}
	re[2] = SceneZone{ZoneX: my.ZoneX + 1, ZoneY: my.ZoneY - 1}

	re[3] = SceneZone{ZoneX: my.ZoneX - 1, ZoneY: my.ZoneY}
	re[4] = SceneZone{ZoneX: my.ZoneX, ZoneY: my.ZoneY}
	re[5] = SceneZone{ZoneX: my.ZoneX + 1, ZoneY: my.ZoneY}

	re[6] = SceneZone{ZoneX: my.ZoneX - 1, ZoneY: my.ZoneY + 1}
	re[7] = SceneZone{ZoneX: my.ZoneX, ZoneY: my.ZoneY + 1}
	re[8] = SceneZone{ZoneX: my.ZoneX + 1, ZoneY: my.ZoneY + 1}

	return re
}

//通过2点坐标位置获取 可视分区(自己分区的周围1圈分区)
func GetVisibleZonesFromWH_Two(x float64, y float64, x2 float64, y2 float64, w float64, h float64) []SceneZone {

	//获取最小点和最大点
	minX := x
	if minX > x2 {
		minX = x2
	}
	minY := y
	if minY > y2 {
		minY = y2
	}
	maxX := x
	if maxX < x2 {
		maxX = x2
	}
	maxY := y
	if maxY < y2 {
		maxY = y2
	}
	//-------------
	leftbottomZone := SceneZone{}
	leftbottomZone.ZoneX = int32(minX / w)
	leftbottomZone.ZoneY = int32(minY / h)
	righttopZone := SceneZone{}
	righttopZone.ZoneX = int32(maxX / w)
	righttopZone.ZoneY = int32(maxY / h)

	re := make([]SceneZone, 0)
	for liney := leftbottomZone.ZoneY - 1; liney <= righttopZone.ZoneY+1; liney++ {
		for linex := leftbottomZone.ZoneX - 1; linex <= righttopZone.ZoneX+1; linex++ {
			onezone := SceneZone{}
			onezone.ZoneX = linex
			onezone.ZoneY = liney
			re = append(re, onezone)
		}
	}

	return re
}

//通过坐标位置获取分区
func GetSceneZone(x float64, y float64) SceneZone {
	re := SceneZone{}
	re.ZoneX = int32(x / width)
	re.ZoneY = int32(y / height)

	return re
}

//通过坐标位置获取 可视分区(自己分区的周围1圈分区)
func GetVisibleZones(x float64, y float64) []SceneZone {
	my := SceneZone{}
	my.ZoneX = int32(x / width)
	my.ZoneY = int32(y / height)

	re := make([]SceneZone, 9)
	//从左下到右上
	re[0] = SceneZone{ZoneX: my.ZoneX - 1, ZoneY: my.ZoneY - 1}
	re[1] = SceneZone{ZoneX: my.ZoneX, ZoneY: my.ZoneY - 1}
	re[2] = SceneZone{ZoneX: my.ZoneX + 1, ZoneY: my.ZoneY - 1}

	re[3] = SceneZone{ZoneX: my.ZoneX - 1, ZoneY: my.ZoneY}
	re[4] = SceneZone{ZoneX: my.ZoneX, ZoneY: my.ZoneY}
	re[5] = SceneZone{ZoneX: my.ZoneX + 1, ZoneY: my.ZoneY}

	re[6] = SceneZone{ZoneX: my.ZoneX - 1, ZoneY: my.ZoneY + 1}
	re[7] = SceneZone{ZoneX: my.ZoneX, ZoneY: my.ZoneY + 1}
	re[8] = SceneZone{ZoneX: my.ZoneX + 1, ZoneY: my.ZoneY + 1}

	return re
}
