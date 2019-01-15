package gamescene1

import (
	//"dq/conf"
	"dq/datamsg"
	"dq/log"
	"dq/network"
	"net"
	//"time"

	//"dq/db"
	"dq/utils"
	//"dq/cyward"
	"dq/gamecore"
	"dq/protobuf"
	//"dq/timer"
	//"dq/vec2d"
	"sync"

	"github.com/golang/protobuf/proto"
)

//游戏部分
type GameScene1Agent struct {
	conn network.Conn

	handles map[string]func(data *protomsg.MsgBase)

	Scenes *utils.BeeMap

	wgScene sync.WaitGroup
}

func (a *GameScene1Agent) GetConnectId() int32 {

	return 0
}
func (a *GameScene1Agent) GetModeType() string {
	return ""
}

func (a *GameScene1Agent) Init() {

	a.Scenes = utils.NewBeeMap()

	//timer.AddRepeatCallback(time.Millisecond*time.Duration((100)), a.Update)

	a.handles = make(map[string]func(data *protomsg.MsgBase))
	a.handles["MsgUserEnterScene"] = a.DoMsgUserEnterScene
	a.handles["CS_SetTarget"] = a.DoCSSetTarget

	//创建场景
	for k := 0; k < 1; k++ {
		scene := gamecore.CreateScene("Map/set_5v5")
		a.Scenes.Set("Map/set_5v5", scene)
		a.wgScene.Add(1)
		go func() {
			scene.Update()
			a.wgScene.Done()
		}()
	}

	//玩家进来

}

func (a *GameScene1Agent) DoMsgUserEnterScene(data *protomsg.MsgBase) {

	log.Info("---------DoMsgUserEnterScene")
	h2 := &protomsg.MsgUserEnterScene{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//	a.testplayer = &Player{}
	//	a.testplayer.Uid = h2.Uid
	//	a.testplayer.ConnectId = h2.ConnectId
	//	a.testplayer.MainBody = a.core.CreateBody(vec2d.Vec2{-30, -30}, vec2d.Vec2{1, 1}, 10)

	//	log.Info("MainBody %v", a.testplayer.MainBody)
}
func (a *GameScene1Agent) DoCSSetTarget(data *protomsg.MsgBase) {

	log.Info("---------DoCSSetTarget")
	h2 := &protomsg.CS_SetTarget{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//	if a.testplayer != nil {
	//		a.testplayer.MainBody.SetTarget(vec2d.Vec2{float64(h2.TargetPos.X), float64(h2.TargetPos.Y)})
	//	}

}

func (a *GameScene1Agent) Update() {
	//	a.core.Update(0.1)

	//	a.SendMyData()
}

func (a *GameScene1Agent) SendMyData() {
	//	if a.testplayer != nil {
	//		//回复客户端
	//		data := &protomsg.MsgBase{}
	//		data.ModeType = "Client"
	//		data.Uid = a.testplayer.Uid
	//		data.ConnectId = a.testplayer.ConnectId
	//		data.MsgType = "SC_LogicFrame"
	//		jd := &protomsg.SC_LogicFrame{}
	//		//jd.Position = a.testplayer.MainBody.Position
	//		jd.Position = &protomsg.Point{}
	//		jd.Position.X = float32(a.testplayer.MainBody.Position.X)
	//		jd.Position.Y = float32(a.testplayer.MainBody.Position.Y)
	//		a.WriteMsgBytes(datamsg.NewMsg1Bytes(data, jd))
	//	}

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
	scenes := a.Scenes.Items()
	for _, v := range scenes {
		v.(*gamecore.Scene).Close()
	}

	a.wgScene.Wait()

	//存储玩家数据

	log.Info("GameScene1Agent OnClose")
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
