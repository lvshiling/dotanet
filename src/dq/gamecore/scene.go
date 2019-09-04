package gamecore

import (
	"dq/conf"
	"dq/cyward"
	"dq/log"
	//"dq/protobuf"
	//"dq/timer"
	"dq/utils"
	"dq/vec2d"
	"math/rand"
	"time"
)

type ReCreateUnit struct {
	//场景中的NPC 死亡后重新创建信息
	ReCreateInfo *conf.Unit
	DeathTime    float64 //死亡时间
}
type ChangeSceneFunc interface {
	PlayerChangeScene(player *Player, doorway conf.DoorWay)
}

type Scene struct {
	conf.SceneFileData                  //场景文件信息
	FirstUpdateTime    int64            //上次更新时间
	MoveCore           *cyward.WardCore //移动核心
	SceneName          string           //场景名字
	CurFrame           int32            //当前帧

	EveryTimeDoRemainTime float32        //每秒钟干得事的剩余时间
	DoorWays              []conf.DoorWay //传送门

	ReCreateUnitInfo map[*ReCreateUnit]*ReCreateUnit //重新创建NPC信息

	Players   map[int32]*Player           //游戏中所有的玩家
	Units     map[int32]*Unit             //游戏中所有的单位
	ZoneUnits map[utils.SceneZone][]*Unit //区域中的单位

	Bullets     map[int32]*Bullet             //游戏中所有的子弹
	ZoneBullets map[utils.SceneZone][]*Bullet //区域中的的子弹

	SceneItems     map[int32]*SceneItem             //游戏场景中所有的道具
	ZoneSceneItems map[utils.SceneZone][]*SceneItem //区域中的的道具

	Halos          map[int32]*Halo             //游戏中所有的光环
	ZoneHalos      map[utils.SceneZone][]*Halo //区域中的的光环
	CanRemoveHalos map[int32]*Halo             //可以被删除的halo 比如击杀单位后 halo无效

	NextAddUnit    *utils.BeeMap //下一帧需要增加的单位
	NextRemoveUnit *utils.BeeMap //下一帧需要删除的单位

	NextAddPlayer    *utils.BeeMap //下一帧需要增加的玩家
	NextRemovePlayer *utils.BeeMap //下一帧需要删除的玩家

	ChangeScene ChangeSceneFunc
	Quit        bool //是否退出
	SceneFrame  int32
}

func CreateScene(data *conf.SceneFileData, parent ChangeSceneFunc) *Scene {
	scene := &Scene{}
	scene.ChangeScene = parent
	scene.SceneFileData = *data
	scene.SceneName = data.ScenePath
	scene.Quit = false
	scene.Init()
	return scene
}

