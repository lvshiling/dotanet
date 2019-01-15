package gamecore

import "dq/cyward"

var UnitID int32 = 10

type BaseProperty struct {
	HP     int32
	MAX_HP int32
	MP     int32
	MAX_MP int32
}

type Unit struct {
	BaseProperty
	InScene *Scene
	ID      int32        //单位唯一ID
	Body    *cyward.Body //移动相关(位置信息) 需要设置移动速度

}

func CreateUnit(scene *Scene) *Unit {
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
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
