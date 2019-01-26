package login

import (
	"dq/datamsg"
	"dq/db"
	"dq/gamecore"
	"dq/log"
	"dq/network"
	"dq/protobuf"
	"dq/utils"
	"math/rand"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
)

type LoginAgent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *protomsg.MsgBase)

	LoginPlayers *utils.BeeMap
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

	a.LoginPlayers = utils.NewBeeMap()

	a.handles = make(map[string]func(data *protomsg.MsgBase))

	a.registerDataHandle("CS_MsgQuickLogin", a.DoQuickLoginData)

	a.registerDataHandle("LoginOut", a.DoLoginOut)

	rand.Seed(time.Now().UnixNano())
}

//DoLoginOut
func (a *LoginAgent) DoLoginOut(data *protomsg.MsgBase) {
	a.LoginPlayers.Delete(data.Uid)
	log.Info("-------loginout :%d", data.Uid)
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

	//检查是否重复登录
	if a.LoginPlayers.Check(uid) {
		a.ReLoginForceDisconnect(a.LoginPlayers.Get(uid).(int32), uid)

		//回复客户端
		data.ModeType = "Client"
		data.Uid = (uid)
		data.MsgType = "SC_Logined"
		jd := &protomsg.SC_Logined{}
		jd.Code = 0 //失败
		jd.Uid = (uid)
		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

		return
	}
	a.LoginPlayers.Set(uid, data.ConnectId)

	//--------------------
	a.NotifyGateLogined(data.ConnectId, uid)

	//回复客户端
	data.ModeType = "Client"
	data.Uid = (uid)
	data.MsgType = "SC_Logined"
	jd := &protomsg.SC_Logined{}
	jd.Code = 1 //成功
	jd.Uid = (uid)
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))

	//通知进入场景
	d1 := &gamecore.UnitProperty{}
	d1.HP = 1000
	d1.MAX_HP = 1000
	d1.MAX_MP = 1000
	d1.MAX_MP = 1000
	d1.Name = "t1"
	d1.Level = 5
	d1.ModeType = "Hero/hero2"
	d1.Experience = 1000
	d1.Experience = 10000
	d1.ControlID = uid
	d1.BaseSpeed = 5
	t2 := protomsg.MsgUserEnterScene{
		Uid:            uid,
		ConnectId:      data.ConnectId,
		SrcServerName:  "",
		DestServerName: "GameScene1",
		SceneName:      "Map/set_5v5",
		Datas:          utils.Struct2Bytes(d1),
	}
	t1 := protomsg.MsgBase{
		ModeType: datamsg.GameScene1,
		MsgType:  "MsgUserEnterScene",
	}
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, &t2))

}

func (a *LoginAgent) NotifyGateLogined(conncetid int32, uid int32) {

	data := &protomsg.MsgBase{}
	data.Uid = (uid)
	data.ModeType = datamsg.GateMode
	data.MsgType = "UserLogin"
	data.ConnectId = (conncetid)

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, nil))

}

//重新登录 旧连接强制断开
func (a *LoginAgent) ReLoginForceDisconnect(conncetid int32, uid int32) {

	data := &protomsg.MsgBase{}
	data.Uid = (uid)
	data.ModeType = datamsg.GateMode
	data.MsgType = "ReLoginForceDisconnect"
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
