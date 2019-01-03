package hall

import (
	"dq/datamsg"
	"dq/log"
	"dq/network"
	"net"
	//"strconv"
	"dq/protobuf"
	"dq/utils"

	"github.com/golang/protobuf/proto"
)

type ScoreAndTime struct {
	Time  int64
	Score int
}

type HallAgent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *protomsg.MsgBase)

	//
	closeFlag *utils.BeeVar
}

func (a *HallAgent) GetConnectId() int32 {

	return 0
}
func (a *HallAgent) GetModeType() string {
	return ""
}

func (a *HallAgent) Init() {

	a.closeFlag = utils.NewBeeVar(false)

	a.handles = make(map[string]func(data *protomsg.MsgBase))

	//玩家断线
	a.handles["Disconnect"] = a.DoDisConnectData

}

func (a *HallAgent) SendMsgToAllClient(msgtype string, jd proto.Message) {

	data := &protomsg.MsgBase{}

	data.ModeType = "Client"
	data.MsgType = msgtype
	data.Uid = -2
	data.ConnectId = -2

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
}

func (a *HallAgent) DoDisConnectData(data *protomsg.MsgBase) {

	log.Info("----DoDisConnectData uid:%d--", data.Uid)

}

func (a *HallAgent) Update() {

}

func (a *HallAgent) Run() {

	a.Init()

	go a.Update()

	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Debug("read message: %v", err)
			break
		}

		go a.doMessage(data)

	}
}

func (a *HallAgent) doMessage(data []byte) {
	//log.Info("----------Hall----readmsg---------")
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

func (a *HallAgent) OnClose() {

}

func (a *HallAgent) WriteMsg(msg interface{}) {

}
func (a *HallAgent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}
func (a *HallAgent) RegisterToGate() {
	t2 := protomsg.MsgRegisterToGate{
		ModeType: datamsg.HallMode,
	}

	t1 := protomsg.MsgBase{
		ModeType: datamsg.GateMode,
		MsgType:  "Register",
	}

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, &t2))

}

func (a *HallAgent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *HallAgent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *HallAgent) Close() {
	a.closeFlag.Set(true)
	a.conn.Close()
}

func (a *HallAgent) Destroy() {
	a.conn.Destroy()
}
