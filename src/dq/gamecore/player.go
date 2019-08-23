package gamecore

import (
	"dq/datamsg"
	"dq/db"
	"dq/protobuf"
	"dq/utils"
	"strconv"

	"github.com/golang/protobuf/proto"
)

type Server interface {
	WriteMsgBytes(msg []byte)
}

type Player struct {
	Uid         int32
	ConnectId   int32
	Characterid int32         //角色id
	MainUnit    *Unit         //主单位
	OtherUnit   *utils.BeeMap //其他单位
	CurScene    *Scene
	ServerAgent Server

	//OtherUnit  *Unit //其他单位

	//组合数据包相关
	LastShowUnit   map[int32]*Unit
	CurShowUnit    map[int32]*Unit
	LastShowBullet map[int32]*Bullet
	CurShowBullet  map[int32]*Bullet
	LastShowHalo   map[int32]*Halo
	CurShowHalo    map[int32]*Halo

	LastShowSceneItem map[int32]*SceneItem
	CurShowSceneItem  map[int32]*SceneItem
	Msg               *protomsg.SC_Update
}

func CreatePlayer(uid int32, connectid int32, characterid int32) *Player {
	re := &Player{}
	re.Uid = uid
	re.ConnectId = connectid
	re.Characterid = characterid
	re.ReInit()
	return re
}
func (this *Player) ReInit() {
	this.MainUnit = nil
	this.LastShowUnit = make(map[int32]*Unit)
	this.CurShowUnit = make(map[int32]*Unit)
	this.LastShowBullet = make(map[int32]*Bullet)
	this.CurShowBullet = make(map[int32]*Bullet)
	this.LastShowHalo = make(map[int32]*Halo)
	this.CurShowHalo = make(map[int32]*Halo)
	this.LastShowSceneItem = make(map[int32]*SceneItem)
	this.CurShowSceneItem = make(map[int32]*SceneItem)

	//
	this.OtherUnit = utils.NewBeeMap()
	this.Msg = &protomsg.SC_Update{}
}

func (this *Player) AddOtherUnit(unit *Unit) {
	if unit == nil {
		return
	}
	unit.ControlID = this.Uid
	this.OtherUnit.Set(unit.ID, unit)
}

//是否可以拾取地面的物品
func (this *Player) CanSelectSceneItem() bool {
	if this.MainUnit == nil || this.MainUnit.Items == nil {
		return false
	}
	for _, v := range this.MainUnit.Items {
		if v == nil {
			return true
		}
	}

	return false
}

//拾取地面物品
func (this *Player) SelectSceneItem(sceneitem *SceneItem) bool {
	if this.CanSelectSceneItem() == false {
		return false
	}
	item := NewItem(sceneitem.TypeID)
	this.MainUnit.AddItem(-1, item)

	return true
}

//遍历删除无效的
func (this *Player) CheckOtherUnit() {

	items := this.OtherUnit.Items()
	for k, v := range items {
		if v == nil || v.(*Unit).IsDisappear() {
			this.OtherUnit.Delete(k)
		}
	}

}

//type DB_CharacterInfo struct {
//	Characterid int32   `json:"characterid"`
//	Uid         int32   `json:"uid"`
//	Name        string  `json:"name"`
//	Typeid      int32   `json:"typeid"`
//	Level       int32   `json:"level"`
//	Experience  int32   `json:"experience"`
//	Gold        int32   `json:"gold"`
//	HP          float32 `json:"hp"`
//	MP          float32 `json:"mp"`
//	SceneName   string  `json:"scenename"`
//	X           float32 `json:"x"`
//	Y           float32 `json:"y"`
//}

