package gamescene1

import (
	"dq/conf"
	"dq/datamsg"
	//"dq/db"
	"dq/log"
	"dq/network"
	"fmt"
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

	//初始化 组队信息
	gamecore.TeamManagerObj.Init(a)

	a.ServerName = datamsg.GameScene1

	a.Scenes = utils.NewBeeMap()
	a.Players = utils.NewBeeMap()

	a.handles = make(map[string]func(data *protomsg.MsgBase))
	a.handles["MsgUserEnterScene"] = a.DoMsgUserEnterScene
	a.handles["Disconnect"] = a.DoDisconnect

	a.handles["CS_PlayerMove"] = a.DoPlayerMove

	a.handles["CS_PlayerAttack"] = a.DoPlayerAttack
	a.handles["CS_PlayerSkill"] = a.DoPlayerSkill

	a.handles["CS_GetUnitInfo"] = a.DoGetUnitInfo
	a.handles["CS_GetBagInfo"] = a.DoGetBagInfo
	a.handles["CS_ChangeItemPos"] = a.DoChangeItemPos
	a.handles["CS_PlayerUpgradeSkill"] = a.DoPlayerUpgradeSkill
	a.handles["CS_ChangeAttackMode"] = a.DoChangeAttackMode

	a.handles["CS_LodingScene"] = a.DoLodingScene

	a.handles["CS_OrganizeTeam"] = a.DoOrganizeTeam
	a.handles["CS_ResponseOrgTeam"] = a.DoResponseOrgTeam
	a.handles["CS_OutTeam"] = a.DoOutTeam

	//创建场景
	allscene := conf.GetAllScene()
	for _, v := range allscene {
		log.Info("scene:%d  %s", v.(*conf.SceneFileData).TypeID, v.(*conf.SceneFileData).ScenePath)
		scene := gamecore.CreateScene(v.(*conf.SceneFileData), a)
		a.Scenes.Set(v.(*conf.SceneFileData).TypeID, scene)
		a.wgScene.Add(1)
		go func() {
			scene.Update()
			a.wgScene.Done()
		}()
	}

}

