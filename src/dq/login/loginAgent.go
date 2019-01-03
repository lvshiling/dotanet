package login

import (
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/network"
	"dq/protobuf"
	"math/rand"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
)

type LoginAgent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *protomsg.MsgBase)
}

func (a *LoginAgent) registerDataHandle(msgtype string, f func(data *protomsg.MsgBase)) {

	a.handles[msgtype] = f

}

func (a *LoginAgent) GetConnectId() int32 {

	return 0
}
func (a *LoginAgent) GetModeType() string {
	return ""
}

func (a *LoginAgent) Init() {

	a.handles = make(map[string]func(data *protomsg.MsgBase))

	a.registerDataHandle("CS_MsgQuickLogin", a.DoQuickLoginData)

	rand.Seed(time.Now().UnixNano())
}

func (a *LoginAgent) DoQuickLoginData(data *protomsg.MsgBase) {

	log.Info("---------DoQuickLoginData")
	h2 := &protomsg.CS_MsgQuickLogin{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	//查询数据
	var uid int32
	if uid = int32(db.DbOne.CheckQuickLogin(h2.Machineid, h2.Platform)); uid > 0 {
		//log.Info("---------user login:%d", uid)
		log.Info("---------user login:%d--name:%s", uid, h2.Name)
	} else {
		uid = int32(db.DbOne.CreateQuickPlayer(h2.Machineid, h2.Platform, h2.Name))
		if uid < 0 {
			log.Info("---------new user lose", uid)
			return
		}
		log.Info("---------new user:%d", uid)
	}

	//--------------------
	a.NotifyGateLogined(data.ConnectId, uid)

	//	//回复客户端
	//	data.ModeType = "Client"
	//	data.Uid = (uid)
	//	data.MsgType = "SC_LoginResponse"
	//	jd := make(map[string]interface{})
	//	jd["result"] = 1 //成功
	//	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

}

func (a *LoginAgent) NotifyGateLogined(conncetid int32, uid int32) {

	data := &protomsg.MsgBase{}
	data.Uid = (uid)
	data.ModeType = datamsg.GateMode
	data.MsgType = "UserLogin"
	data.ConnectId = (conncetid)

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))

}

func (a *LoginAgent) Run() {

	a.Init()

	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		go a.doMessage(data)

	}
}

func (a *LoginAgent) doMessage(data []byte) {
	//log.Info("----------login----readmsg---------")
	h1 := &protomsg.MsgBase{}
	err := proto.Unmarshal(data, h1)
	if err != nil {
		log.Info("--error")
	} else {

		//log.Info("--MsgType:" + h1.MsgType)
		if f, ok := a.handles[h1.MsgType]; ok {
			f(h1)
		}

	}

}

func (a *LoginAgent) OnClose() {

}

func (a *LoginAgent) WriteMsg(msg interface{}) {

}
func (a *LoginAgent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}
func (a *LoginAgent) RegisterToGate() {
	t2 := protomsg.MsgRegisterToGate{
		ModeType: datamsg.LoginMode,
	}

	t1 := protomsg.MsgBase{
		ModeType: datamsg.GateMode,
		MsgType:  "Register",
	}

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, &t2))

}

func (a *LoginAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *LoginAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *LoginAgent) Close() {
	a.conn.Close()
}

func (a *LoginAgent) Destroy() {
	a.conn.Destroy()
}
