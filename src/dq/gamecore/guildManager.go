package gamecore

import (
	"dq/log"
	"dq/protobuf"
	//"dq/timer"
	"dq/conf"
	"dq/db"
	"dq/utils"
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
	guild.(*GuildInfo).Characters.Set(characterinfo.Characterid, guildchainfo)
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
	Characters            *utils.BeeMap //公会成员
	RequestJoinCharacters *utils.BeeMap //请求加入公会角色
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
		if v.(*db.DB_GuildInfo).Name == name {
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
	guildBaseInfo.Name = guild.Name
	guildBaseInfo.Level = guild.Level
	guildBaseInfo.Experience = guild.Experience
	guildBaseInfo.MaxExperience = int32(10000)
	guildBaseInfo.CharacterCount = int32(guild.Characters.Size())
	guildBaseInfo.MaxCount = int32(100)
	guildBaseInfo.PresidentName = ""
	president := guild.Characters.Get(guild.PresidentCharacterid)
	if president != nil {
		guildBaseInfo.PresidentName = president.(*protomsg.GuildChaInfo).Name
	}
	guildBaseInfo.Joinaudit = guild.Joinaudit
	guildBaseInfo.Joinlevellimit = guild.Joinlevellimit
	return guildBaseInfo
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
	chaitems := guild.Characters.Items()
	for _, v := range chaitems {
		one := &v.(*GuildCharacterInfo).GuildChaInfo
		guildinfo.Characters = append(guildinfo.Characters, one)
	}

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
	guild.Notice = "欢迎来到(" + name + ")大家庭!"
	guild.Joinaudit = 0
	guild.Joinlevellimit = 1
	guild.Characters = utils.NewBeeMap()
	guild.RequestJoinCharacters = utils.NewBeeMap()
	//数据库创建信息获得ID
	_, id := db.DbOne.CreateGuild(name)
	if id < 0 {
		return nil
	}
	guild.Id = id
	//把公会加入列表
	this.Guilds.Set(guild.Id, guild)

	return guild

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
		guild.Characters = utils.NewBeeMap()
		guild.RequestJoinCharacters = utils.NewBeeMap()

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
			guild.Characters.Set(guildchainfo.Characterid, guildchainfo)
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

			guild.RequestJoinCharacters.Set(guildchainfo.Characterid, guildchainfo)
		}

		this.Guilds.Set(v.Id, guild)
	}

}

func (this *GuildManager) Close() {
	//存入数据库
	log.Info("GuildManager save")
}
