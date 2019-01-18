package gamecore

import (
	"dq/datamsg"
	"dq/protobuf"
)

type Server interface {
	WriteMsgBytes(msg []byte)
}

type Player struct {
	Uid         int32
	ConnectId   int32
	MainUnit    *Unit //主单位
	CurScene    *Scene
	ServerAgent Server

	//OtherUnit  *Unit //其他单位

	//组合数据包相关
	LastUnitDatas map[int32]*protomsg.UnitDatas
	CurUnitDatas  map[int32]*protomsg.UnitDatas
	Msg           *protomsg.SC_Update
}

func CreatePlayer(uid int32, connectid int32) *Player {
	re := &Player{}
	re.Uid = uid
	re.ConnectId = connectid
	re.LastUnitDatas = make(map[int32]*protomsg.UnitDatas)
	re.CurUnitDatas = make(map[int32]*protomsg.UnitDatas)

	return re
}

//获取差异包
//func GetChangedValueOfUnitDatas(last *protomsg.UnitDatas, cur *protomsg.UnitDatas) *protomsg.UnitDatas {
//	re := &protomsg.UnitDatas{}

//}

//开始组合客户端显示单位数据
func (this *Player) StartComClientData() {

	this.CurUnitDatas = make(map[int32]*protomsg.UnitDatas)
	//
	this.Msg = &protomsg.SC_Update{}

	//	for _,v := range this.LastUnitDatas{
	//		this.Msg.RemoveUnits = append()
	//	}
}

//添加客户端显示单位数据包
func (this *Player) AddUnitData(data protomsg.UnitDatas) {

	this.CurUnitDatas[data.ID] = &data

	if _, ok := this.LastUnitDatas[data.ID]; ok {
		//旧单位(只更新变化的值)
	} else {
		//新的单位数据
		this.Msg.NewUnits = append(this.Msg.NewUnits, &data)
	}

}

func (this *Player) SendMsg() {

	//删除的单位 id
	for k, _ := range this.LastUnitDatas {
		if _, ok := this.CurUnitDatas[k]; !ok {
			this.Msg.RemoveUnits = append(this.Msg.RemoveUnits, k)
		}
	}

	//回复客户端
	data := &protomsg.MsgBase{}
	data.ConnectId = this.ConnectId
	data.ModeType = "Client"
	data.Uid = this.Uid
	data.MsgType = "SC_Update"

	this.ServerAgent.WriteMsgBytes(datamsg.NewMsg1Bytes(data, this.Msg))

	this.LastUnitDatas = this.CurUnitDatas
}

//退出场景
func (this *Player) OutScene() {
	if this.CurScene != nil {
		this.CurScene.PlayerGoout(this)
	}
}

//进入场景
func (this *Player) GoInScene(scene *Scene, datas []byte) {
	if this.CurScene != nil {
		this.CurScene.PlayerGoout(this)
	}
	this.CurScene = scene

	this.CurScene.PlayerGoin(this, datas)
}
