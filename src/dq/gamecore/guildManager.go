package gamecore

import (
	"dq/log"
	"dq/protobuf"
	//"dq/timer"
	"dq/conf"
	"dq/db"
	"dq/utils"
	"strconv"
	"sync"
	"time"
)

var (
	GuildManagerObj = &GuildManager{}
)

//type ServerInterface interface {
//	GetPlayerByID(id int32) *Player
//}

//公会成员信息
type GuildCharacterInfo struct {
	protomsg.GuildChaInfo
	GuildId int32 //公会ID

}

func NewGuildCharacterInfo(characterinfo *db.DB_CharacterInfo) *GuildCharacterInfo {

	guild := GuildManagerObj.Guilds.Get(characterinfo.GuildId)
	if guild == nil {
		return nil
	}

	guildchainfo := &GuildCharacterInfo{}
	//重新设置公会成员信息
	guild.(*GuildInfo).CharactersMap.Set(characterinfo.Characterid, guildchainfo)
	guildchainfo.Uid = characterinfo.Uid
	guildchainfo.Characterid = characterinfo.Characterid
	guildchainfo.GuildId = characterinfo.GuildId
	guildchainfo.Name = characterinfo.Name
	guildchainfo.Level = characterinfo.Level
	guildchainfo.Typeid = characterinfo.Typeid
	guildchainfo.PinLevel = characterinfo.GuildPinLevel
	guildchainfo.PinExperience = characterinfo.GuildPinExperience
	guildchainfo.Post = characterinfo.GuildPost

	return guildchainfo
}

//公会信息
type GuildInfo struct {
	db.DB_GuildInfo
	CharactersMap            *utils.BeeMap //公会成员
	RequestJoinCharactersMap *utils.BeeMap //请求加入公会角色
}

//公会管理器
type GuildManager struct {
	Guilds      *utils.BeeMap //当前服务器组队信息
	OperateLock *sync.RWMutex //同步操作锁
	Server      ServerInterface
}

//初始化
func (this *GuildManager) Init(server ServerInterface) {
	log.Info("----------GuildManager Init---------")
	this.Guilds = utils.NewBeeMap()
	this.Server = server
	this.OperateLock = new(sync.RWMutex)

	this.LoadDataFromDB()

}

//检查是否有同名的公会存在
func (this *GuildManager) CheckName(name string) bool {
	items := this.Guilds.Items()
	for _, v := range items {
		//重名了
		if v.(*GuildInfo).Name == name {
			return true
		}
	}
	return false

}

//获取所有公会简短信息
func (this GuildManager) GetAllGuildsInfo() *protomsg.SC_GetAllGuildsInfo {
	protoallguilds := &protomsg.SC_GetAllGuildsInfo{}
	protoallguilds.CreatePriceType = int32(conf.Conf.NormalInfo.CreateGuildPriceType)
	protoallguilds.CreatePrice = int32(conf.Conf.NormalInfo.CreateGuildPrice)
	protoallguilds.Guilds = make([]*protomsg.GuildShortInfo, 0)
	allguilds := this.Guilds.Items()
	for _, v := range allguilds {
		one := this.GuildInfo2ProtoGuildShortInfo(v.(*GuildInfo))
		protoallguilds.Guilds = append(protoallguilds.Guilds, one)
	}
	return protoallguilds
}

//本公会信息转成proto的公会简短信息
func (this *GuildManager) GuildInfo2ProtoGuildShortInfo(guild *GuildInfo) *protomsg.GuildShortInfo {
	guildBaseInfo := &protomsg.GuildShortInfo{}
	guildBaseInfo.ID = guild.Id
	guildBaseInfo.Name = guild.Name
	guildBaseInfo.Level = guild.Level
	guildBaseInfo.Experience = guild.Experience
	guildBaseInfo.MaxExperience = int32(10000)
	guildBaseInfo.CharacterCount = int32(guild.CharactersMap.Size())
	guildBaseInfo.MaxCount = int32(100)
	guildBaseInfo.PresidentName = ""
	guildBaseInfo.Notice = guild.Notice
	president := guild.CharactersMap.Get(guild.PresidentCharacterid)
	if president != nil {
		guildBaseInfo.PresidentName = president.(*GuildCharacterInfo).Name
	}
	guildBaseInfo.Joinaudit = guild.Joinaudit
	guildBaseInfo.Joinlevellimit = guild.Joinlevellimit
	return guildBaseInfo
}

