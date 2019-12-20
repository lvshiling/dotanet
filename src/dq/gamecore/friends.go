package gamecore

import (
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"strconv"
)

type FriendInfo struct {
	protomsg.FriendInfoMsg
	//	Uid         int32
	//	Characterid int32 //角色id
	//	Name        string
	//	Level       int32
	//	Typeid      int32
	//	State       int32 //1在线 2离线
}

type Friends struct {
	MyPlayer         *Player       //载体
	MyFriendsInfo    *utils.BeeMap //好友
	MyFriendsRequest *utils.BeeMap //好友请求
}

//被请求加为好友
func (this *Friends) AddFriendRequestData(id1 int32, data *FriendInfo) {
	this.MyFriendsRequest.Set(id1, data)
}

//被请求加为好友
func (this *Friends) AddFriendData(id1 int32, data *FriendInfo) {
	this.MyFriendsInfo.Set(id1, data)
}

//被请求加为好友
func (this *Friends) AddFriendRequest(h2 *protomsg.CS_AddFriendRequest, friend *Player) {
	if this.MyPlayer == nil {
		return
	}
	//如果角色ID为0 则把当前玩家正在玩的角色置为角色
	if friend != nil {
		if h2.Characterid <= 0 {
			h2.Characterid = friend.Characterid
		}
	}
	//该角色在线
	if friend != nil && friend.MyFriends != nil &&
		friend.Characterid == h2.Characterid {
		friendinfo := &FriendInfo{}
		friendinfo.Uid = this.MyPlayer.Uid
		friendinfo.Characterid = this.MyPlayer.Characterid
		if this.MyPlayer.MainUnit != nil {
			friendinfo.Name = this.MyPlayer.MainUnit.Name
			friendinfo.Level = this.MyPlayer.MainUnit.Level
			friendinfo.Typeid = this.MyPlayer.MainUnit.TypeID
			friendinfo.State = 1
		}

		friend.MyFriends.AddFriendRequestData(friendinfo.Characterid, friendinfo)

	} else {
		//该角色不在线
		db.DbOne.AddFriendsRequest(h2.Characterid, this.MyPlayer.Characterid)
	}

}

//回复好友请求//Result 1同意  2拒绝
func (this *Friends) AddFriendResponse(msg *protomsg.CS_AddFriendResponse, player *Player) {
	finfo := this.MyFriendsRequest.Get(msg.FriendInfo.Characterid)
	if finfo == nil || msg == nil {
		return
	}
	//删除好友请求
	this.MyFriendsRequest.Delete(msg.FriendInfo.Characterid)
	if msg.Result != 1 {
		return
	}

	if this.MyPlayer == nil {
		return
	}

	//同意加为好友
	friendinfo := &FriendInfo{}
	friendinfo.FriendInfoMsg = *msg.FriendInfo
	//自己加对面为好友
	this.MyFriendsInfo.Set(msg.FriendInfo.Characterid, friendinfo)
	//对面加自己为好友
	if player == nil || player.Characterid != msg.FriendInfo.Characterid || player.MyFriends == nil {
		//离线情况
		db.DbOne.AddFriends(msg.FriendInfo.Characterid, this.MyPlayer.Characterid)
	} else {
		//
		friendinfo1 := &FriendInfo{}
		friendinfo1.Uid = this.MyPlayer.Uid
		friendinfo1.Characterid = this.MyPlayer.Characterid
		if this.MyPlayer.MainUnit != nil {
			friendinfo1.Name = this.MyPlayer.MainUnit.Name
			friendinfo1.Level = this.MyPlayer.MainUnit.Level
			friendinfo1.Typeid = this.MyPlayer.MainUnit.TypeID
			friendinfo1.State = 1
		}
		player.MyFriends.AddFriendData(friendinfo1.Characterid, friendinfo1)
	}

}

//获取好友信息字符串
func (this *Friends) GetDBStr() (string, string) {
	friends := ""
	items := this.MyFriendsInfo.Items()
	for _, v := range items {
		friends += ";" + strconv.Itoa(int(v.(*FriendInfo).Characterid))
	}
	friendsrequest := ""
	item1s := this.MyFriendsRequest.Items()
	for _, v := range item1s {
		friendsrequest += ";" + strconv.Itoa(int(v.(*FriendInfo).Characterid))
	}
	return friends, friendsrequest
}

//返回传给客户端的数据
func (this *Friends) GetSCData() *protomsg.SC_GetFriendsList {
	msg := &protomsg.SC_GetFriendsList{}
	msg.Friends = make([]*protomsg.FriendInfoMsg, 0)
	all1 := this.MyFriendsInfo.Items()
	for _, v := range all1 {
		msg.Friends = append(msg.Friends, &(v.(*FriendInfo).FriendInfoMsg))
	}

	msg.FriendsRequest = make([]*protomsg.FriendInfoMsg, 0)
	all2 := this.MyFriendsRequest.Items()
	for _, v := range all2 {
		msg.FriendsRequest = append(msg.FriendsRequest, &(v.(*FriendInfo).FriendInfoMsg))
	}

	return msg
}

//创建好友
func NewFriends(friendsid string, friendsrequest string, myplayer *Player) *Friends {
	log.Info("NewFriends:%s   %s", friendsid, friendsrequest)

	friends := &Friends{}
	friends.MyFriendsInfo = utils.NewBeeMap()
	friends.MyPlayer = myplayer
	//解析所有好友
	allfriendid := utils.GetInt32FromString3(friendsid, ";")
	players := make([]db.DB_CharacterInfo, 0)
	db.DbOne.GetCharactersInfoByCharacterids(allfriendid, &players)
	for _, v := range players {
		friendinfo := &FriendInfo{}
		friendinfo.Uid = v.Uid
		friendinfo.Characterid = v.Characterid
		friendinfo.Name = v.Name
		friendinfo.Level = v.Level
		friendinfo.Typeid = v.Typeid
		friendinfo.State = 2
		friends.MyFriendsInfo.Set(friendinfo.Characterid, friendinfo)
	}
	//解析好友请求
	friends.MyFriendsRequest = utils.NewBeeMap()
	allfriendidrequest := utils.GetInt32FromString3(friendsrequest, ";")
	playersrequest := make([]db.DB_CharacterInfo, 0)
	db.DbOne.GetCharactersInfoByCharacterids(allfriendidrequest, &playersrequest)
	for _, v := range playersrequest {
		friendinfo := &FriendInfo{}
		friendinfo.Uid = v.Uid
		friendinfo.Characterid = v.Characterid
		friendinfo.Name = v.Name
		friendinfo.Level = v.Level
		friendinfo.Typeid = v.Typeid
		friendinfo.State = 2
		friends.MyFriendsRequest.Set(friendinfo.Characterid, friendinfo)
	}

	return friends
}