//初始化
func (this *Scene) Init() {
	this.SceneFrame = 40
	this.CurFrame = 0
	this.EveryTimeDoRemainTime = 1

	this.FirstUpdateTime = time.Now().UnixNano()

	this.NextAddUnit = utils.NewBeeMap()
	this.NextRemoveUnit = utils.NewBeeMap()

	this.NextAddPlayer = utils.NewBeeMap()
	this.NextRemovePlayer = utils.NewBeeMap()

	//
	this.ReCreateUnitInfo = make(map[*ReCreateUnit]*ReCreateUnit)

	this.Players = make(map[int32]*Player)
	this.Units = make(map[int32]*Unit)
	this.ZoneUnits = make(map[utils.SceneZone][]*Unit)

	this.Bullets = make(map[int32]*Bullet)
	this.ZoneBullets = make(map[utils.SceneZone][]*Bullet)

	this.SceneItems = make(map[int32]*SceneItem)
	this.ZoneSceneItems = make(map[utils.SceneZone][]*SceneItem)

	this.Halos = make(map[int32]*Halo)
	this.ZoneHalos = make(map[utils.SceneZone][]*Halo)
	this.CanRemoveHalos = make(map[int32]*Halo)

	scenedata := conf.GetSceneData(this.ScenePath)

	//场景碰撞区域
	this.MoveCore = cyward.CreateWardCore()
	for _, v := range scenedata.Collides {
		//log.Info("Collide %v", v)
		if v.IsRect == true {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			r := vec2d.Vec2{v.Width, v.Height}
			this.MoveCore.CreateBody(pos, r, 1, 2)
		} else {
			pos := vec2d.Vec2{v.CenterX, v.CenterY}
			this.MoveCore.CreateBodyPolygon(pos, v.Points, 1, 2)
		}
	}
	//场景分区数据 创建100个单位
	createunitdata := conf.GetCreateUnitData(this.CreateUnit)
	for _, v := range createunitdata.Units {
		oneunit := this.CreateUnitByConf(v)
		if oneunit != nil {
			//			item := NewItem(20)
			//			oneunit.AddItem(-1, item)
			log.Info("createunity")
		}

		//log.Info("createunity :%v", v)
	}
	//传送门显示
	this.DoorWays = createunitdata.DoorWays
	for _, v := range this.DoorWays {
		halo := NewHalo(v.HaloTypeID, 1)
		halo.SetParent(nil)
		halo.IsActive = false
		halo.Position = vec2d.Vec2{X: v.X, Y: v.Y}
		if halo != nil {
			this.AddHalo(halo)
		}
	}

	//创建道具
	this.CreateSceneItem(1, vec2d.Vec2{X: 34, Y: 84})
	//创建道具
	this.CreateSceneItem(2, vec2d.Vec2{X: 36, Y: 84})

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
func (this *Scene) CreateUnitByConf(v conf.Unit) *Unit {
	unit := CreateUnit(this, v.TypeID)
	unit.SetAI(NewNormalAI(unit))
	pos := vec2d.Vec2{v.X, v.Y}
	r := vec2d.Vec2{unit.CollisionR, unit.CollisionR}
	unit.Body = this.MoveCore.CreateBody(pos, r, 0, 1)
	unit.Body.Direction = vec2d.Vec2{X: 0, Y: 1}
	unit.Body.Direction.Rotate(-v.Rotation)
	unit.SetReCreateInfo(&v)
	//log.Info("Collide %v  %f", unit.Body.Direction, v.Rotation)
	this.Units[unit.ID] = unit
	return unit
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

//传送门检查
func (this *Scene) DoDoorWay() {
	if this.ChangeScene == nil {
		return
	}

	for _, v := range this.DoorWays {
		pos := vec2d.Vec2{X: v.X, Y: v.Y}
		for _, player := range this.Players {
			if player != nil && player.MainUnit != nil &&
				player.MainUnit.Body != nil && player.MainUnit.Level >= v.NeedLevel {
				mypos := player.MainUnit.Body.Position
				subpos := vec2d.Sub(mypos, pos)
				distanse := subpos.Length()
				if distanse <= v.R {
					//传送到其他场景
					log.Info("chuan song :%v", v)
					this.ChangeScene.PlayerChangeScene(player, v)
				}
			}
		}
	}
}

//
func (this *Scene) EveryTimeDo(dt float32) {

	this.EveryTimeDoRemainTime -= float32(dt)
	if this.EveryTimeDoRemainTime <= 0 {
		//do
		this.EveryTimeDoRemainTime += 1
		this.DoReCreateUnit()
		this.DoDoorWay()

		//道具
		this.UpdateSceneItem(dt)
		this.DoRemoveSceneItem()

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

		time1 := utils.GetCurTimeOfSecond()
		this.EveryTimeDo(1 / float32(this.SceneFrame))

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
			//ZoneSceneItems
			if _, ok := this.ZoneSceneItems[vzone]; ok {
				//遍历区域中的单位
				for _, sceneitem := range this.ZoneSceneItems[vzone] {
					player.AddSceneItemData(sceneitem)
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
	//道具分区
	this.ZoneSceneItems = make(map[utils.SceneZone][]*SceneItem)
	for _, v := range this.SceneItems {

		zone := utils.GetSceneZone((v.Position.X), (v.Position.Y))
		this.ZoneSceneItems[zone] = append(this.ZoneSceneItems[zone], v)
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

//单位死亡后创建道具
func (this Scene) CreateSceneItems(typeid []int32, centerpos vec2d.Vec2) {
	//CheckItemCollision
	positions := make([]vec2d.Vec2, 8)
	size := 1.0
	positions[0] = vec2d.Vec2{X: centerpos.X - size, Y: centerpos.Y + size}
	positions[1] = vec2d.Vec2{X: centerpos.X, Y: centerpos.Y + size}
	positions[2] = vec2d.Vec2{X: centerpos.X + size, Y: centerpos.Y + size}

	positions[3] = vec2d.Vec2{X: centerpos.X - size, Y: centerpos.Y}
	positions[4] = vec2d.Vec2{X: centerpos.X + size, Y: centerpos.Y}

	positions[5] = vec2d.Vec2{X: centerpos.X - size, Y: centerpos.Y - size}
	positions[6] = vec2d.Vec2{X: centerpos.X, Y: centerpos.Y - size}
	positions[7] = vec2d.Vec2{X: centerpos.X + size, Y: centerpos.Y - size}

	createitemid := 0
	startindex := rand.Intn(8)
	for i := 0; i < 8; i++ {
		posindex := startindex + i
		if posindex >= 8 {
			posindex = 0
		}
		if this.MoveCore.CheckItemCollision(positions[posindex]) == true {
			this.CreateSceneItem(typeid[createitemid], positions[posindex])
			createitemid++
			if createitemid >= len(typeid) {
				break
			}
		}
	}
}

//创建场景道具
func (this *Scene) CreateSceneItem(typeid int32, pos vec2d.Vec2) {
	sceneitem := NewSceneItem(typeid, pos)
	this.AddSceneItem(sceneitem)
}

//增加场景道具
func (this *Scene) AddSceneItem(sceneitem *SceneItem) {
	this.SceneItems[sceneitem.ID] = sceneitem
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

//删除道具
func (this *Scene) DoRemoveSceneItem() {
	//ZoneBullets
	for k, v := range this.SceneItems {
		if v.IsDone() {
			delete(this.SceneItems, k)
		}
	}
}

//更新道具
func (this *Scene) UpdateSceneItem(dt float32) {
	for _, v := range this.SceneItems {
		v.Update()
		if v.IsDone() == false {
			//遍历所有玩家
			for _, player := range this.Players {
				unit := player.MainUnit
				if unit == nil || unit.Body == nil {
					continue
				}
				//				if player.CanSelectSceneItem() == false {
				//					continue
				//				}
				//LengthSquared
				dir := vec2d.Sub(unit.Body.Position, v.Position)
				if dir.LengthSquared() <= 1 {
					if player.SelectSceneItem(v) == true {
						v.BeSelect()
						break
					}
				}
			}
		}
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

//处理重新创建NPC
func (this *Scene) DoReCreateUnit() {
	curtime := utils.GetCurTimeOfSecond()
	for k, v := range this.ReCreateUnitInfo {

		if curtime-v.DeathTime >= v.ReCreateInfo.ReCreateTime {
			this.CreateUnitByConf(*(v.ReCreateInfo))
			delete(this.ReCreateUnitInfo, k)
		}
	}
}

func (this *Scene) RemoveUnit(unit *Unit) {
	if unit == nil {
		return

	}
	//删除主单位
	this.NextRemoveUnit.Set(unit.ID, unit)
	if unit.ReCreateInfo != nil {
		rc := &ReCreateUnit{}
		rc.ReCreateInfo = unit.ReCreateInfo
		rc.DeathTime = utils.GetCurTimeOfSecond()
		this.ReCreateUnitInfo[rc] = rc
		//this.ReCreateUnitInfo = append(this.ReCreateUnitInfo)
	}

}

//玩家退出
func (this *Scene) PlayerGoout(player *Player) {
	//删除主单位
	//this.NextRemoveUnit.Set(player.MainUnit.ID, player.MainUnit)
	this.RemoveUnit(player.MainUnit)
	items := player.OtherUnit.Items()
	for _, v := range items {
		this.RemoveUnit(v.(*Unit))
	}

	this.NextRemovePlayer.Set(player.Uid, player)
}
