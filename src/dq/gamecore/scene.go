package gamecore

import (
	"dq/conf"
	"dq/cyward"
	"dq/log"
	"dq/utils"
	"dq/vec2d"
	"time"
)

var SceneFrame = 20

type Scene struct {
	LastUpdateTime int64            //上次更新时间
	MoveCore       *cyward.WardCore //移动核心
	SceneName      string           //场景名字

	Units *utils.BeeMap //游戏中的单位

	NextAddUnit    *utils.BeeMap //下一帧需要增加的单位
	NextRemoveUnit *utils.BeeMap //下一帧需要删除的单位

	Quit bool //是否退出
}

func CreateScene(name string) *Scene {
	scene := &Scene{}
	scene.SceneName = name
	scene.Quit = false
	scene.Init()
	return scene
}

//初始化
func (this *Scene) Init() {

	this.LastUpdateTime = time.Now().UnixNano()

	this.Units = utils.NewBeeMap()
	this.NextAddUnit = utils.NewBeeMap()
	this.NextRemoveUnit = utils.NewBeeMap()

	scenedata := conf.GetSceneData(this.SceneName)

	this.MoveCore = cyward.CreateWardCore()
	for _, v := range scenedata.Collides {
		log.Info("Collide %v", v)
		if v.IsRect == true {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			r := vec2d.Vec2{v.Width, v.Height}
			this.MoveCore.CreateBody(pos, r, 1)
		} else {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			this.MoveCore.CreateBodyPolygon(pos, v.Points, 1)
		}
	}
}

func (this *Scene) Update() {

	for {
		this.DoAddAndRemoveUnit()

		this.MoveCore.Update(1 / float64(SceneFrame))

		//处理单位逻辑
		units := this.Units.Items()
		for _, v := range units {
			v.(*Unit).PreUpdate(1 / float64(SceneFrame))
		}
		for _, v := range units {
			v.(*Unit).Update(1 / float64(SceneFrame))
		}
		sencond := time.Second
		onetime := int64(1 / float64(SceneFrame) * float64(sencond))
		t1 := time.Now().UnixNano()
		if (t1 - this.LastUpdateTime) < onetime {
			time.Sleep(time.Duration(onetime - (t1 - this.LastUpdateTime)))
		}
		this.LastUpdateTime = t1

		if this.Quit {
			break
		}

	}

}

//增加单位
func (this *Scene) DoAddAndRemoveUnit() {
	//增加
	itemsadd := this.NextAddUnit.Items()
	for k, v := range itemsadd {
		if v.(*Unit).Body == nil {

			//设置移动核心body
			pos := vec2d.Vec2{0.0, 0.0}
			r := vec2d.Vec2{1, 1}

			v.(*Unit).Body = this.MoveCore.CreateBody(pos, r, 0)
		}
		this.Units.Set(k, v)
	}
	this.NextAddUnit.DeleteAll()

	//删除
	itemsremove := this.NextRemoveUnit.Items()
	for k, _ := range itemsremove {
		this.Units.Delete(k)
	}
	this.NextRemoveUnit.DeleteAll()

}

func (this *Scene) Close() {
	this.Quit = true
}

//玩家进入
func (this *Scene) PlayerGoin(player *Player) {
	if player.MainUnit == nil {
		player.MainUnit = CreateUnit(this)

		this.NextAddUnit.Set(player.MainUnit.ID, player.MainUnit)
	}
}

//玩家退出
func (this *Scene) PlayerGoout(player *Player) {

}
