package datamsg

import (
	"dq/protobuf"

	"github.com/golang/protobuf/proto"
)

var LoginMode = "Login"   //登录模块
var GateMode = "Gate"     //网关模块
var ClientMode = "Client" //客户端模块

var HallMode = "Hall" //大厅模块

var GameScene1 = "GameScene1" //游戏场景1

//消息类型
var SC_Heart = "SC_Heart"
var CS_Heart = "CS_Heart"

func NewMsg1Bytes(data *protomsg.MsgBase, jd proto.Message) []byte {

	if jd != nil {
		jdbytes, _ := proto.Marshal(jd)
		data.Datas = jdbytes
	}
	data1, err1 := proto.Marshal(data)
	if err1 == nil {
		return data1
	}
	return []byte("")
}
