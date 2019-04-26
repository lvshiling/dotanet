package utils

var width = 16.0
var height = 16.0

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
