package gamecore

var width = 16.0
var height = 16.0

type SceneZone struct {
	ZoneX int32
	ZoneY int32
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
