package gamecore

import (
	//"dq/conf"
	//"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
)

var SceneItemID int32 = 100

func GetSceneItemID() int32 {
	SceneItemID++
	if SceneItemID >= 100000000 {
		SceneItemID = 100
	}
	return SceneItemID
}

//场景里的道具
type SceneItem struct {
	ID         int32
	TypeID     int32      //类型
	Position   vec2d.Vec2 //场景里的位置
	LifeTime   float64    //剩余生命时间 时间到了就删除
	CreateTime float64    // 创建的时间

	IsOver bool //是否结束

	//发送数据部分
	ClientData *protomsg.SceneItemDatas //客户端显示数据
	//ClientDataSub *protomsg.BulletDatas //客户端显示差异数据
}

//刷新客户端显示数据
func (this *SceneItem) FreshClientData() {

	if this.ClientData == nil {
		this.ClientData = &protomsg.SceneItemDatas{}
	}

	this.ClientData.ID = this.ID
	this.ClientData.TypeID = this.TypeID
	this.ClientData.X = float32(this.Position.X)
	this.ClientData.Y = float32(this.Position.Y)
}

//更新 没秒一次
func (this *SceneItem) Update() {
	curtime := utils.GetCurTimeOfSecond()
	if curtime-this.CreateTime > this.LifeTime {
		this.IsOver = true
	}
}

func (this *SceneItem) IsDone() bool {
	return this.IsOver
}

//创建buf
func NewSceneItem(typeid int32, pos vec2d.Vec2) *SceneItem {
	item := &SceneItem{}
	item.TypeID = typeid
	item.Position = pos
	item.LifeTime = 11120.0
	item.CreateTime = utils.GetCurTimeOfSecond()
	item.IsOver = false
	//唯一ID处理
	item.ID = GetSceneItemID()

	item.FreshClientData()
	return item
}
