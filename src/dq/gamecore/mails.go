package gamecore

import (
	"dq/conf"
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"encoding/json"
	"strconv"
)

//ItemType:10000表示金币 10001表示砖石  其他表示道具ID
type RewardsConfig struct {
	ItemType int32
	Count    int32
}
type MailInfo struct {
	db.DB_MailInfo
	Reward []RewardsConfig
}

type Mails struct {
	MyPlayer    *Player       //载体
	MyMailsInfo *utils.BeeMap //好友
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
		msg.Rewards = make([]*protomsg.MailRewards, 0)
		for _, v := range oneinfo.(*MailInfo).Reward {
			onereward := &protomsg.MailRewards{}
			onereward.ItemType = v.ItemType
			onereward.Count = v.Count
			msg.Rewards = append(msg.Rewards, onereward)
		}
	}
	return msg
}

//领取邮件奖励
func (this *Mails) GetMailRewards(id int32) *protomsg.SC_GetMailRewards {

	if this.MyPlayer == nil {
		return nil
	}

	msg := &protomsg.SC_GetMailRewards{}
	msg.Id = id
	msg.Result = 0 //1表示成功 0表示失败
	oneinfo := this.MyMailsInfo.Get(id)
	if oneinfo != nil {
		itemcount := int32(0)
		for _, v := range oneinfo.(*MailInfo).Reward {
			if conf.IsBagItem(v.ItemType) {
				itemcount++
			}
		}
		//检查背包空位是否足够
		if this.MyPlayer.GetBagNilCount() < itemcount {
			return msg
		}
		msg.Result = 1
		for _, v := range oneinfo.(*MailInfo).Reward {
			this.MyPlayer.AddItem2Bag(v.ItemType, v.Count)
		}
		//已经被领取
		oneinfo.(*MailInfo).Getstate = 1
		db.DbOne.SaveMail(oneinfo.(*MailInfo).DB_MailInfo)
	}

	return msg
}

//创建邮件系统
func NewMails(mials string, myplayer *Player) *Mails {
	log.Info("NewMails:%s ", mials)

	mails := &Mails{}
	mails.MyMailsInfo = utils.NewBeeMap()
	mails.MyPlayer = myplayer

	//解析所有邮件
	allmialsid := utils.GetInt32FromString3(mials, ";")
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
