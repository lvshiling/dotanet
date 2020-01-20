package gamecore

import (
	"dq/log"
	"dq/protobuf"
	//"dq/timer"
	"dq/db"
	"dq/utils"
	"sync"
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
	Characters *utils.BeeMap //公会成员
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

		this.Guilds.Set(v.Id, guild)
	}

}

func (this *GuildManager) Close() {
	//存入数据库
	log.Info("GuildManager save")
}