//
func (a *GameScene1Agent) DoDisconnect(data *protomsg.MsgBase) {

	log.Info("---------DoDisconnect")

	player := a.Players.Get(data.Uid)
	if player != nil {
		//退出之前的场景
		if player.(*gamecore.Player).ConnectId == data.ConnectId {

			log.Info("---------DoDisconnect--delete")

			gamecore.TeamManagerObj.LeaveTeam(player.(*gamecore.Player))

			player.(*gamecore.Player).SaveDB()

			player.(*gamecore.Player).OutScene()
			a.Players.Delete(data.Uid)
			//存档 数据库

		} else {
			log.Info("---------DoDisconnect--ConnectId fail")
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
func (a *GameScene1Agent) PlayerChangeScene(player *gamecore.Player, doorway conf.DoorWay) {
	if player == nil {
		return
	}
	dbdata := player.GetDBData()
	if dbdata == nil {
		return
	}
	dbdata.X = doorway.NextX
	dbdata.Y = doorway.NextY
	//h2 := &protomsg.MsgUserEnterScene{}

	h2 := &protomsg.MsgUserEnterScene{
		Uid:            player.Uid,
		ConnectId:      player.ConnectId,
		SrcServerName:  "",
		DestServerName: datamsg.GameScene1, //
		SceneID:        doorway.NextSceneID,
		Datas:          utils.Struct2Bytes(dbdata), //数据库中的角色信息
	}
	a.DoUserEnterScene(h2)
}
func (a *GameScene1Agent) GetPlayerByID(uid int32) *gamecore.Player {
	player := a.Players.Get(uid)
	if player == nil {
		return nil
	}
	return player.(*gamecore.Player)
}

func (a *GameScene1Agent) DoUserEnterScene(h2 *protomsg.MsgUserEnterScene) {
	if h2 == nil {
		return
	}

	//	characterinfo := db.DB_CharacterInfo{}
	//	utils.Bytes2Struct(h2.Datas, &characterinfo)
	//	log.Info("---------datas:%v", characterinfo)

	//如果目的地服务器是本服务器
	if h2.DestServerName == a.ServerName {

		scene := a.Scenes.Get(h2.SceneID)
		log.Info("enter scene :%d", h2.SceneID)
		if scene == nil {
			log.Info("no scene :%d", h2.SceneID)
			return
		}

		player := a.Players.Get(h2.Uid)
		if player == nil {
			player = gamecore.CreatePlayer(h2.Uid, h2.ConnectId, -1)
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
		msg.SceneID = scene.(*gamecore.Scene).TypeID
		player.(*gamecore.Player).SendMsgToClient("SC_NewScene", msg)

		log.Info("SendMsgToClient SC_NewScene")

	}

}

func (a *GameScene1Agent) DoMsgUserEnterScene(data *protomsg.MsgBase) {

	log.Info("---------DoMsgUserEnterScene")
	h2 := &protomsg.MsgUserEnterScene{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	a.DoUserEnterScene(h2)

}

//切换攻击模式
func (a *GameScene1Agent) DoChangeAttackMode(data *protomsg.MsgBase) {

	log.Info("---------DoChangeAttackMode")
	h2 := &protomsg.CS_ChangeAttackMode{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).ChangeAttackMode(h2)

}

//升级技能
//DoPlayerUpgradeSkill
func (a *GameScene1Agent) DoPlayerUpgradeSkill(data *protomsg.MsgBase) {

	log.Info("---------DoPlayerUpgradeSkill")
	h2 := &protomsg.CS_PlayerUpgradeSkill{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).UpgradeSkill(h2)

}

//DoChangeItemPos
func (a *GameScene1Agent) DoChangeItemPos(data *protomsg.MsgBase) {

	log.Info("---------DoChangeItemPos")
	h2 := &protomsg.CS_ChangeItemPos{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).ChangeItemPos(h2)

	a.SendUnitInfo(player.(*gamecore.Player).MainUnit, player.(*gamecore.Player))
	a.SendBagInfo(player.(*gamecore.Player))

}

func (a *GameScene1Agent) SendBagInfo(player *gamecore.Player) {
	if player == nil {
		return
	}
	msg := &protomsg.SC_BagInfo{}
	msg.Equips = make([]*protomsg.UnitEquip, 0)
	for _, v := range player.BagInfo {
		if v != nil {
			equip := &protomsg.UnitEquip{}
			equip.Pos = v.Index
			equip.TypdID = v.TypeID
			msg.Equips = append(msg.Equips, equip)
		}
	}

	player.SendMsgToClient("SC_BagInfo", msg)
}

//DoGetBagInfo
func (a *GameScene1Agent) DoGetBagInfo(data *protomsg.MsgBase) {

	log.Info("---------DoGetBagInfo")
	h2 := &protomsg.CS_GetBagInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("---------%d", h2.UnitID)
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	a.SendBagInfo(player.(*gamecore.Player))

}

func (a *GameScene1Agent) SendUnitInfo(unit *gamecore.Unit, player *gamecore.Player) {
	unitdata := &protomsg.UnitBoardDatas{}
	unitdata.ID = unit.ID
	unitdata.Name = unit.Name
	unitdata.AttributeStrength = unit.AttributeStrength
	unitdata.AttributeAgility = unit.AttributeAgility
	unitdata.AttributeIntelligence = unit.AttributeIntelligence
	unitdata.Attack = unit.Attack
	unitdata.AttackSpeed = unit.AttackSpeed
	unitdata.AttackRange = unit.AttackRange
	unitdata.MoveSpeed = float32(unit.MoveSpeed)
	unitdata.MagicScale = unit.MagicScale
	unitdata.MPRegain = unit.MPRegain
	unitdata.PhysicalAmaor = unit.PhysicalAmaor
	unitdata.PhysicalResist = unit.PhysicalResist
	unitdata.MagicAmaor = unit.MagicAmaor
	unitdata.StatusAmaor = unit.StatusAmaor
	unitdata.Dodge = unit.Dodge
	unitdata.HPRegain = unit.HPRegain
	unitdata.AttributePrimary = int32(unit.AttributePrimary)
	unitdata.DropItems = unit.NPCItemDropInfo
	unitdata.RemainExperience = unit.RemainExperience
	//道具栏
	unitdata.Equips = make([]*protomsg.UnitEquip, 0)
	for k, v := range unit.Items {
		equip := &protomsg.UnitEquip{}
		equip.Pos = int32(k)
		if v != nil {
			equip.TypdID = v.TypeID
		} else {
			equip.TypdID = 0
		}
		unitdata.Equips = append(unitdata.Equips, equip)
	}

	msg := &protomsg.SC_UnitInfo{}
	msg.UnitData = unitdata
	player.SendMsgToClient("SC_UnitInfo", msg)
}

//DoGetUnitInfo
func (a *GameScene1Agent) DoGetUnitInfo(data *protomsg.MsgBase) {

	log.Info("---------DoGetUnitInfo")
	h2 := &protomsg.CS_GetUnitInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("---------%d", h2.UnitID)
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	curscene := player.(*gamecore.Player).CurScene
	if curscene == nil {
		return
	}
	unit := curscene.FindUnitByID(h2.UnitID)
	if unit == nil {
		return
	}
	a.SendUnitInfo(unit, player.(*gamecore.Player))
}

//DoPlayerSkill
func (a *GameScene1Agent) DoPlayerSkill(data *protomsg.MsgBase) {

	log.Info("---------DoPlayerSkill")
	h2 := &protomsg.CS_PlayerSkill{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	log.Info("---------%v  %f  %f", h2, h2.X, h2.Y)

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).SkillCmd(h2)

}

func (a *GameScene1Agent) DoPlayerAttack(data *protomsg.MsgBase) {

	log.Info("---------DoPlayerAttack")
	h2 := &protomsg.CS_PlayerAttack{}
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

	player.(*gamecore.Player).AttackCmd(h2)

}

func (a *GameScene1Agent) DoOrganizeTeam(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_OrganizeTeam{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player1 := a.Players.Get(h2.Player1)
	player2 := a.Players.Get(h2.Player2)
	if player1 == nil || player2 == nil {
		return
	}
	if h2.Player1 == h2.Player2 {
		return
	}

	gamecore.TeamManagerObj.OrganizeTeam(player1.(*gamecore.Player), player2.(*gamecore.Player))
}
func (a *GameScene1Agent) DoResponseOrgTeam(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_ResponseOrgTeam{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	gamecore.TeamManagerObj.ResponseOrgTeam(h2, player.(*gamecore.Player))
}

//a.handles["CS_OutTeam"] = a.DoOutTeam
func (a *GameScene1Agent) DoOutTeam(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_OutTeam{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player1 := a.Players.Get(h2.OutPlayerUID)
	if player1 == nil {
		return
	}
	if player == player1 {
		gamecore.TeamManagerObj.LeaveTeam(player1.(*gamecore.Player))
	} else {
		gamecore.TeamManagerObj.OutTeam(player.(*gamecore.Player), player1.(*gamecore.Player))
	}

}

//CS_LodingScene
func (a *GameScene1Agent) DoLodingScene(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerOperate")

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).IsLoadedSceneSucc = true

}

func (a *GameScene1Agent) DoPlayerMove(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerOperate")
	h2 := &protomsg.CS_PlayerMove{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	//log.Info("---------%v", h2)

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
	gamecore.TeamManagerObj.Close()

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
