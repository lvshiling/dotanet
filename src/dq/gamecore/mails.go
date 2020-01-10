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

//测试邮件
func (this *Mails) TestMail() {

	if this.MyPlayer == nil {
		return
	}

	mi := &MailInfo{}
	mi.Sendname = "系统"
	mi.Title = "测试"
	mi.Content = "恭喜你获得以下道具:"
	mi.RecUid = this.MyPlayer.Uid
	mi.RecCharacterid = this.MyPlayer.Characterid
	mi.Reward = make([]RewardsConfig, 0)
	tt := RewardsConfig{ItemType: 10, Count: 1, Level: 1}
	mi.Reward = append(mi.Reward, tt)
	tt1 := RewardsConfig{ItemType: 10000, Count: 1000, Level: 1}
	mi.Reward = append(mi.Reward, tt1)
	tt2 := RewardsConfig{ItemType: 10001, Count: 100, Level: 1}
	mi.Reward = append(mi.Reward, tt2)
	rewards, _ := json.Marshal(mi.Reward)
	mi.Rewardstr = string(rewards)
	mi.Getstate = 0

	db.DbOne.CreateAndSaveMail(&mi.DB_MailInfo)

	this.MyMailsInfo.Set(mi.Id, mi)
}

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
