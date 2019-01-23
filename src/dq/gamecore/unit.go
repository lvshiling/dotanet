package gamecore

import (
	"dq/cyward"
	//"dq/log"
	"dq/protobuf"
	"dq/utils"
	"strings"
)

var UnitID int32 = 10

type UnitProperty struct {
	//基础数据 当前数据
	HP            int32
	MAX_HP        int32
	MP            int32
	MAX_MP        int32
	Name          string
	Level         int32
	ModeType      string
	Experience    int32
	MaxExperience int32

	BaseAttack      int32   //基础攻击力
	BaseAttackRange float32 //基础攻击范围

	//复合数据 会随时变动的数据 比如受buff影响攻击力降低  (每帧动态计算)
	Attack      int32   //攻击力 (基础攻击力+属性影响+buff影响)
	AttackRange float32 //攻击范围
}

type Unit struct {
	UnitProperty
	InScene  *Scene
	MyPlayer *Player
	ID       int32        //单位唯一ID
	Body     *cyward.Body //移动相关(位置信息) 需要设置移动速度
	State    UnitState    //逻辑状态

	//发送数据部分
	ClientData    *protomsg.UnitDatas //客户端显示数据
	ClientDataSub *protomsg.UnitDatas //客户端显示差异数据
}

func CreateUnit(scene *Scene) *Unit {
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	unitre.Init()

	return unitre
}

func CreateUnitByPlayer(scene *Scene, id int32, player *Player, datas []byte) *Unit {
	unitre := &Unit{}
	unitre.ID = id
	unitre.InScene = scene
	unitre.MyPlayer = player
	utils.Bytes2Struct(datas, &unitre.UnitProperty)
	unitre.Init()

	return unitre
}

//初始化
func (this *Unit) Init() {

}

//
//更新 范围影响的buff
func (this *Unit) PreUpdate(dt float64) {

}

//行为命令
func (this *Unit) DoCommand() {

}

//更新
func (this *Unit) Update(dt float64) {
	//设置是否有碰撞  设置移动速度 和逻辑状态

	//

	//逻辑状态更新
	this.State.Update(dt)
}

//刷新客户端显示数据
func (this *Unit) FreshClientData() {

	if this.ClientData == nil {
		this.ClientData = &protomsg.UnitDatas{}
	}

	this.ClientData.HP = this.HP
	this.ClientData.MaxHP = this.MAX_HP
	this.ClientData.MP = this.MP
	this.ClientData.MaxMP = this.MAX_MP
	this.ClientData.Name = this.Name
	this.ClientData.Level = this.Level
	this.ClientData.ID = this.ID
	this.ClientData.ModeType = this.ModeType
	this.ClientData.Experience = this.Experience
	this.ClientData.MaxExperience = this.MaxExperience

	this.ClientData.X = float32(this.Body.Position.X)
	this.ClientData.Y = float32(this.Body.Position.Y)
}

//func (this *Unit) OnePropSub(prop interface{}){

//}

//刷新客户端显示差异数据
func (this *Unit) FreshClientDataSub() {

	if this.ClientDataSub == nil {
		this.ClientDataSub = &protomsg.UnitDatas{}
	}
	if this.ClientData == nil {
		this.FreshClientData()
		*this.ClientDataSub = *this.ClientData
		return
	}

	//当前数据与上一次数据对比 相减
	//字符串部分
	if strings.Compare(this.Name, this.ClientData.Name) != 0 {
		this.ClientDataSub.Name = this.Name
	}
	if strings.Compare(this.ModeType, this.ClientData.ModeType) != 0 {
		this.ClientDataSub.ModeType = this.ModeType
	}
	//数值部分
	this.ClientDataSub.HP = this.HP - this.ClientData.HP
	this.ClientDataSub.MaxHP = this.MAX_HP - this.ClientData.MaxHP
	this.ClientDataSub.MP = this.MP - this.ClientData.MP
	this.ClientDataSub.MaxMP = this.MAX_MP - this.ClientData.MaxMP
	this.ClientDataSub.Level = this.Level - this.ClientData.Level
	this.ClientDataSub.Experience = this.Experience - this.ClientData.Experience
	this.ClientDataSub.MaxExperience = this.MaxExperience - this.ClientData.MaxExperience
	this.ClientDataSub.X = float32(this.Body.Position.X) - this.ClientData.X
	this.ClientDataSub.Y = float32(this.Body.Position.Y) - this.ClientData.Y
}
