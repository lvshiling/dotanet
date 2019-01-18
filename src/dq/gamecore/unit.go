package gamecore

import (
	"dq/cyward"
	"dq/protobuf"
)

var UnitID int32 = 10

type BaseProperty struct {
	HP     int32
	MAX_HP int32
	MP     int32
	MAX_MP int32
}

type Unit struct {
	BaseProperty
	InScene  *Scene
	MyPlayer *Player
	ID       int32        //单位唯一ID
	Body     *cyward.Body //移动相关(位置信息) 需要设置移动速度

	ClientData protomsg.UnitDatas //客户端显示数据
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
	unitre.Init()

	return unitre
}

//初始化
func (this *Unit) Init() {
	this.HP = 500
	this.MAX_HP = 500
	this.MP = 100
	this.MAX_MP = 100
}

//
//更新 范围影响的buff
func (this *Unit) PreUpdate(dt float64) {

}

//更新
func (this *Unit) Update(dt float64) {
	//设置是否有碰撞  设置移动速度 和逻辑状态
}

//type UnitDatas struct {
//	Name                 string   `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
//	Level                int32    `protobuf:"varint,2,opt,name=Level,proto3" json:"Level,omitempty"`
//	HP                   int32    `protobuf:"varint,3,opt,name=HP,proto3" json:"HP,omitempty"`
//	MP                   int32    `protobuf:"varint,4,opt,name=MP,proto3" json:"MP,omitempty"`
//	X                    float32  `protobuf:"fixed32,5,opt,name=X,proto3" json:"X,omitempty"`
//	Y                    float32  `protobuf:"fixed32,6,opt,name=Y,proto3" json:"Y,omitempty"`
//	ID                   int32    `protobuf:"varint,7,opt,name=ID,proto3" json:"ID,omitempty"`
//	ModeType             string   `protobuf:"bytes,8,opt,name=ModeType,proto3" json:"ModeType,omitempty"`
//	MaxHP                int32    `protobuf:"varint,9,opt,name=MaxHP,proto3" json:"MaxHP,omitempty"`
//	MaxMP                int32    `protobuf:"varint,10,opt,name=MaxMP,proto3" json:"MaxMP,omitempty"`
//	Experience           int32    `protobuf:"varint,11,opt,name=Experience,proto3" json:"Experience,omitempty"`
//	XXX_NoUnkeyedLiteral struct{} `json:"-"`
//	XXX_unrecognized     []byte   `json:"-"`
//	XXX_sizecache        int32    `json:"-"`
//}
//刷新客户端显示数据
func (this *Unit) FreshClientData() {
	this.ClientData.HP = this.HP
	this.ClientData.MaxHP = this.MAX_HP
	this.ClientData.MP = this.MP
	this.ClientData.MaxMP = this.MAX_MP
	this.ClientData.Name = "test1"
	this.ClientData.Level = 1
	this.ClientData.ID = this.ID
	this.ClientData.ModeType = "Hero/hero1"
	this.ClientData.Experience = 0
	this.ClientData.MaxExperience = 100

	this.ClientData.X = float32(this.Body.Position.X)
	this.ClientData.Y = float32(this.Body.Position.Y)
}
