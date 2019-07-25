package gamecore

import (
	"dq/conf"
	"dq/cyward"
	"dq/log"
	//"dq/protobuf"
	//"dq/timer"
	"dq/utils"
	"dq/vec2d"
	"time"
)

type Scene struct {
	FirstUpdateTime int64            //上次更新时间
	MoveCore        *cyward.WardCore //移动核心
	SceneName       string           //场景名字
	CurFrame        int32            //当前帧

	Players   map[int32]*Player           //游戏中所有的玩家
	Units     map[int32]*Unit             //游戏中所有的单位
	ZoneUnits map[utils.SceneZone][]*Unit //区域中的单位

	Bullets     map[int32]*Bullet             //游戏中所有的子弹
	ZoneBullets map[utils.SceneZone][]*Bullet //区域中的的子弹

	Halos          map[int32]*Halo             //游戏中所有的光环
	ZoneHalos      map[utils.SceneZone][]*Halo //区域中的的光环
	CanRemoveHalos map[int32]*Halo             //可以被删除的halo 比如击杀单位后 halo无效

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
	this.SceneFrame = 40
	this.CurFrame = 0

	this.FirstUpdateTime = time.Now().UnixNano()

	this.NextAddUnit = utils.NewBeeMap()
	this.NextRemoveUnit = utils.NewBeeMap()

	this.NextAddPlayer = utils.NewBeeMap()
	this.NextRemovePlayer = utils.NewBeeMap()

	this.Players = make(map[int32]*Player)
	this.Units = make(map[int32]*Unit)
	this.ZoneUnits = make(map[utils.SceneZone][]*Unit)

	this.Bullets = make(map[int32]*Bullet)
	this.ZoneBullets = make(map[utils.SceneZone][]*Bullet)

	this.Halos = make(map[int32]*Halo)
	this.ZoneHalos = make(map[utils.SceneZone][]*Halo)
	this.CanRemoveHalos = make(map[int32]*Halo)

	scenedata := conf.GetSceneData(this.SceneName)

	//场景碰撞区域
	this.MoveCore = cyward.CreateWardCore()
	for _, v := range scenedata.Collides {
		log.Info("Collide %v", v)
		if v.IsRect == true {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			r := vec2d.Vec2{v.Width, v.Height}
			this.MoveCore.CreateBody(pos, r, 1, 2)
		} else {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			this.MoveCore.CreateBodyPolygon(pos, v.Points, 1, 2)
		}
	}
	//	//场景分区数据 创建100个单位
	//	for i := 0; i < 20; i++ {
	//		for j := 0; j < 20; j++ {
	//			unit := CreateUnit(this, 2)
	//			unit.SetAI(NewNormalAI(unit))
	//			//设置移动核心body
	//			pos := vec2d.Vec2{float64(-63 + j*6), float64(-63 + i*6)}
	//			r := vec2d.Vec2{unit.CollisionR, unit.CollisionR}
	//			unit.Body = this.MoveCore.CreateBody(pos, r, 0, 1)
	//			//unit.Body.Tag = i*20 + j
	//			this.Units[unit.ID] = unit

	//			//timer.AddCallback(time.Second*15+time.Second*time.Duration(((i*20)+j)*2), this.UnitBlink, unit)
	//		}

	//	}
	//	for i := 0; i < 4; i++ {
	//		//创建英雄
	//		hero1 := CreateUnit(this, 5+int32(i))
	//		hero1.AttackMode = 3 //全体攻击模式
	//		hero1.SetAI(NewNormalAI(hero1))
	//		//设置移动核心body
	//		pos1 := vec2d.Vec2{float64(2 + i*2), float64(5)}
	//		r1 := vec2d.Vec2{hero1.CollisionR, hero1.CollisionR}
	//		hero1.Body = this.MoveCore.CreateBody(pos1, r1, 0, 1)
	//		this.Units[hero1.ID] = hero1
	//	}

	//	//创建英雄2
	//	hero2 := CreateUnit(this, 15)
	//	hero2.AttackMode = 1 //和平攻击模式
	//	hero2.SetAI(NewNormalAI(hero2))
	//	//hero2.AddSkill(52, 4)
	//	//设置移动核心body
	//	pos2 := vec2d.Vec2{float64(10), float64(5)}
	//	r2 := vec2d.Vec2{hero2.CollisionR, hero2.CollisionR}
	//	hero2.Body = this.MoveCore.CreateBody(pos2, r2, 0, 1)
	//	this.Units[hero2.ID] = hero2

}

