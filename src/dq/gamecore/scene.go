package gamecore

import (
	"dq/conf"
	"dq/cyward"
	"dq/log"
	//"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"time"
)

type Scene struct {
	FirstUpdateTime int64            //上次更新时间
	MoveCore        *cyward.WardCore //移动核心
	SceneName       string           //场景名字
	CurFrame        int32            //当前帧

	Players   map[int32]*Player     //游戏中所有的玩家
	Units     map[int32]*Unit       //游戏中所有的单位
	ZoneUnits map[SceneZone][]*Unit //区域中的单位

	NextAddUnit    *utils.BeeMap //下一帧需要增加的单位
	NextRemoveUnit *utils.BeeMap //下一帧需要删除的单位

	NextAddPlayer    *utils.BeeMap //下一帧需要增加的玩家
	NextRemovePlayer *utils.BeeMap //下一帧需要删除的玩家

	Quit       bool //是否退出
	SceneFrame int32
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
	this.SceneFrame = 20
	this.CurFrame = 0

	this.FirstUpdateTime = time.Now().UnixNano()

	this.NextAddUnit = utils.NewBeeMap()
	this.NextRemoveUnit = utils.NewBeeMap()

	this.NextAddPlayer = utils.NewBeeMap()
	this.NextRemovePlayer = utils.NewBeeMap()

	this.Players = make(map[int32]*Player)
	this.Units = make(map[int32]*Unit)
	this.ZoneUnits = make(map[SceneZone][]*Unit)

	scenedata := conf.GetSceneData(this.SceneName)

	//场景碰撞区域
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
	//场景分区数据 创建100个单位
	for i := 0; i < 20; i++ {
		for j := 0; j < 20; j++ {
			unit := CreateUnit(this)
			unit.ModeType = "Hero/hero1"
			//设置移动核心body
			pos := vec2d.Vec2{float64(-63 + j*6), float64(-63 + i*6)}
			r := vec2d.Vec2{0.3, 0.3}
			unit.Body = this.MoveCore.CreateBody(pos, r, 10)
			this.Units[unit.ID] = unit
		}

	}

}

func (this *Scene) Update() {

	log.Info("Update start")
	t1 := time.Now().UnixNano()
	log.Info("t1:%d", (t1)/1e6)
	for {
		//log.Info("Update loop")
		//t1 := time.Now().UnixNano()
		//log.Info("main time:%d", (t1)/1e6)

		this.DoAddAndRemoveUnit()

		this.DoLogic()

		this.DoMove()

		this.DoZone()

		this.DoSendData()

		this.CurFrame++

		this.DoSleep()

		//处理分区

		if this.Quit {
			break
		}

	}

	t2 := time.Now().UnixNano()
	log.Info("t2:%d   delta:%d    frame:%d", (t2)/1e6, (t2-t1)/1e6, this.CurFrame)

}

//同步数据
func (this *Scene) DoSendData() {

	//生成单位的 客户端 显示数据
	for _, v := range this.Units {
		v.FreshClientDataSub()
		v.FreshClientData()
	}

	//遍历所有玩家
	for _, player := range this.Players {
		v := player.MainUnit
		if v == nil {
			continue
		}
		zones := GetVisibleZones((v.Body.Position.X), (v.Body.Position.Y))
		//遍历可视区域
		for _, vzone := range zones {
			if _, ok := this.ZoneUnits[vzone]; ok {
				//遍历区域中的单位
				for _, unit := range this.ZoneUnits[vzone] {
					player.AddUnitData(unit)
				}
			}
		}
		player.SendUpdateMsg(this.CurFrame)
	}
}

//处理移动
func (this *Scene) DoMove() {
	this.MoveCore.Update(1 / float64(this.SceneFrame))
}

//处理分区
func (this *Scene) DoZone() {
	this.ZoneUnits = make(map[SceneZone][]*Unit)
	for _, v := range this.Units {

		zone := GetSceneZone((v.Body.Position.X), (v.Body.Position.Y))
		this.ZoneUnits[zone] = append(this.ZoneUnits[zone], v)

	}
}

//处理单位逻辑
func (this *Scene) DoLogic() {
	//处理单位逻辑
	for _, v := range this.Units {
		v.PreUpdate(1 / float64(this.SceneFrame))
	}
	for _, v := range this.Units {
		v.Update(1 / float64(this.SceneFrame))
	}
}

//处理sleep
func (this *Scene) DoSleep() {
	sencond := time.Second
	onetime := int64(1 / float64(this.SceneFrame) * float64(sencond))
	t1 := time.Now().UnixNano()

	nexttime := this.FirstUpdateTime + onetime*int64(this.CurFrame)

	sleeptime := nexttime - t1

	//log.Info("main time:%d    %d", (t1-this.LastUpdateTime)/1e6, onetime/1e6)
	//log.Info("sleep :%d   %d    %d", (t1)/1e6, this.LastUpdateTime/1e6, onetime/1e6)
	if sleeptime > 0 {
		time.Sleep(time.Duration(sleeptime))
	} else {

	}

}

//增加单位
func (this *Scene) DoAddAndRemoveUnit() {
	//增加
	itemsadd := this.NextAddUnit.Items()
	for k, v := range itemsadd {
		if v.(*Unit).Body != nil {
			this.MoveCore.RemoveBody(v.(*Unit).Body)
			v.(*Unit).Body = nil
		}

		//设置移动核心body
		pos := vec2d.Vec2{0, 0}
		r := vec2d.Vec2{0.3, 0.3}

		v.(*Unit).Body = this.MoveCore.CreateBody(pos, r, 0)

		this.Units[k.(int32)] = v.(*Unit)

		this.NextAddUnit.Delete(k)

		//this.Players[]
	}
	//this.NextAddUnit.DeleteAll()

	//删除
	itemsremove := this.NextRemoveUnit.Items()
	for k, v := range itemsremove {
		this.MoveCore.RemoveBody(v.(*Unit).Body)
		//v.(*Unit).Body
		//this.Units.Delete(k)
		delete(this.Units, k.(int32))

		this.NextRemoveUnit.Delete(k)
	}
	//this.NextRemoveUnit.DeleteAll()

	//增加玩家
	playeradd := this.NextAddPlayer.Items()
	for k, v := range playeradd {
		this.Players[k.(int32)] = v.(*Player)

		this.NextAddPlayer.Delete(k)
	}
	//this.NextAddPlayer.DeleteAll()

	//删除玩家
	playerremove := this.NextRemovePlayer.Items()
	for k, _ := range playerremove {
		delete(this.Players, k.(int32))

		this.NextRemovePlayer.Delete(k)
	}
	//this.NextRemovePlayer.DeleteAll()

}

func (this *Scene) Close() {
	this.Quit = true
}

//玩家进入
func (this *Scene) PlayerGoin(player *Player, datas []byte) {
	if player.MainUnit == nil {
		player.MainUnit = CreateUnitByPlayer(this, player, datas)

	}

	this.NextAddUnit.Set(player.MainUnit.ID, player.MainUnit)

	this.NextAddPlayer.Set(player.Uid, player)

}

//玩家退出
func (this *Scene) PlayerGoout(player *Player) {
	//删除主单位
	this.NextRemoveUnit.Set(player.MainUnit.ID, player.MainUnit)

	this.NextRemovePlayer.Set(player.Uid, player)
}
