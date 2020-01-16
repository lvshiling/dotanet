package gamecore

import (
	"dq/conf"
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"encoding/json"
	"sort"
	"strconv"
	"sync"
)

//ItemType:10000表示金币 10001表示砖石  其他表示道具ID
type RewardsConfig struct {
	ItemType int32
	Count    int32
	Level    int32
}
type MailInfo struct {
	db.DB_MailInfo
	Reward []RewardsConfig
}

type Mails struct {
	MyPlayer    *Player       //载体
	MyMailsInfo *utils.BeeMap //邮件信息
	lock        *sync.RWMutex //同步操作锁
}

//获取邮件信息字符串
func (this *Mails) GetDBStr() string {
	mailsid := ""
	mailsitem := this.MyMailsInfo.Items()
	for _, v := range mailsitem {
		mailsid += ";" + strconv.Itoa(int(v.(*MailInfo).Id))
	}
	return mailsid
}

////删除已经领取附件的邮件(没有附件)
func (this *Mails) DeleteNoRewardMails() bool {
	all1 := this.MyMailsInfo.Items()
	for k, v := range all1 {
		//已经领取
		if v.(*MailInfo).Getstate == 1 {
			this.MyMailsInfo.Delete(k)
		}
	}
	return true
}

//获取邮件列表
func (this *Mails) GetMailsList() *protomsg.SC_GetMailsList {
	msg := &protomsg.SC_GetMailsList{}
	msg.Mails = make([]*protomsg.MailShortInfoMsg, 0)
	all1 := this.MyMailsInfo.Items()
	for _, v := range all1 {
		msim := &protomsg.MailShortInfoMsg{}
		msim.Id = v.(*MailInfo).Id
		msim.SendName = v.(*MailInfo).Sendname
		msim.Title = v.(*MailInfo).Title
		msim.State = v.(*MailInfo).Getstate
		msim.Date = v.(*MailInfo).Date
		msg.Mails = append(msg.Mails, msim)
	}
	msg.MailUpperLimit = int32(conf.Conf.NormalInfo.MailUpperLimit)
	return msg
}

//获取具体邮件信息
func (this *Mails) GetOneMailInfo(id int32) *protomsg.SC_GetMailInfo {
	msg := &protomsg.SC_GetMailInfo{}
	oneinfo := this.MyMailsInfo.Get(id)
	if oneinfo != nil {
		msg.Id = oneinfo.(*MailInfo).Id
		msg.SendName = oneinfo.(*MailInfo).Sendname
		msg.Title = oneinfo.(*MailInfo).Title
		msg.State = oneinfo.(*MailInfo).Getstate
		msg.Content = oneinfo.(*MailInfo).Content
		msg.Date = oneinfo.(*MailInfo).Date
		msg.Rewards = make([]*protomsg.MailRewards, 0)
		for _, v := range oneinfo.(*MailInfo).Reward {
			onereward := &protomsg.MailRewards{}
			onereward.ItemType = v.ItemType
			onereward.Count = v.Count
			onereward.Level = v.Level
			msg.Rewards = append(msg.Rewards, onereward)
		}
	}
	return msg
}

//领取邮件奖励
func (this *Mails) GetMailRewards(id int32) *protomsg.SC_GetMailRewards {

	this.lock.Lock()
	defer this.lock.Unlock()

	if this.MyPlayer == nil {
		return nil
	}

	msg := &protomsg.SC_GetMailRewards{}
	msg.Id = id
	msg.Result = 0 //1表示成功 0表示失败
	oneinfo := this.MyMailsInfo.Get(id)
	if oneinfo != nil {

		if oneinfo.(*MailInfo).Getstate == 1 {
			//已经被领取
			return msg
		}

		if this.MyPlayer.AddItemS2Bag(oneinfo.(*MailInfo).Reward) == false {
			return msg
		}

		msg.Result = 1
		//已经被领取
		oneinfo.(*MailInfo).Getstate = 1
		db.DbOne.SaveMail(oneinfo.(*MailInfo).DB_MailInfo)
	}

	return msg
}