//获取申请列表信息
func (this *GuildManager) GetJoinGuildPlayer(id int32) *protomsg.SC_GetJoinGuildPlayer {
	guildinfo := &protomsg.SC_GetJoinGuildPlayer{}
	guild1 := this.Guilds.Get(id)
	if guild1 == nil {
		return nil
	}
	guild := guild1.(*GuildInfo)

	//公会申请成员信息
	guildinfo.RequestCharacters = make([]*protomsg.GuildChaInfo, 0)
	rechaitems := guild.RequestJoinCharactersMap.Items()
	for _, v := range rechaitems {
		one := &v.(*GuildCharacterInfo).GuildChaInfo
		guildinfo.RequestCharacters = append(guildinfo.RequestCharacters, one)
	}

	return guildinfo
}

//获取公会信息
func (this *GuildManager) GetGuildInfo(id int32) *protomsg.SC_GetGuildInfo {
	guildinfo := &protomsg.SC_GetGuildInfo{}
	guild1 := this.Guilds.Get(id)
	if guild1 == nil {
		return nil
	}
	guild := guild1.(*GuildInfo)

	//公会信息
	guildinfo.GuildBaseInfo = this.GuildInfo2ProtoGuildShortInfo(guild)
	//公会成员信息
	guildinfo.Characters = make([]*protomsg.GuildChaInfo, 0)
	chaitems := guild.CharactersMap.Items()
	for _, v := range chaitems {
		one := &v.(*GuildCharacterInfo).GuildChaInfo
		guildinfo.Characters = append(guildinfo.Characters, one)
	}
	//	//公会申请成员信息
	//	guildinfo.RequestCharacters = make([]*protomsg.GuildChaInfo, 0)
	//	rechaitems := guild.RequestJoinCharactersMap.Items()
	//	for _, v := range rechaitems {
	//		one := &v.(*GuildCharacterInfo).GuildChaInfo
	//		guildinfo.RequestCharacters = append(guildinfo.RequestCharacters, one)
	//	}

	return guildinfo
}

//创建公会
func (this *GuildManager) CreateGuild(name string) *GuildInfo {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	if this.CheckName(name) == true {
		return nil
	}

	guild := &GuildInfo{}
	guild.Createday = time.Now().Format("2006-01-02")
	guild.Name = name
	guild.Level = 1
	guild.Notice = "欢迎来到(" + name + ")大家庭!"
	guild.Joinaudit = 0
	guild.Joinlevellimit = 1
	guild.CharactersMap = utils.NewBeeMap()
	guild.RequestJoinCharactersMap = utils.NewBeeMap()
	//数据库创建信息获得ID
	_, id := db.DbOne.CreateGuild(name, guild.Createday)
	if id < 0 {
		return nil
	}
	guild.Id = id
	//把公会加入列表
	this.Guilds.Set(guild.Id, guild)

	return guild

}

