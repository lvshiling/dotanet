package gamescene1

import (
	"dq/conf"
	"dq/datamsg"
	//"dq/db"
	"dq/log"
	"dq/network"
	"fmt"
	"net"
	"time"

	//"dq/db"
	"dq/utils"
	//"dq/cyward"
	"dq/gamecore"
	"dq/protobuf"
	//"dq/timer"
	//"dq/vec2d"
	"dq/wordsfilter"
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
	a.handles["CS_DestroyItem"] = a.DoDestroyItem
	a.handles["CS_PlayerUpgradeSkill"] = a.DoPlayerUpgradeSkill
	a.handles["CS_ChangeAttackMode"] = a.DoChangeAttackMode

	a.handles["CS_LodingScene"] = a.DoLodingScene
	a.handles["CS_UseAI"] = a.DoUseAI

	a.handles["CS_LookVedioSucc"] = a.DoLookVedioSucc

	a.handles["CS_OrganizeTeam"] = a.DoOrganizeTeam
	a.handles["CS_ResponseOrgTeam"] = a.DoResponseOrgTeam
	a.handles["CS_OutTeam"] = a.DoOutTeam

	//商店
	a.handles["CS_GetStoreData"] = a.DoGetStoreData
	a.handles["CS_BuyCommodity"] = a.DoBuyCommodity
	//立即复活
	a.handles["CS_QuickRevive"] = a.DoQuickRevive

	//聊天信息
	a.handles["CS_ChatInfo"] = a.DoChatInfo

	//好友相关
	a.handles["CS_AddFriendRequest"] = a.DoAddFriendRequest
	a.handles["CS_RemoveFriend"] = a.DoRemoveFriend
	a.handles["CS_AddFriendResponse"] = a.DoAddFriendResponse
	a.handles["CS_GetFriendsList"] = a.DoGetFriendsList

	//邮件相关
	a.handles["CS_GetMailsList"] = a.DoGetMailsList
	a.handles["CS_GetMailInfo"] = a.DoGetMailInfo
	a.handles["CS_GetMailRewards"] = a.DoGetMailRewards

	//创建场景
	allscene := conf.GetAllScene()
	for _, v := range allscene {
		log.Info("scene:%d  %s", v.(*conf.SceneFileData).TypeID, v.(*conf.SceneFileData).ScenePath)
		if v.(*conf.SceneFileData).IsOpen != 1 {
			continue
		}
		log.Info("scene succ:%d ", v.(*conf.SceneFileData).TypeID)
		scene := gamecore.CreateScene(v.(*conf.SceneFileData), a)
		time.Sleep(time.Duration(33/len(allscene)) * time.Millisecond)
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

	log.Info("GameScene1Agent---------DoDisconnect")

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

func (a *GameScene1Agent) DoDestroyItem(data *protomsg.MsgBase) {

	log.Info("---------CS_DestroyItem")
	h2 := &protomsg.CS_DestroyItem{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	player.(*gamecore.Player).DestroyItem(h2)

	a.SendBagInfo(player.(*gamecore.Player))

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
			equip.Level = v.Level
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
			equip.Level = v.Level
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

//商店
func (a *GameScene1Agent) DoGetStoreData(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetStoreData{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).SendMsgToClient("SC_StoreData", conf.GetStoreData2SC_StoreData())

}
func (a *GameScene1Agent) DoBuyCommodity(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_BuyCommodity{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	//商品信息
	cominfo := conf.GetStoreFileData(h2.TypeID)
	player.(*gamecore.Player).BuyItem(cominfo)
}

//好友相关 请求把目标加为好友
func (a *GameScene1Agent) DoAddFriendRequest(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_AddFriendRequest{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyFriends == nil {
		return
	}

	friend := a.Players.Get(h2.Uid)
	if friend == nil {
		player.(*gamecore.Player).MyFriends.AddFriendRequest(h2, nil)
	} else {
		player.(*gamecore.Player).MyFriends.AddFriendRequest(h2, friend.(*gamecore.Player))
	}

}
func (a *GameScene1Agent) DoRemoveFriend(data *protomsg.MsgBase) {
	//	h2 := &protomsg.CS_RemoveFriend{}
	//	err := proto.Unmarshal(data.Datas, h2)
	//	if err != nil {
	//		log.Info(err.Error())
	//		return
	//	}
	//	player := a.Players.Get(data.Uid)
	//	if player == nil {
	//		return
	//	}

}
func (a *GameScene1Agent) CheckOnline(uid int32, characterid int32) bool {
	p1 := a.Players.Get(uid)
	if p1 != nil && p1.(*gamecore.Player).Characterid == characterid {
		return true
	}

	return false
}

//回复好友请求
func (a *GameScene1Agent) DoAddFriendResponse(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_AddFriendResponse{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyFriends == nil {
		return
	}

	friend := a.Players.Get(h2.FriendInfo.Uid)
	if friend == nil {
		player.(*gamecore.Player).MyFriends.AddFriendResponse(h2, nil)
	} else {
		player.(*gamecore.Player).MyFriends.AddFriendResponse(h2, friend.(*gamecore.Player))
	}

	//重新发送好友信息
	d1 := player.(*gamecore.Player).MyFriends.GetSCData()
	for k, v := range d1.Friends {
		if a.CheckOnline(v.Uid, v.Characterid) == true {
			d1.Friends[k].State = 1
		} else {
			d1.Friends[k].State = 2
		}

	}

	player.(*gamecore.Player).SendMsgToClient("SC_GetFriendsList", d1)

}
func (a *GameScene1Agent) DoGetFriendsList(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetFriendsList{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyFriends == nil {
		return
	}

	//*protomsg.SC_GetFriendsList
	d1 := player.(*gamecore.Player).MyFriends.GetSCData()
	for k, v := range d1.Friends {
		if a.CheckOnline(v.Uid, v.Characterid) == true {
			d1.Friends[k].State = 1
		} else {
			d1.Friends[k].State = 2
		}

	}
	player.(*gamecore.Player).SendMsgToClient("SC_GetFriendsList", d1)
}

//邮件信息
//邮件相关
//	a.handles["CS_GetMailsList"] = a.DoGetMailsList
//	a.handles["CS_GetMailInfo"] = a.DoGetMailInfo
//	a.handles["CS_GetMailRewards"] = a.DoGetMailRewards
func (a *GameScene1Agent) DoGetMailRewards(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetMailRewards{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyMails == nil {
		return
	}
	mail := player.(*gamecore.Player).MyMails.GetMailRewards(h2.Id)

	player.(*gamecore.Player).SendMsgToClient("SC_GetMailRewards", mail)
}
func (a *GameScene1Agent) DoGetMailInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetMailInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyMails == nil {
		return
	}
	mail := player.(*gamecore.Player).MyMails.GetOneMailInfo(h2.Id)

	player.(*gamecore.Player).SendMsgToClient("SC_GetMailInfo", mail)
}
func (a *GameScene1Agent) DoGetMailsList(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_GetMailsList{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil || player.(*gamecore.Player).MyMails == nil {
		return
	}
	mails := player.(*gamecore.Player).MyMails.GetMailsList()

	player.(*gamecore.Player).SendMsgToClient("SC_GetMailsList", mails)
}

//聊天信息
func (a *GameScene1Agent) DoChatInfo(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_ChatInfo{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	mainunit := player.(*gamecore.Player).MainUnit
	if mainunit == nil {
		return
	}

	//过滤非法字符
	h2.Content = wordsfilter.WF.DoReplace(h2.Content)

	////聊天频道 1附近 2全服 3私聊 4队伍
	if h2.Channel == 1 {
		if player.(*gamecore.Player).CurScene == nil {
			return
		}
		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.Content = h2.Content //内容过滤
		allplayer := player.(*gamecore.Player).CurScene.GetAllPlayerUseLock()
		for _, v := range allplayer {
			if v == nil {
				continue
			}
			v.SendMsgToClient("SC_ChatInfo", msg)
		}

	} else if h2.Channel == 2 {

	} else if h2.Channel == 3 {
		destplayer := a.Players.Get(h2.DestPlayerUID)
		if destplayer == nil {
			//未找到当前玩家
			return
		}

		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.DestPlayerUID = h2.DestPlayerUID
		msg.Content = h2.Content //内容过滤
		player.(*gamecore.Player).SendMsgToClient("SC_ChatInfo", msg)
		destplayer.(*gamecore.Player).SendMsgToClient("SC_ChatInfo", msg)

	} else if h2.Channel == 4 {
		team := gamecore.TeamManagerObj.GetTeam(player.(*gamecore.Player))
		if team == nil {
			return
		}
		msg := &protomsg.SC_ChatInfo{}
		msg.Channel = h2.Channel
		msg.Time = time.Now().Format("15:04")
		msg.SrcName = mainunit.Name
		msg.SrcPlayerUID = data.Uid
		msg.Content = h2.Content //内容过滤
		allplayer := team.Players.Items()
		for _, v := range allplayer {
			if v == nil {
				continue
			}
			v.(*gamecore.Player).SendMsgToClient("SC_ChatInfo", msg)
		}
	}

}

func (a *GameScene1Agent) DoQuickRevive(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_QuickRevive{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	if player.(*gamecore.Player).MainUnit != nil {
		player.(*gamecore.Player).MainUnit.QuickRevive = h2
	}
}

func (a *GameScene1Agent) DoLookVedioSucc(data *protomsg.MsgBase) {
	h2 := &protomsg.CS_LookVedioSucc{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}
	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}
	if player.(*gamecore.Player).MainUnit != nil {
		player.(*gamecore.Player).MainUnit.LookViewGetDiamond = h2
	}

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

//a.handles["CS_UseAI"] = a.DoUseAI
func (a *GameScene1Agent) DoUseAI(data *protomsg.MsgBase) {

	//log.Info("---------DoPlayerOperate")
	h2 := &protomsg.CS_UseAI{}
	err := proto.Unmarshal(data.Datas, h2)
	if err != nil {
		log.Info(err.Error())
		return
	}

	player := a.Players.Get(data.Uid)
	if player == nil {
		return
	}

	player.(*gamecore.Player).UseAI(h2.AIid)

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