func (this *Player) GetDBData() *db.DB_CharacterInfo {
	if this.MainUnit == nil {
		return nil
	}
	dbdata := db.DB_CharacterInfo{}
	dbdata.Characterid = this.Characterid
	dbdata.Uid = this.Uid
	dbdata.Name = this.MainUnit.Name
	dbdata.Typeid = this.MainUnit.TypeID
	dbdata.Level = this.MainUnit.Level
	dbdata.Experience = this.MainUnit.Experience
	//dbdata.Gold = this.MainUnit.Gold
	dbdata.HP = float32(this.MainUnit.HP) / float32(this.MainUnit.MAX_HP)
	dbdata.MP = float32(this.MainUnit.MP) / float32(this.MainUnit.MAX_MP)
	if this.CurScene != nil {
		dbdata.SceneID = this.CurScene.TypeID
	} else {
		dbdata.SceneID = 1
	}
	if this.MainUnit.Body == nil {
		dbdata.X = 0
		dbdata.Y = 0
	} else {
		dbdata.X = float32(this.MainUnit.Body.Position.X)
		dbdata.Y = float32(this.MainUnit.Body.Position.Y)
	}
	//技能
	for _, v := range this.MainUnit.Skills {
		dbdata.Skill += v.ToDBString() + ";"
	}
	//道具
	if this.MainUnit.Items != nil {
		item1 := this.MainUnit.Items[0]
		if item1 == nil {
			dbdata.Item1 = ""
		} else {
			dbdata.Item1 = strconv.Itoa(int(item1.TypeID))
		}
		item2 := this.MainUnit.Items[1]
		if item2 == nil {
			dbdata.Item2 = ""
		} else {
			dbdata.Item2 = strconv.Itoa(int(item2.TypeID))
		}
		item3 := this.MainUnit.Items[2]
		if item3 == nil {
			dbdata.Item3 = ""
		} else {
			dbdata.Item3 = strconv.Itoa(int(item3.TypeID))
		}
		item4 := this.MainUnit.Items[3]
		if item4 == nil {
			dbdata.Item4 = ""
		} else {
			dbdata.Item4 = strconv.Itoa(int(item4.TypeID))
		}
		item5 := this.MainUnit.Items[4]
		if item5 == nil {
			dbdata.Item5 = ""
		} else {
			dbdata.Item5 = strconv.Itoa(int(item5.TypeID))
		}
		item6 := this.MainUnit.Items[5]
		if item6 == nil {
			dbdata.Item6 = ""
		} else {
			dbdata.Item6 = strconv.Itoa(int(item6.TypeID))
		}
	}

	return &dbdata
}

//存档数据
func (this *Player) SaveDB() {
	dbdata := this.GetDBData()
	db.DbOne.SaveCharacter(*dbdata)

}

//清除显示状态  切换场景的时候需要调用
func (this *Player) ClearShowData() {
	this.LastShowUnit = make(map[int32]*Unit)
	this.CurShowUnit = make(map[int32]*Unit)
	this.LastShowBullet = make(map[int32]*Bullet)
	this.CurShowBullet = make(map[int32]*Bullet)
	this.LastShowHalo = make(map[int32]*Halo)
	this.CurShowHalo = make(map[int32]*Halo)

	this.LastShowSceneItem = make(map[int32]*SceneItem)
	this.CurShowSceneItem = make(map[int32]*SceneItem)
	//

	this.Msg = &protomsg.SC_Update{}
}

//添加客户端显示单位数据包
func (this *Player) AddUnitData(unit *Unit) {

	if this.MainUnit == nil {
		return
	}
	//检查玩家主单位是否能看见 目标单位
	if this.MainUnit.CanSeeTarget(unit) == false {
		return
	}

	this.CurShowUnit[unit.ID] = unit

	if _, ok := this.LastShowUnit[unit.ID]; ok {
		//旧单位(只更新变化的值)
		d1 := *unit.ClientDataSub
		this.Msg.OldUnits = append(this.Msg.OldUnits, &d1)
	} else {
		//新的单位数据
		d1 := *unit.ClientData
		this.Msg.NewUnits = append(this.Msg.NewUnits, &d1)
	}

}

//
//添加客户端显示子弹数据包
func (this *Player) AddHaloData(halo *Halo) {

	//如果客户端不需要显示
	if halo.ClientIsShow() == false {
		return
	}

	this.CurShowHalo[halo.ID] = halo

	if _, ok := this.LastShowHalo[halo.ID]; ok {
		//旧单位(只更新变化的值)
		d1 := *halo.ClientDataSub
		this.Msg.OldHalos = append(this.Msg.OldHalos, &d1)
	} else {
		//新的单位数据
		d1 := *halo.ClientData
		this.Msg.NewHalos = append(this.Msg.NewHalos, &d1)
	}

}

//this.LastShowSceneItem = make(map[int32]*SceneItem)
//this.CurShowSceneItem = make(map[int32]*SceneItem)
//添加客户端显示子弹数据包
func (this *Player) AddSceneItemData(sceneitem *SceneItem) {

	this.CurShowSceneItem[sceneitem.ID] = sceneitem

	if _, ok := this.LastShowSceneItem[sceneitem.ID]; ok {
		//旧单位(只更新变化的值)
		//d1 := *bullet.ClientDataSub
		//this.Msg.OldBullets = append(this.Msg.OldBullets, &d1)
	} else {
		//新的单位数据
		d1 := *sceneitem.ClientData
		this.Msg.NewSceneItems = append(this.Msg.NewSceneItems, &d1)
	}

}