//踢人
func (this *GuildManager) DeleteGuildPlayer(player *Player, data *protomsg.CS_DeleteGuildPlayer, targetplayer *Player) {
	if player == nil || data == nil || player.MyGuild == nil || player == targetplayer {
		return
	}
	//回复玩家加入公会
	postdata := conf.GetGuildPostFileData(player.MyGuild.Post)
	if postdata == nil || postdata.DeletePlayerWriteAble != 1 {
		//没有权限 31
		player.SendNoticeWordToClient(31)
		return
	}
	//找到当前公会
	guild1 := this.Guilds.Get(player.MyGuild.GuildId)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return
	}
	guild := guild1.(*GuildInfo)

	//申请的玩家角色
	character := targetplayer
	if character == nil { //离线
		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterid(data.Characterid, &players)
		if len(players) <= 0 {
			//找不到该用户
			return
		}
		//存档数据库
		players[0].GuildId = 0
		players[0].GuildPinLevel = int32(1)
		players[0].GuildPinExperience = int32(0)
		players[0].GuildPost = int32(1)
		db.DbOne.SaveCharacter(players[0])

	} else { //在线
		if character.MyGuild != nil {
			//对方已经有公会了 32
			character.MyGuild = nil
		}
	}

	//把数据存入公会中
	guild.CharactersMap.Delete(data.Characterid)
}

//回复加入公会的申请
func (this *GuildManager) ResponseJoinGuild(player *Player, data *protomsg.CS_ResponseJoinGuildPlayer, targetplayer *Player) {
	if player == nil || data == nil || player.MyGuild == nil {
		return
	}
	//回复玩家加入公会
	postdata := conf.GetGuildPostFileData(player.MyGuild.Post)
	if postdata == nil || postdata.ResponseJoinPlayerWriteAble != 1 {
		//没有权限 31
		player.SendNoticeWordToClient(31)
		return
	}
	//找到当前公会
	guild1 := this.Guilds.Get(player.MyGuild.GuildId)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return
	}
	guild := guild1.(*GuildInfo)

	//删除申请请求
	chainfo := guild.RequestJoinCharactersMap.Get(data.Characterid)
	if chainfo == nil {
		//找不到该玩家
		return
	}

	guild.RequestJoinCharactersMap.Delete(data.Characterid)

	//如果是拒绝就到此为止
	if data.Result != 1 {
		return
	}

	//申请的玩家角色
	character := targetplayer
	if character == nil { //离线
		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterid(data.Characterid, &players)
		if len(players) <= 0 {
			//找不到该用户
			return
		}
		if players[0].GuildId > 0 {
			//对方已经有公会了 32
			player.SendNoticeWordToClient(32)
			return
		}

		//存档数据库
		players[0].GuildId = player.MyGuild.GuildId
		players[0].GuildPinLevel = int32(1)
		players[0].GuildPinExperience = int32(0)
		players[0].GuildPost = int32(1)
		db.DbOne.SaveCharacter(players[0])
		//把数据存入公会中
		guildchainfo := &GuildCharacterInfo{}
		guildchainfo.Uid = players[0].Uid
		guildchainfo.Characterid = players[0].Characterid
		guildchainfo.Name = players[0].Name
		guildchainfo.Level = players[0].Level
		guildchainfo.Typeid = players[0].Typeid
		guildchainfo.GuildId = players[0].GuildId
		guildchainfo.PinLevel = players[0].GuildPinLevel
		guildchainfo.PinExperience = players[0].GuildPinExperience
		guildchainfo.Post = players[0].GuildPost
		guild.CharactersMap.Set(guildchainfo.Characterid, guildchainfo)

	} else { //在线
		if character.MyGuild != nil {
			//对方已经有公会了 32
			player.SendNoticeWordToClient(32)
		} else {
			//角色创建的公会信息
			character.NewAddGuildInfo(player.MyGuild.GuildId, 1)
		}
	}

	//message CS_ResponseJoinGuildPlayer{
	//    int32 Characterid = 1;
	//    int32 Result = 2; //1表示同意 其他表示不同意
	//}
}

