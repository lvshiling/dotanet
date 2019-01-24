package gamecore

import (
	"dq/datamsg"
	"dq/protobuf"

	"github.com/golang/protobuf/proto"
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
	LastShowUnit map[int32]*Unit
	CurShowUnit  map[int32]*Unit
	Msg          *protomsg.SC_Update
}

func CreatePlayer(uid int32, connectid int32) *Player {
	re := &Player{}
	re.Uid = uid
	re.ConnectId = connectid
	re.LastShowUnit = make(map[int32]*Unit)
	re.CurShowUnit = make(map[int32]*Unit)
	re.Msg = &protomsg.SC_Update{}
	return re
}

//添加客户端显示单位数据包
func (this *Player) AddUnitData(unit *Unit) {

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

func (this *Player) SendUpdateMsg(curframe int32) {

	//删除的单位 id
	for k, _ := range this.LastShowUnit {
		if _, ok := this.CurShowUnit[k]; !ok {
			this.Msg.RemoveUnits = append(this.Msg.RemoveUnits, k)
		}
	}

	//回复客户端
	this.Msg.CurFrame = curframe
	this.SendMsgToClient("SC_Update", this.Msg)

	//重置数据
	this.LastShowUnit = this.CurShowUnit
	this.CurShowUnit = make(map[int32]*Unit)
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
}

//进入场景
func (this *Player) GoInScene(scene *Scene, datas []byte) {
	if this.CurScene != nil {
		this.CurScene.PlayerGoout(this)
	}
	this.CurScene = scene

	this.CurScene.PlayerGoin(this, datas)
}

//玩家移动操作
func (this *Player) MoveCmd(data *protomsg.CS_PlayerMove) {
	for _, v := range data.IDs {
		if this.MainUnit.ID == v {
			this.MainUnit.MoveCmd(data)
		}
	}
}