//添加客户端显示子弹数据包
func (this *Player) AddBulletData(bullet *Bullet) {

	//如果客户端不需要显示
	if bullet.ClientIsShow() == false {
		return
	}

	this.CurShowBullet[bullet.ID] = bullet

	if _, ok := this.LastShowBullet[bullet.ID]; ok {
		//旧单位(只更新变化的值)
		d1 := *bullet.ClientDataSub
		this.Msg.OldBullets = append(this.Msg.OldBullets, &d1)
	} else {
		//新的单位数据
		d1 := *bullet.ClientData
		this.Msg.NewBullets = append(this.Msg.NewBullets, &d1)
	}

}
func (this *Player) AddHurtValue(hv *protomsg.MsgPlayerHurt) {
	if hv == nil || hv.HurtAllValue == 0 {
		return
	}

	this.Msg.PlayerHurt = append(this.Msg.PlayerHurt, hv)
}

func (this *Player) SendUpdateMsg(curframe int32) {

	//删除的单位 id
	for k, _ := range this.LastShowUnit {
		if _, ok := this.CurShowUnit[k]; !ok {
			this.Msg.RemoveUnits = append(this.Msg.RemoveUnits, k)
		}
	}
	//删除的子弹 id
	for k, _ := range this.LastShowBullet {
		if _, ok := this.CurShowBullet[k]; !ok {
			this.Msg.RemoveBullets = append(this.Msg.RemoveBullets, k)
		}
	}
	//Halo
	//删除的Halo id
	for k, _ := range this.LastShowHalo {
		if _, ok := this.CurShowHalo[k]; !ok {
			this.Msg.RemoveHalos = append(this.Msg.RemoveHalos, k)
		}
	}
	//删除的sceneitem id
	for k, _ := range this.LastShowSceneItem {
		if _, ok := this.CurShowSceneItem[k]; !ok {
			this.Msg.RemoveSceneItems = append(this.Msg.RemoveSceneItems, k)
		}
	}

	//回复客户端
	this.Msg.CurFrame = curframe
	this.SendMsgToClient("SC_Update", this.Msg)

	//重置数据
	this.LastShowUnit = this.CurShowUnit
	this.CurShowUnit = make(map[int32]*Unit)

	this.LastShowBullet = this.CurShowBullet
	this.CurShowBullet = make(map[int32]*Bullet)

	this.LastShowHalo = this.CurShowHalo
	this.CurShowHalo = make(map[int32]*Halo)

	this.LastShowSceneItem = this.CurShowSceneItem
	this.CurShowSceneItem = make(map[int32]*SceneItem)
	this.Msg = &protomsg.SC_Update{}

}

func (this *Player) SendMsgToClient(msgtype string, msg proto.Message) {
	data := &protomsg.MsgBase{}
	data.ConnectId = this.ConnectId
	data.ModeType = "Client"
	data.Uid = this.Uid
	data.MsgType = msgtype

	this.ServerAgent.WriteMsgBytes(datamsg.NewMsg1Bytes(data, msg))

}

//退出场景
func (this *Player) OutScene() {

	if this.CurScene != nil {
		this.CurScene.PlayerGoout(this)
	}
	this.ReInit()

}

//进入场景
func (this *Player) GoInScene(scene *Scene, datas []byte) {
	if this.CurScene != nil {
		this.CurScene.PlayerGoout(this)
	}
	this.CurScene = scene

	this.CurScene.PlayerGoin(this, datas)
	//this.ReInit()
}

//玩家移动操作
func (this *Player) MoveCmd(data *protomsg.CS_PlayerMove) {
	this.CheckOtherUnit()
	for _, v := range data.IDs {
		if this.MainUnit.ID == v {
			this.MainUnit.MoveCmd(data)

			this.CheckOtherUnit()
			items := this.OtherUnit.Items()
			for _, v1 := range items {
				v1.(*Unit).PlayerControl_MoveCmd(data)
			}
		}

	}
}

//SkillCmd
func (this *Player) SkillCmd(data *protomsg.CS_PlayerSkill) {
	this.CheckOtherUnit()
	if this.MainUnit.ID == data.ID {
		this.MainUnit.PlayerControl_SkillCmd(data)
	}
}

//玩家攻击操作
func (this *Player) AttackCmd(data *protomsg.CS_PlayerAttack) {
	this.CheckOtherUnit()
	for _, v := range data.IDs {
		if this.MainUnit.ID == v {
			this.MainUnit.AttackCmd(data)

			this.CheckOtherUnit()
			items := this.OtherUnit.Items()
			for _, v1 := range items {
				v1.(*Unit).PlayerControl_AttackCmd(data)
			}
		}
	}
}