//申请加入公会
func (this *GuildManager) RequestJoinGuild(player *Player, guildid int32) {
	if player == nil || guildid <= 0 {
		return
	}
	if player.MyGuild != nil {
		//已经有公会了 不能加入
		player.SendNoticeWordToClient(28)
		return
	}
	guild1 := this.Guilds.Get(guildid)
	if guild1 == nil {
		//不存在该公会
		player.SendNoticeWordToClient(29)
		return
	}
	guild := guild1.(*GuildInfo)

	//角色信息
	guildchainfo := &GuildCharacterInfo{}
	guildchainfo.Uid = player.Uid
	guildchainfo.Characterid = player.Characterid
	if player.MainUnit != nil {
		guildchainfo.Name = player.MainUnit.Name
		guildchainfo.Level = player.MainUnit.Level
		guildchainfo.Typeid = player.MainUnit.TypeID
	}

	guild.RequestJoinCharactersMap.Set(player.Characterid, guildchainfo)

	player.SendNoticeWordToClient(30)

	log.Info("----------RequestJoinGuild--")
}

//从数据库载入数据
func (this *GuildManager) LoadDataFromDB() {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	commoditys := make([]db.DB_GuildInfo, 0)
	db.DbOne.GetGuilds(&commoditys)
	for _, v := range commoditys {
		//log.Info("----------ExchangeManager load %d %v", v.Id, &commoditys[k])
		guild := &GuildInfo{}
		guild.DB_GuildInfo = v
		guild.CharactersMap = utils.NewBeeMap()
		guild.RequestJoinCharactersMap = utils.NewBeeMap()

		//解析公会成员数据
		allguildids := utils.GetInt32FromString3(v.Characters, ";")
		players := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterids(allguildids, &players)
		for _, v1 := range players {
			guildchainfo := &GuildCharacterInfo{}
			guildchainfo.Uid = v1.Uid
			guildchainfo.Characterid = v1.Characterid
			guildchainfo.Name = v1.Name
			guildchainfo.Level = v1.Level
			guildchainfo.Typeid = v1.Typeid
			guildchainfo.GuildId = v1.GuildId
			guildchainfo.PinLevel = v1.GuildPinLevel
			guildchainfo.PinExperience = v1.GuildPinExperience
			guildchainfo.Post = v1.GuildPost
			guild.CharactersMap.Set(guildchainfo.Characterid, guildchainfo)
		}
		//解析请求加入公会的角色

		requestallguildids := utils.GetInt32FromString3(v.RequestJoinCharacters, ";")
		requestplayers := make([]db.DB_CharacterInfo, 0)
		db.DbOne.GetCharactersInfoByCharacterids(requestallguildids, &requestplayers)
		for _, v1 := range requestplayers {
			guildchainfo := &GuildCharacterInfo{}
			guildchainfo.Uid = v1.Uid
			guildchainfo.Characterid = v1.Characterid
			guildchainfo.Name = v1.Name
			guildchainfo.Level = v1.Level
			guildchainfo.Typeid = v1.Typeid

			guild.RequestJoinCharactersMap.Set(guildchainfo.Characterid, guildchainfo)
		}

		this.Guilds.Set(v.Id, guild)
	}

}

//type GuildInfo struct {
//	db.DB_GuildInfo
//	CharactersMap            *utils.BeeMap //公会成员
//	RequestJoinCharactersMap *utils.BeeMap //请求加入公会角色
//}
func (this *GuildManager) SaveDBGuildInfo(guild *GuildInfo) {
	if guild == nil {
		return
	}
	chaitems := guild.CharactersMap.Items()
	guild.Characters = ""
	for _, item := range chaitems {
		guild.Characters += strconv.Itoa(int(item.(*GuildCharacterInfo).Characterid)) + ";"
	}

	joinitems := guild.RequestJoinCharactersMap.Items()
	guild.RequestJoinCharacters = ""
	for _, item := range joinitems {
		guild.RequestJoinCharacters += strconv.Itoa(int(item.(*GuildCharacterInfo).Characterid)) + ";"
	}

	db.DbOne.SaveGuild(guild.DB_GuildInfo)
}

func (this *GuildManager) Close() {
	//存入数据库
	log.Info("GuildManager save")
	guilditems := this.Guilds.Items()
	for _, guild := range guilditems {
		this.SaveDBGuildInfo(guild.(*GuildInfo))
	}
}