func (this *Scene) UnitBlink(unit interface{}) {
	//	if unit != nil {
	//		targetpos := vec2d.Vec2{X: 5, Y: 5}
	//		unit.(*Unit).Body.BlinkToPos(targetpos)
	//	}
}

//通过ID查找单位
func (this *Scene) FindUnitByID(id int32) *Unit {
	return this.Units[id]
}

//获取可视范围内的所有单位
func (this *Scene) FindVisibleUnitsByPos(pos vec2d.Vec2) []*Unit {

	units := make([]*Unit, 0)

	zones := utils.GetVisibleZones((pos.X), (pos.Y))
	//遍历可视区域
	for _, vzone := range zones {
		if _, ok := this.ZoneUnits[vzone]; ok {
			//遍历区域中的单位
			for _, unit := range this.ZoneUnits[vzone] {
				units = append(units, unit)

			}
		}
	}

	return units
}

//获取可视范围内的所有单位
func (this *Scene) FindVisibleUnits(my *Unit) []*Unit {

	units := make([]*Unit, 0)

	zones := utils.GetVisibleZones((my.Body.Position.X), (my.Body.Position.Y))
	//遍历可视区域
	for _, vzone := range zones {
		if _, ok := this.ZoneUnits[vzone]; ok {
			//遍历区域中的单位
			for _, unit := range this.ZoneUnits[vzone] {
				units = append(units, unit)

			}
		}
	}

	return units
}