//购买商品邮件
func Create_UnShelfCommodityMail_Mail(itemid int32, level int32) *MailInfo {
	mi := &MailInfo{}
	mi.Sendname = "系统"
	mi.Title = "返还道具"
	mi.Content = "你在交易所上架的该道具现已下架返还给你:"

	mi.Reward = make([]RewardsConfig, 0)
	tt := RewardsConfig{ItemType: itemid, Count: 1, Level: level}
	mi.Reward = append(mi.Reward, tt)
	rewards, _ := json.Marshal(mi.Reward)
	mi.Rewardstr = string(rewards)
	mi.Getstate = 0
	return mi
}
func (this *Mails) UnShelfCommodityMail(itemid int32, level int32) {

	if this.MyPlayer == nil {
		return
	}
	mi := Create_UnShelfCommodityMail_Mail(itemid, level)
	mi.RecUid = this.MyPlayer.Uid
	mi.RecCharacterid = this.MyPlayer.Characterid

	db.DbOne.CreateAndSaveMail(&mi.DB_MailInfo)
	this.MyMailsInfo.Set(mi.Id, mi)
}

//购买商品邮件
func Create_BuyCommodityMail_Mail(itemid int32, level int32) *MailInfo {
	mi := &MailInfo{}
	mi.Sendname = "系统"
	mi.Title = "购买道具"
	mi.Content = "恭喜你购得以下道具:"

	mi.Reward = make([]RewardsConfig, 0)
	tt := RewardsConfig{ItemType: itemid, Count: 1, Level: level}
	mi.Reward = append(mi.Reward, tt)
	rewards, _ := json.Marshal(mi.Reward)
	mi.Rewardstr = string(rewards)
	mi.Getstate = 0
	return mi
}
func (this *Mails) BuyCommodityMail(itemid int32, level int32) {

	if this.MyPlayer == nil {
		return
	}
	mi := Create_BuyCommodityMail_Mail(itemid, level)
	mi.RecUid = this.MyPlayer.Uid
	mi.RecCharacterid = this.MyPlayer.Characterid

	db.DbOne.CreateAndSaveMail(&mi.DB_MailInfo)
	this.MyMailsInfo.Set(mi.Id, mi)
}

//购买商品邮件
func Create_SellCommodityMail_Mail(pricetype int32, price int32) *MailInfo {
	mi := &MailInfo{}
	mi.Sendname = "系统"
	mi.Title = "售出道具"
	mi.Content = "恭喜你通过售卖道具获得以下收入:"
	mi.Reward = make([]RewardsConfig, 0)
	tt := RewardsConfig{ItemType: pricetype, Count: price, Level: 1}
	mi.Reward = append(mi.Reward, tt)
	rewards, _ := json.Marshal(mi.Reward)
	mi.Rewardstr = string(rewards)
	mi.Getstate = 0
	return mi
}

//售卖商品邮件
func (this *Mails) SellCommodityMail(pricetype int32, price int32) {

	if this.MyPlayer == nil {
		return
	}
	mi := Create_SellCommodityMail_Mail(pricetype, price)
	mi.RecUid = this.MyPlayer.Uid
	mi.RecCharacterid = this.MyPlayer.Characterid
	db.DbOne.CreateAndSaveMail(&mi.DB_MailInfo)

	this.MyMailsInfo.Set(mi.Id, mi)
}

////售卖道具获利邮件
//func SellItemMail(mailinfo MailInfo) {
//	db.DbOne.CreateAndSaveMail(&mailinfo)
//}

//创建邮件系统
func NewMails(mialstr string, myplayer *Player) *Mails {
	log.Info("NewMails:%s ", mialstr)

	mails := &Mails{}
	mails.MyMailsInfo = utils.NewBeeMap()
	mails.MyPlayer = myplayer
	mails.lock = new(sync.RWMutex)

	//mails.TestMail()

	//解析所有邮件
	allmialsid := utils.GetIntFromString3(mialstr, ";")
	//删除超过保存邮件上限的已经邮件
	maxcount := (conf.Conf.NormalInfo.MailUpperLimit)
	if len(allmialsid) > maxcount {
		sort.Ints(allmialsid)
		allmialsid = allmialsid[len(allmialsid)-maxcount:]
	}

	mialsinfo := make([]db.DB_MailInfo, 0)
	db.DbOne.GetMailsInfoByids(allmialsid, &mialsinfo)
	for _, v := range mialsinfo {
		mi := &MailInfo{}
		mi.DB_MailInfo = v

		mi.Reward = []RewardsConfig{}
		err := json.Unmarshal([]byte(v.Rewardstr), &mi.Reward)
		if err != nil {
			log.Info("NewMails Reward err:%s ", err.Error())
		}
		mails.MyMailsInfo.Set(v.Id, mi)
	}

	return mails
}
