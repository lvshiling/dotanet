package gamescene1

import (
	"dq/conf"
	"dq/datamsg"
	"dq/log"
	"dq/network"
	"net"
	"time"

	//"dq/db"
	//"dq/utils"
	"dq/cyward"
	"dq/protobuf"
	"dq/timer"
	"dq/vec2d"

	"github.com/golang/protobuf/proto"
)

//游戏部分
type GameScene1Agent struct {
	conn network.Conn

	userdata string

	handles map[string]func(data *protomsg.MsgBase)

	core *cyward.WardCore
}

func (a *GameScene1Agent) GetConnectId() int32 {

	return 0
}
func (a *GameScene1Agent) GetModeType() string {
	return ""
}

func (a *GameScene1Agent) Init() {

	scenedata := conf.GetSceneData("Map/set_5v5")

	for _, v := range scenedata.Collides {
		log.Info("Collide %v", v)
		//		if v.IsRect == true{
		//			log.Info("IsRect= true %v",v)
		//		}else{

		//		}
	}

	a.core = cyward.CreateWardCore()

	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			pos := vec2d.Vec2{float64(0 + i*40), float64(0 + j*40)}
			r := vec2d.Vec2{float64(10 + i/2), float64(10 + j/2)}
			t := a.core.CreateBody(pos, r, 100.0)
			t.SetTag(i*10 + j)

		}

	}

	timer.AddRepeatCallback(time.Millisecond*time.Duration((20)), a.Update)

	a.handles = make(map[string]func(data *protomsg.MsgBase))

	//玩家进来

}
func (a *GameScene1Agent) Update() {
	a.core.Update(0.02)
}

func (a *GameScene1Agent) Run() {

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

func (a *GameScene1Agent) doMessage(data []byte) {
	//log.Info("----------game5g----readmsg---------")
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

func (a *GameScene1Agent) OnClose() {

}

func (a *GameScene1Agent) WriteMsg(msg interface{}) {

}
func (a *GameScene1Agent) WriteMsgBytes(msg []byte) {

	err := a.conn.WriteMsg(msg)
	if err != nil {
		log.Error("write message  error: %v", err)
	}
}
func (a *GameScene1Agent) RegisterToGate() {
	t2 := protomsg.MsgRegisterToGate{
		ModeType: datamsg.GameScene1,
	}

	t1 := protomsg.MsgBase{
		ModeType: datamsg.GateMode,
		MsgType:  "Register",
	}

	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, &t2))

}

func (a *GameScene1Agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *GameScene1Agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *GameScene1Agent) Close() {
	a.conn.Close()
}

func (a *GameScene1Agent) Destroy() {
	a.conn.Destroy()
}