func (this *Scene) Update() {

	log.Info("Update start")
	t1 := time.Now().UnixNano()
	log.Info("t1:%d", (t1)/1e6)
	for {
		//log.Info("Update loop")
		//t1 := time.Now().UnixNano()
		//log.Info("main time:%d", (t1)/1e6)

		time1 := utils.GetCurTimeOfSecond()
		this.DoRemoveBullet()
		this.DoRemoveHalo()
		this.DoAddAndRemoveUnit()

		time2 := utils.GetCurTimeOfSecond()
		this.DoLogic()
		time3 := utils.GetCurTimeOfSecond()
		this.UpdateHalo(1 / float32(this.SceneFrame))
		time4 := utils.GetCurTimeOfSecond()
		this.UpdateBullet(1 / float32(this.SceneFrame))
		time5 := utils.GetCurTimeOfSecond()

		this.DoMove()
		time6 := utils.GetCurTimeOfSecond()
		this.DoZone()
		time7 := utils.GetCurTimeOfSecond()
		this.DoSendData()
		time8 := utils.GetCurTimeOfSecond()

		if time8-time1 >= 10.01 {
			log.Info("time:%f %f %f %f %f %f %f  ", time2-time1, time3-time2, time4-time3, time5-time4,
				time6-time5, time7-time6, time8-time7)
		}

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
	//生成子弹的 客户端 显示数据
	for _, v := range this.Bullets {
		v.FreshClientDataSub()
		v.FreshClientData()
	}

	//生成光环的 客户端 显示数据
	for _, v := range this.Halos {
		v.FreshClientDataSub()
		v.FreshClientData()
	}

	//遍历所有玩家
	for _, player := range this.Players {
		v := player.MainUnit
		if v == nil {
			continue
		}
		zones := utils.GetVisibleZones((v.Body.Position.X), (v.Body.Position.Y))
		//遍历可视区域
		for _, vzone := range zones {
			if _, ok := this.ZoneUnits[vzone]; ok {
				//遍历区域中的单位
				for _, unit := range this.ZoneUnits[vzone] {
					player.AddUnitData(unit)
				}
			}
			if _, ok := this.ZoneBullets[vzone]; ok {
				//遍历区域中的单位
				for _, bullet := range this.ZoneBullets[vzone] {
					player.AddBulletData(bullet)
				}
			}
			//
			if _, ok := this.ZoneHalos[vzone]; ok {
				//遍历区域中的单位
				for _, halo := range this.ZoneHalos[vzone] {
					player.AddHaloData(halo)
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
	//单位分区
	this.ZoneUnits = make(map[utils.SceneZone][]*Unit)
	for _, v := range this.Units {

		zone := utils.GetSceneZone((v.Body.Position.X), (v.Body.Position.Y))
		this.ZoneUnits[zone] = append(this.ZoneUnits[zone], v)

	}
	//子弹分区
	this.ZoneBullets = make(map[utils.SceneZone][]*Bullet)
	for _, v := range this.Bullets {

		zone := utils.GetSceneZone((v.Position.X), (v.Position.Y))
		this.ZoneBullets[zone] = append(this.ZoneBullets[zone], v)
	}
	//光环分区
	this.ZoneHalos = make(map[utils.SceneZone][]*Halo)
	for _, v := range this.Halos {

		zone := utils.GetSceneZone((v.Position.X), (v.Position.Y))
		this.ZoneHalos[zone] = append(this.ZoneHalos[zone], v)
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
	//log.Info("sleep :%d   ", sleeptime/1e6)
	if sleeptime > 0 {
		time.Sleep(time.Duration(sleeptime))
	} else {

	}

}

//增加光环
func (this *Scene) AddHalo(halo *Halo) {
	if halo == nil {
		return
	}
	this.Halos[halo.ID] = halo

	if halo.KilledInvalid == 1 {
		this.CanRemoveHalos[halo.ID] = halo
	}

	log.Info("------AddHalo----%d", halo.TypeID)
}

//删除光环
func (this *Scene) RemoveHalo(id int32) {
	delete(this.Halos, id)
	delete(this.CanRemoveHalos, id)
}

//获取光环
func (this *Scene) ForbiddenHalo(id int32, isForbidden bool) {
	halo, ok := this.Halos[id]
	if ok && halo != nil {
		halo.IsForbidden = isForbidden
	}
}

//删除击杀单位后无效光环
func (this *Scene) RemoveHaloForKilled(parent *Unit) {
	for k, v := range this.CanRemoveHalos {
		if v.Parent == parent && v.KilledInvalid == 1 {
			delete(this.CanRemoveHalos, k)
			delete(this.Halos, k)
			return
		}
	}
}

//删除时间结束的光环
func (this *Scene) DoRemoveHalo() {
	//ZoneBullets
	for k, v := range this.Halos {
		if v.IsDone() {
			//log.Info("------DoRemoveHalo----%d", v.TypeID)
			delete(this.Halos, k)
			delete(this.CanRemoveHalos, k)
		}
	}
}

//更新光环
func (this *Scene) UpdateHalo(dt float32) {
	for _, v := range this.Halos {
		v.Update(dt)
	}
}

//增加子弹
func (this *Scene) AddBullet(bullet *Bullet) {
	this.Bullets[bullet.ID] = bullet
}

//获取子弹
func (this *Scene) GetBulletByID(id int32) *Bullet {
	re := this.Bullets[id]
	return re
}

//删除子弹
func (this *Scene) DoRemoveBullet() {
	//ZoneBullets
	for k, v := range this.Bullets {
		if v.IsDone() {
			delete(this.Bullets, k)
		}
	}
}

//更新子弹
func (this *Scene) UpdateBullet(dt float32) {
	for _, v := range this.Bullets {
		v.Update(dt)
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
		r := vec2d.Vec2{v.(*Unit).CollisionR, v.(*Unit).CollisionR}
		v.(*Unit).Body = this.MoveCore.CreateBody(pos, r, 0, 1)
		v.(*Unit).Body.BlinkToPos(v.(*Unit).InitPosition, 0)

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
		v.(*Unit).OnDestroy()
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
		//player.Characterid =
	}

	this.NextAddUnit.Set(player.MainUnit.ID, player.MainUnit)

	this.NextAddPlayer.Set(player.Uid, player)

}

func (this *Scene) RemoveUnit(unit *Unit) {
	if unit == nil {
		return

	}
	//删除主单位
	this.NextRemoveUnit.Set(unit.ID, unit)
}

//玩家退出
func (this *Scene) PlayerGoout(player *Player) {
	//删除主单位
	//this.NextRemoveUnit.Set(player.MainUnit.ID, player.MainUnit)
	this.RemoveUnit(player.MainUnit)

	this.NextRemovePlayer.Set(player.Uid, player)
}
