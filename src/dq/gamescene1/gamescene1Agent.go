package gamescene1

import (
	"fmt"
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

	ServerName string
	Scenes     *utils.BeeMap
	Players    *utils.BeeMap

	wgScene sync.WaitGroup
}

func (a *GameScene1Agent) GetConnectId() int32 {

	return 0
}
func (a *GameScene1Agent) GetModeType() string {
	return ""
}

func (a *GameScene1Agent) Init() {

	a.ServerName = datamsg.GameScene1

	a.Scenes = utils.NewBeeMap()
	a.Players = utils.NewBeeMap()

	a.handles = make(map[string]func(data *protomsg.MsgBase))
	a.handles["MsgUserEnterScene"] = a.DoMsgUserEnterScene
	a.handles["Disconnect"] = a.DoDisconnect

	a.handles["CS_PlayerMove"] = a.DoPlayerMove

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

//
func (a *GameScene1Agent) DoDisconnect(data *protomsg.MsgBase) {

	log.Info("---------DoDisconnect")

	player := a.Players.Get(data.Uid)
	if player != nil {
		//退出之前的场景
		if player.(*gamecore.Player).ConnectId == data.ConnectId {

			log.Info("---------DoDisconnect--delete")
			player.(*gamecore.Player).OutScene()
			a.Players.Delete(data.Uid)
		}

	}

	//LoginOut
	t1 := protomsg.MsgBase{
		ModeType:  datamsg.LoginMode,
		MsgType:   "LoginOut",
		Uid:       data.Uid,
		ConnectId: data.ConnectId,
	}
	a.WriteMsgBytes(datamsg.NewMsg1Bytes(&t1, nil))

}

func (a *GameScene1Agent) DoMsgUserEnterScene(data *protomsg.MsgBase) {

	log.Info("---------DoMsgUserEnterScene")
	h2 := &protomsg.MsgUserEnterScene{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	log.Info("---------datas:%d---%s", len(h2.Datas), string(h2.Datas))

	//如果目的地服务器是本服务器
	if h2.DestServerName == a.ServerName {

		scene := a.Scenes.Get(h2.SceneName)
		if scene == nil {
			return
		}

		player := a.Players.Get(h2.Uid)
		if player == nil {
			player = gamecore.CreatePlayer(h2.Uid, h2.ConnectId)
			player.(*gamecore.Player).ServerAgent = a
			a.Players.Set(player.(*gamecore.Player).Uid, player)
		} else {
			//			//重新连接
			//			if player.(*gamecore.Player).ConnectId != h2.ConnectId {
			//				player.(*gamecore.Player).ConnectId = h2.ConnectId
			//				player.(*gamecore.Player).ClearShowData()
			//			}

		}

		//退出之前的场景
		player.(*gamecore.Player).OutScene()

		//进入新场景
		player.(*gamecore.Player).GoInScene(scene.(*gamecore.Scene), h2.Datas)

		//发送场景信息给玩家
		msg := &protomsg.SC_NewScene{}
		msg.Name = scene.(*gamecore.Scene).SceneName
		msg.LogicFps = int32(scene.(*gamecore.Scene).SceneFrame)
		msg.CurFrame = scene.(*gamecore.Scene).CurFrame
		msg.ServerName = a.ServerName
		player.(*gamecore.Player).SendMsgToClient("SC_NewScene", msg)

		log.Info("SendMsgToClient SC_NewScene")

	}

}

func (a *GameScene1Agent) DoPlayerMove(data *protomsg.MsgBase) {

	log.Info("---------DoPlayerOperate")
	h2 := &protomsg.CS_PlayerMove{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("---------%v", h2)

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).MoveCmd(h2)

}

//
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

	log.Debug("GameScene1Agent OnClose")
	fmt.Print("-----------")
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
		ModeType: a.ServerName,
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
