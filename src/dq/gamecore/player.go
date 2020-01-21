package gamecore

import (
	"dq/conf"
	"dq/datamsg"
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/wordsfilter"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
)

var MaxBagCount int32 = 25

type Server interface {
	WriteMsgBytes(msg []byte)
}

var AutoSaveTime float32 = 10

//场景里的道具
type BagItem struct {
	TypeID int32 //类型
	Index  int32 //位置索引
	Level  int32 //等级
}

//道具技能CD
type ItemSkillCDData struct {
	TypeID       int32   //类型
	RemainCDTime float32 //剩余CD时间
}

type Player struct {
	Uid               int32
	ConnectId         int32
	Characterid       int32         //角色id
	MainUnit          *Unit         //主单位
	OtherUnit         *utils.BeeMap //其他单位
	CurScene          *Scene
	ServerAgent       Server
	IsLoadedSceneSucc bool //是否loading成功

	TeamID int32 //组队信息

	MyFriends *Friends            //好友
	MyMails   *Mails              //邮件系统
	MyGuild   *GuildCharacterInfo //公会系统

	BagInfo []*BagItem

	ItemSkillCDDataInfo map[int32]*ItemSkillCDData

	lock *sync.RWMutex //同步操作锁

	Buy *sync.RWMutex //同步操作锁

	AutoSaveRemainTime float32 //自动保存 剩余时间

	//OtherUnit  *Unit //其他单位

	//组合数据包相关
	LastShowUnit   map[int32]*Unit
	CurShowUnit    map[int32]*Unit
	LastShowBullet map[int32]*Bullet
	CurShowBullet  map[int32]*Bullet
	LastShowHalo   map[int32]*Halo
	CurShowHalo    map[int32]*Halo

	LastShowSceneItem map[int32]*SceneItem
	CurShowSceneItem  map[int32]*SceneItem
	Msg               *protomsg.SC_Update
}

func CreatePlayer(uid int32, connectid int32, characterid int32) *Player {
	re := &Player{}
	re.lock = new(sync.RWMutex)
	re.Buy = new(sync.RWMutex)
	re.Uid = uid
	re.ConnectId = connectid
	re.Characterid = characterid
	re.AutoSaveRemainTime = AutoSaveTime
	re.ReInit()
	return re
}

//获取道具技能CD信息
func (this *Player) GetItemSkillCDInfo(typeid int32) *ItemSkillCDData {
	if val, ok := this.ItemSkillCDDataInfo[typeid]; ok {
		return val
	}
	return nil
}

//保存道具技能CD信息
func (this *Player) SaveItemSkillCDInfo(skill *Skill) {
	if skill == nil {
		return
	}

	//log.Info("---SaveItemSkillCDInfo-%d  %f", skill.TypeID, skill.RemainCDTime)

	bagitem := &ItemSkillCDData{}
	bagitem.TypeID = skill.TypeID
	if skill.RemainSkillCount >= 1 {
		bagitem.RemainCDTime = 0
	} else {
		bagitem.RemainCDTime = skill.RemainCDTime
	}
	this.ItemSkillCDDataInfo[bagitem.TypeID] = bagitem

}

//载入道具技能CD信息
func (this *Player) LoadItemSkillCDFromDB(itemskillcd string) {
	if len(itemskillcd) <= 0 {
		return
	}
	itemskillcds := strings.Split(itemskillcd, ";")
	for _, v := range itemskillcds {
		itemskill := utils.GetFloat32FromString3(v, ",")
		if len(itemskill) < 2 {
			continue
		}

		bagitem := &ItemSkillCDData{}
		bagitem.TypeID = int32(itemskill[0])
		bagitem.RemainCDTime = itemskill[1]
		this.ItemSkillCDDataInfo[bagitem.TypeID] = bagitem

	}
}

//载入背包信息 从数据库数据
func (this *Player) LoadBagInfoFromDB(baginfo string) {
	if len(baginfo) <= 0 {
		return
	}
	bagitems := strings.Split(baginfo, ";")
	for _, v := range bagitems {
		item := utils.GetInt32FromString3(v, ",")
		if len(item) < 2 {
			continue
		}
		if item[0] >= MaxBagCount {
			return
		}
		bagitem := &BagItem{}
		bagitem.Index = item[0]
		bagitem.TypeID = item[1]
		if len(item) >= 3 { //等级
			bagitem.Level = item[2]
		} else {
			bagitem.Level = 1
		}
		this.BagInfo[bagitem.Index] = bagitem

	}
}

func (this *Player) ReInit() {
	this.MainUnit = nil
	this.LastShowUnit = make(map[int32]*Unit)
	this.CurShowUnit = make(map[int32]*Unit)
	this.LastShowBullet = make(map[int32]*Bullet)
	this.CurShowBullet = make(map[int32]*Bullet)
	this.LastShowHalo = make(map[int32]*Halo)
	this.CurShowHalo = make(map[int32]*Halo)
	this.LastShowSceneItem = make(map[int32]*SceneItem)
	this.CurShowSceneItem = make(map[int32]*SceneItem)

	this.BagInfo = make([]*BagItem, MaxBagCount)

	this.ItemSkillCDDataInfo = make(map[int32]*ItemSkillCDData)

	//
	this.OtherUnit = utils.NewBeeMap()
	this.Msg = &protomsg.SC_Update{}

	this.IsLoadedSceneSucc = false
}

//使用AI
func (this *Player) UseAI(id int32) {
	if this.MainUnit == nil {
		return
	}
	this.MainUnit.AttackMode = 3 //全体攻击模式
	this.MainUnit.SetAI(NewNormalAI(this.MainUnit))
}

//添加其他可控制的单位
func (this *Player) AddOtherUnit(unit *Unit) {
	if unit == nil {
		return
	}
	unit.ControlID = this.Uid
	this.OtherUnit.Set(unit.ID, unit)
}

//是否可以拾取地面的物品
//func (this *Player) CanSelectSceneItem() bool {
//	if this.MainUnit == nil || this.MainUnit.Items == nil {
//		return false
//	}
//	for _, v := range this.MainUnit.Items {
//		if v == nil {
//			return true
//		}
//	}
//	for _, v := range this.BagInfo {
//		if v == nil {
//			return true
//		}
//	}

//	return false
//}

func (this *Player) UpgradeSkill(data *protomsg.CS_PlayerUpgradeSkill) {
	if this.MainUnit == nil {
		return
	}

	this.MainUnit.UpgradeSkill(data)
}

func (this *Player) ChangeAttackMode(data *protomsg.CS_ChangeAttackMode) {
	if this.MainUnit == nil {
		return
	}
	this.MainUnit.ChangeAttackMode(data)

	this.CheckOtherUnit()
	items := this.OtherUnit.Items()
	for _, v1 := range items {
		v1.(*Unit).ChangeAttackMode(data)
	}
}

//交换道具位置 背包位置
func (this *Player) ChangeItemPos(data *protomsg.CS_ChangeItemPos) {
	//1表示装备栏 2表示背包
	if data.SrcType == 1 {
		if data.SrcPos < 0 || data.SrcPos >= UnitEquitCount {
			return
		}
	} else {
		if data.SrcPos < 0 || data.SrcPos >= MaxBagCount {
			return
		}
	}

	if data.DestType == 1 {
		if data.DestPos < 0 || data.DestPos >= UnitEquitCount {
			return
		}
	} else {
		if data.DestPos < 0 || data.DestPos >= MaxBagCount {
			return
		}
	}
	if this.MainUnit == nil {
		return
	}

	log.Info("-------ChangeItemPos:%d  %d  %d  %d", data.SrcPos, data.DestPos, data.SrcType, data.DestType)

	this.lock.Lock()
	//1表示装备栏 2表示背包
	if data.SrcType == 1 {
		src := this.MainUnit.Items[data.SrcPos]
		if data.DestType == 1 {
			dest := this.MainUnit.Items[data.DestPos]
			//只交换位置
			if dest != nil {
				dest.SetIndex(data.SrcPos)
			}

			this.MainUnit.Items[data.SrcPos] = dest
			if src != nil {
				src.SetIndex(data.DestPos)
			}

			this.MainUnit.Items[data.DestPos] = src
		} else {
			dest := this.BagInfo[data.DestPos]
			//删除角色身上的装备
			this.MainUnit.RemoveItem(data.SrcPos)
			if dest != nil {
				item := NewItem(dest.TypeID, dest.Level)
				this.MainUnit.AddItem(data.SrcPos, item)
			}
			//删除背包道具
			this.BagInfo[data.DestPos] = nil
			if src != nil {
				item := &BagItem{}
				item.Index = data.DestPos
				item.TypeID = src.TypeID
				item.Level = src.Level
				this.BagInfo[data.DestPos] = item
			}

		}

	} else {
		src := this.BagInfo[data.SrcPos]
		if data.DestType == 1 {
			dest := this.MainUnit.Items[data.DestPos]

			//删除角色身上的装备
			this.MainUnit.RemoveItem(data.DestPos)
			if src != nil {
				item := NewItem(src.TypeID, src.Level)
				this.MainUnit.AddItem(data.DestPos, item)
			}
			//删除背包道具
			this.BagInfo[data.SrcPos] = nil
			if dest != nil {
				item := &BagItem{}
				item.Index = data.SrcPos
				item.TypeID = dest.TypeID
				item.Level = dest.Level
				this.BagInfo[data.SrcPos] = item
			}
		} else {
			dest := this.BagInfo[data.DestPos]
			if src != nil && dest != nil && src.TypeID == dest.TypeID && src.Level == dest.Level {
				//相同道具且等级相同为合成
				maxlevel := int32(4)
				itemdata := conf.GetItemData(dest.TypeID)
				if itemdata != nil {
					maxlevel = itemdata.MaxLevel
				}
				if dest.Level >= maxlevel {
					//不能超过4级 21
					this.SendNoticeWordToClient(21)
				} else {
					//合成成功 22
					this.BagInfo[data.SrcPos] = nil
					this.BagInfo[data.DestPos].Level++
					this.SendNoticeWordToClient(22)
				}
			} else {
				//只交换位置
				this.BagInfo[data.SrcPos] = dest
				if this.BagInfo[data.SrcPos] != nil {
					this.BagInfo[data.SrcPos].Index = data.SrcPos
				}
				//src.Index = data.DestPos
				this.BagInfo[data.DestPos] = src
				if this.BagInfo[data.DestPos] != nil {
					this.BagInfo[data.DestPos].Index = data.DestPos
				}
			}

		}
	}

	this.lock.Unlock()
}

//交换道具位置 背包位置
func (this *Player) DestroyItem(data *protomsg.CS_DestroyItem) {

	if data.SrcPos < 0 || data.SrcPos >= MaxBagCount {
		return
	}
	if this.MainUnit == nil {
		return
	}

	log.Info("-------DestroyItem:%d ", data.SrcPos)

	this.lock.Lock()
	defer this.lock.Unlock()
	this.BagInfo[data.SrcPos] = nil

}

//获取道具
func (this *Player) AddItem(typeid int32, level int32) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.MainUnit == nil || this.MainUnit.Items == nil {
		return false
	}

	for _, v := range this.MainUnit.Items {
		if v == nil {

			item := NewItem(typeid, level)
			this.MainUnit.AddItem(-1, item)
			this.SendGetItemNotice(typeid, level)
			//this.lock.Unlock()
			return true
		}
	}
	for k, v := range this.BagInfo {
		if v == nil {
			item := &BagItem{}
			item.Index = int32(k)
			item.TypeID = typeid
			item.Level = level
			this.BagInfo[k] = item
			this.SendGetItemNotice(typeid, level)
			//this.lock.Unlock()
			return true
		}
	}
	//this.lock.Unlock()
	return false
}

//获取我要出售界面信息
func (this *Player) GetSellUIInfo() *protomsg.SC_GetSellUIInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	data := &protomsg.SC_GetSellUIInfo{}
	data.ShelfExchangeLimit = int32(conf.Conf.NormalInfo.ShelfExchangeLimit)
	data.SellExchangeTax = float32(conf.Conf.NormalInfo.SellExchangeTax)
	data.ShelfExchangeFeePriceType = int32(conf.Conf.NormalInfo.ShelfExchangeFeePriceType)
	data.ShelfExchangeFeePrice = int32(conf.Conf.NormalInfo.ShelfExchangeFeePrice)
	data.AutoUnShelfTime = int32(conf.Conf.NormalInfo.AutoUnShelfTime)
	data.Commoditys = ExchangeManagerObj.GetPlayerSelling(this.Characterid)
	data.Equips = make([]*protomsg.UnitEquip, 0)
	for _, v := range this.BagInfo {
		if v != nil {
			equip := &protomsg.UnitEquip{}
			equip.Pos = v.Index
			equip.TypdID = v.TypeID
			equip.Level = v.Level
			data.Equips = append(data.Equips, equip)
		}
	}

	return data
}

//上架道具到交易所
func (this *Player) ShelfBagItem2Exchange(pos int32) (bool, *BagItem) {
	this.lock.Lock()
	defer this.lock.Unlock()

	shouxufeiprice := int32(conf.Conf.NormalInfo.ShelfExchangeFeePrice)
	shouxufeipricetype := int32(conf.Conf.NormalInfo.ShelfExchangeFeePriceType)

	maxcount := int32(conf.Conf.NormalInfo.ShelfExchangeLimit)

	if pos < 0 || pos >= MaxBagCount {
		return false, nil
	}
	if this.MainUnit == nil {
		return false, nil
	}
	item := this.BagInfo[pos]
	if item == nil {
		return false, nil
	}
	//是否超过售卖上限
	if ExchangeManagerObj.GetPlayerSellingCount(this.Characterid) >= maxcount {
		//操作上限
		this.SendNoticeWordToClient(23)
		return false, nil
	}

	//扣手续费
	//手续费不足
	if this.BuyItemSubMoney(shouxufeipricetype, shouxufeiprice) == false {
		//货币不足
		this.SendNoticeWordToClient(shouxufeipricetype)
		return false, nil
	}

	//删除道具
	this.BagInfo[pos] = nil
	return true, item
}

//获取多个道具到背包[]RewardsConfig
func (this *Player) AddItemS2Bag(items []RewardsConfig) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	itemcount := int32(0)
	for _, v := range items {
		if conf.IsBagItem(v.ItemType) {
			itemcount++
		}
	}
	//检查背包空位是否足够
	if this.GetBagNilCount() < itemcount {
		this.SendNoticeWordToClient(7)
		return false
	}

	for _, v := range items {
		////:10000表示金币 10001表示砖石  其他表示道具ID
		if v.ItemType == 10000 {
			//扣钱
			this.MainUnit.Gold += v.Count
			continue
		} else if v.ItemType == 10001 {
			this.MainUnit.Diamond += v.Count
			continue
		}
		for k, v1 := range this.BagInfo {
			if v1 == nil {
				item := &BagItem{}
				item.Index = int32(k)
				item.TypeID = v.ItemType
				item.Level = v.Level
				this.BagInfo[k] = item
				this.SendGetItemNotice(v.ItemType, v.Level)
				break
			}
		}
	}

	return true
}

//发送获取道具提示
func (this *Player) SendGetItemNotice(typeid int32, level int32) {
	itemdata := conf.GetItemData(typeid)
	if itemdata == nil {
		return
	}
	this.SendNoticeWordToClient(25, itemdata.ItemName, "lv."+strconv.Itoa(int(level)))
}

//获取道具到背包
//func (this *Player) AddItem2Bag(typeid int32, count int32) bool {
//	this.lock.Lock()
//	defer this.lock.Unlock()
//	////:10000表示金币 10001表示砖石  其他表示道具ID
//	if typeid == 10000 {
//		//扣钱
//		this.MainUnit.Gold += count
//		return true
//	} else if typeid == 10001 {
//		this.MainUnit.Diamond += count
//		return true
//	}
//	for k, v := range this.BagInfo {
//		if v == nil {
//			item := &BagItem{}
//			item.Index = int32(k)
//			item.TypeID = typeid
//			this.BagInfo[k] = item
//			//
//			this.SendGetItemNotice(typeid,1)
//			return true
//		}
//	}
//	return false
//}

//创建公会
func (this *Player) CreateGuild(data *protomsg.CS_CreateGuild) {
	//this.MyGuild = NewGuildCharacterInfo(&characterinfo)
	if this.MyGuild != nil {
		//已经有公会了
		return
	}
	//含有非法字符
	if wordsfilter.WF.DoContains(data.Name) == true {
		this.SendNoticeWordToClient(26)
		return
	}
	//
	pricetype := int32(conf.Conf.NormalInfo.CreateGuildPriceType)
	price := int32(conf.Conf.NormalInfo.CreateGuildPrice)
	if this.BuyItemSubMoneyLock(pricetype, price) == false {
		//钱不够
		this.SendNoticeWordToClient(pricetype)
		return
	}
	guildinfo := GuildManagerObj.CreateGuild(data.Name)
	if guildinfo == nil {
		//重名
		this.SendNoticeWordToClient(27)
		//还回已经扣除的钱
		this.BuyItemSubMoneyLock(pricetype, 0-price)
		return
	}

	//创建公会成功 (补上创始人ID)
	guildinfo.PresidentCharacterid = this.Characterid
	//角色创建的公会信息
	characterinfo := db.DB_CharacterInfo{}
	characterinfo.GuildId = guildinfo.Id
	characterinfo.Uid = this.Uid
	characterinfo.Characterid = this.Characterid
	if this.MainUnit != nil {
		characterinfo.Name = this.MainUnit.Name
		characterinfo.Level = this.MainUnit.Level
		characterinfo.Typeid = this.MainUnit.TypeID
	}
	characterinfo.GuildPinLevel = int32(1)
	characterinfo.GuildPinExperience = int32(0)
	characterinfo.GuildPost = int32(10)
	this.MyGuild = NewGuildCharacterInfo(&characterinfo)
}

//获取背包空位
func (this *Player) GetBagNilCount() int32 {
	count := int32(0)
	for _, v := range this.BagInfo {
		if v == nil {
			count++
		}
	}
	return count
}

//拾取地面物品
func (this *Player) SelectSceneItem(sceneitem *SceneItem) bool {
	return this.AddItem(sceneitem.TypeID, 1)
}

//遍历删除无效的
func (this *Player) CheckOtherUnit() {

	items := this.OtherUnit.Items()
	for k, v := range items {
		if v == nil || v.(*Unit).IsDisappear() {
			this.OtherUnit.Delete(k)
		}
	}

}
func (this *Player) BuyItemSubMoney(pricetype int32, price int32) bool {

	//扣钱
	if pricetype == 10000 {
		if this.MainUnit.Gold >= price {
			this.MainUnit.Gold -= price
			return true
		}

	}
	if pricetype == 10001 {
		if this.MainUnit.Diamond >= price {
			this.MainUnit.Diamond -= price
			return true
		}
	}
	return false
}

//买东西扣钱
func (this *Player) BuyItemSubMoneyLock(pricetype int32, price int32) bool {
	this.Buy.Lock()
	defer this.Buy.Unlock()
	return this.BuyItemSubMoney(pricetype, price)
}

//购买道具
func (this *Player) BuyItem(cominfo *conf.CommodityData) bool {
	this.Buy.Lock()
	defer this.Buy.Unlock()
	if cominfo == nil {
		return false
	}
	if this.CheckPrice(cominfo.PriceType, cominfo.Price) == false {
		//货币不足
		this.SendNoticeWordToClient(cominfo.PriceType)
		return false
	}

	if this.AddItem(cominfo.ItemID, cominfo.Level) == false {
		//背包已满
		this.SendNoticeWordToClient(7)
		return false
	}
	//扣钱
	if this.BuyItemSubMoney(cominfo.PriceType, cominfo.Price) == false {
		//货币不足
		this.SendNoticeWordToClient(cominfo.PriceType)
		return false
	}

	//购买成功
	this.SendNoticeWordToClient(8)

	return true

}

//检查是否有足够的货币
func (this *Player) CheckPrice(pricetype int32, price int32) bool {
	if this.MainUnit == nil {
		return false
	}

	if pricetype == 10000 {
		if this.MainUnit.Gold >= price {
			return true
		} else {
			return false
		}
	}

	if pricetype == 10001 {
		if this.MainUnit.Diamond >= price {
			return true
		} else {
			return false
		}
	}

	return false
}

//type DB_CharacterInfo struct {
//	Characterid int32   `json:"characterid"`
//	Uid         int32   `json:"uid"`
//	Name        string  `json:"name"`
//	Typeid      int32   `json:"typeid"`
//	Level       int32   `json:"level"`
//	Experience  int32   `json:"experience"`
//	Gold        int32   `json:"gold"`
//	HP          float32 `json:"hp"`
//	MP          float32 `json:"mp"`
//	SceneName   string  `json:"scenename"`
//	X           float32 `json:"x"`
//	Y           float32 `json:"y"`
//}

func (this *Player) GetDBData() *db.DB_CharacterInfo {
	if this.MainUnit == nil {
		return nil
	}
	dbdata := db.DB_CharacterInfo{}
	dbdata.Characterid = this.Characterid
	dbdata.Uid = this.Uid
	dbdata.Name = this.MainUnit.Name
	dbdata.Typeid = this.MainUnit.TypeID
	dbdata.Level = this.MainUnit.Level
	dbdata.Experience = this.MainUnit.Experience
	dbdata.Gold = this.MainUnit.Gold
	dbdata.Diamond = this.MainUnit.Diamond
	dbdata.HP = float32(this.MainUnit.HP) / float32(this.MainUnit.MAX_HP)
	dbdata.MP = float32(this.MainUnit.MP) / float32(this.MainUnit.MAX_MP)
	dbdata.RemainExperience = this.MainUnit.RemainExperience
	dbdata.GetExperienceDay = this.MainUnit.GetExperienceDay
	dbdata.RemainReviveTime = this.MainUnit.RemainReviveTime
	//击杀相关
	dbdata.KillCount = this.MainUnit.KillCount
	dbdata.ContinuityKillCount = this.MainUnit.ContinuityKillCount
	dbdata.DieCount = this.MainUnit.DieCount
	dbdata.KillGetGold = this.MainUnit.KillGetGold
	dbdata.WatchVedioCountOneDay = this.MainUnit.WatchVedioCountOneDay

	if this.CurScene != nil {
		dbdata.SceneID = this.CurScene.TypeID
	} else {
		dbdata.SceneID = 1
	}
	if this.MainUnit.Body == nil {
		dbdata.X = 0
		dbdata.Y = 0
	} else {
		dbdata.X = float32(this.MainUnit.Body.Position.X)
		dbdata.Y = float32(this.MainUnit.Body.Position.Y)
	}

	//好友
	if this.MyFriends != nil {
		friends, friendsrequest := this.MyFriends.GetDBStr()
		dbdata.Friends = friends
		dbdata.FriendsRequest = friendsrequest
	}

	//邮件
	if this.MyMails != nil {
		mials := this.MyMails.GetDBStr()
		dbdata.Mails = mials
	}
	//公会数据
	if this.MyGuild != nil {
		dbdata.GuildId = this.MyGuild.GuildId
		dbdata.GuildPinLevel = this.MyGuild.PinLevel
		dbdata.GuildPinExperience = this.MyGuild.PinExperience
		dbdata.GuildPost = this.MyGuild.Post
	} else {
		dbdata.GuildId = 0
		dbdata.GuildPinLevel = 1
		dbdata.GuildPinExperience = 0
		dbdata.GuildPost = 0
	}

	//技能
	for _, v := range this.MainUnit.Skills {
		dbdata.Skill += v.ToDBString() + ";"
	}
	//道具
	if this.MainUnit.Items != nil {
		item1 := this.MainUnit.Items[0]
		if item1 == nil {
			dbdata.Item1 = ""
		} else {
			dbdata.Item1 = strconv.Itoa(int(item1.TypeID)) + "," + strconv.Itoa(int(item1.Level))
		}
		item2 := this.MainUnit.Items[1]
		if item2 == nil {
			dbdata.Item2 = ""
		} else {
			dbdata.Item2 = strconv.Itoa(int(item2.TypeID)) + "," + strconv.Itoa(int(item2.Level))
		}
		item3 := this.MainUnit.Items[2]
		if item3 == nil {
			dbdata.Item3 = ""
		} else {
			dbdata.Item3 = strconv.Itoa(int(item3.TypeID)) + "," + strconv.Itoa(int(item3.Level))
		}
		item4 := this.MainUnit.Items[3]
		if item4 == nil {
			dbdata.Item4 = ""
		} else {
			dbdata.Item4 = strconv.Itoa(int(item4.TypeID)) + "," + strconv.Itoa(int(item4.Level))
		}
		item5 := this.MainUnit.Items[4]
		if item5 == nil {
			dbdata.Item5 = ""
		} else {
			dbdata.Item5 = strconv.Itoa(int(item5.TypeID)) + "," + strconv.Itoa(int(item5.Level))
		}
		item6 := this.MainUnit.Items[5]
		if item6 == nil {
			dbdata.Item6 = ""
		} else {
			dbdata.Item6 = strconv.Itoa(int(item6.TypeID)) + "," + strconv.Itoa(int(item6.Level))
		}
	}

	//背包信息
	baginfo := ""
	for _, v := range this.BagInfo {
		if v != nil {
			itemstr := strconv.Itoa(int(v.Index)) + "," + strconv.Itoa(int(v.TypeID)) + "," + strconv.Itoa(int(v.Level)) + ";"
			baginfo += itemstr
		}
	}
	dbdata.BagInfo = baginfo

	//道具技能CD信息
	for _, v := range this.MainUnit.ItemSkills {
		this.SaveItemSkillCDInfo(v)
	}
	itemskillcdinfo := ""
	for _, v := range this.ItemSkillCDDataInfo {
		if v != nil {
			itemstr := strconv.Itoa(int(v.TypeID)) + "," + strconv.FormatFloat(float64(v.RemainCDTime), 'f', 4, 32) + ";"
			itemskillcdinfo += itemstr
		}
	}
	dbdata.ItemSkillCDInfo = itemskillcdinfo

	return &dbdata
}

//存档数据
func (this *Player) SaveDB() {

	//if this.MainUnit != nil

	dbdata := this.GetDBData()
	if dbdata == nil {
		return
	}
	db.DbOne.SaveCharacter(*dbdata)

}

//清除显示状态  切换场景的时候需要调用
func (this *Player) ClearShowData() {
	this.LastShowUnit = make(map[int32]*Unit)
	this.CurShowUnit = make(map[int32]*Unit)
	this.LastShowBullet = make(map[int32]*Bullet)
	this.CurShowBullet = make(map[int32]*Bullet)
	this.LastShowHalo = make(map[int32]*Halo)
	this.CurShowHalo = make(map[int32]*Halo)

	this.LastShowSceneItem = make(map[int32]*SceneItem)
	this.CurShowSceneItem = make(map[int32]*SceneItem)
	//

	this.Msg = &protomsg.SC_Update{}
}

//添加客户端显示单位数据包
func (this *Player) AddUnitData(unit *Unit) {

	if this.MainUnit == nil {
		return
	}
	//检查玩家主单位是否能看见 目标单位
	if this.MainUnit.CanSeeTarget(unit) == false {
		return
	}

	this.CurShowUnit[unit.ID] = unit

	if _, ok := this.LastShowUnit[unit.ID]; ok {
		//旧单位(只更新变化的值)
		d1 := *unit.ClientDataSub
		this.Msg.OldUnits = append(this.Msg.OldUnits, &d1)
	} else {
		//新的单位数据
		d1 := *unit.ClientData
		this.Msg.NewUnits = append(this.Msg.NewUnits, &d1)
	}

}

//
//添加客户端显示子弹数据包
func (this *Player) AddHaloData(halo *Halo) {

	//如果客户端不需要显示
	if halo.ClientIsShow() == false {
		return
	}

	this.CurShowHalo[halo.ID] = halo

	if _, ok := this.LastShowHalo[halo.ID]; ok {
		//旧单位(只更新变化的值)
		d1 := *halo.ClientDataSub
		this.Msg.OldHalos = append(this.Msg.OldHalos, &d1)
	} else {
		//新的单位数据
		d1 := *halo.ClientData
		this.Msg.NewHalos = append(this.Msg.NewHalos, &d1)
	}

}

//this.LastShowSceneItem = make(map[int32]*SceneItem)
//this.CurShowSceneItem = make(map[int32]*SceneItem)
//添加客户端显示子弹数据包
func (this *Player) AddSceneItemData(sceneitem *SceneItem) {

	this.CurShowSceneItem[sceneitem.ID] = sceneitem

	if _, ok := this.LastShowSceneItem[sceneitem.ID]; ok {
		//旧单位(只更新变化的值)
		//d1 := *bullet.ClientDataSub
		//this.Msg.OldBullets = append(this.Msg.OldBullets, &d1)
	} else {
		//新的单位数据
		d1 := *sceneitem.ClientData
		this.Msg.NewSceneItems = append(this.Msg.NewSceneItems, &d1)
	}

}

//添加客户端显示子弹数据包
func (this *Player) AddBulletData(bullet *Bullet) {

	//如果客户端不需要显示
	if bullet.ClientIsShow() == false {
		return
	}

	this.CurShowBullet[bullet.ID] = bullet

	if _, ok := this.LastShowBullet[bullet.ID]; ok {
		//旧单位(只更新变化的值)
		d1 := *bullet.ClientDataSub
		this.Msg.OldBullets = append(this.Msg.OldBullets, &d1)
	} else {
		//新的单位数据
		d1 := *bullet.ClientData
		this.Msg.NewBullets = append(this.Msg.NewBullets, &d1)
	}

}
func (this *Player) AddHurtValue(hv *protomsg.MsgPlayerHurt) {
	if hv == nil || (hv.HurtAllValue == 0 && hv.GetGold == 0) {
		return
	}

	this.Msg.PlayerHurt = append(this.Msg.PlayerHurt, hv)
}

func (this *Player) AutoSaveDB() {

	if this.CurScene != nil {
		this.AutoSaveRemainTime -= 1.0 / float32(this.CurScene.SceneFrame)
		if this.AutoSaveRemainTime <= 0 {
			this.AutoSaveRemainTime += AutoSaveTime
			go func() {
				this.SaveDB()
			}()
		}
	}

	//AutoSaveTime
}

func (this *Player) Update(curframe int32) {
	this.AutoSaveDB()
	this.SendUpdateMsg(curframe)

}

func (this *Player) SendUpdateMsg(curframe int32) {

	//删除的单位 id
	for k, _ := range this.LastShowUnit {
		if _, ok := this.CurShowUnit[k]; !ok {
			this.Msg.RemoveUnits = append(this.Msg.RemoveUnits, k)
		}
	}
	//删除的子弹 id
	for k, _ := range this.LastShowBullet {
		if _, ok := this.CurShowBullet[k]; !ok {
			this.Msg.RemoveBullets = append(this.Msg.RemoveBullets, k)
		}
	}
	//Halo
	//删除的Halo id
	for k, _ := range this.LastShowHalo {
		if _, ok := this.CurShowHalo[k]; !ok {
			this.Msg.RemoveHalos = append(this.Msg.RemoveHalos, k)
		}
	}
	//删除的sceneitem id
	for k, _ := range this.LastShowSceneItem {
		if _, ok := this.CurShowSceneItem[k]; !ok {
			this.Msg.RemoveSceneItems = append(this.Msg.RemoveSceneItems, k)
		}
	}

	//回复客户端
	this.Msg.CurFrame = curframe
	this.SendMsgToClient("SC_Update", this.Msg)

	//重置数据
	this.LastShowUnit = this.CurShowUnit
	this.CurShowUnit = make(map[int32]*Unit)

	this.LastShowBullet = this.CurShowBullet
	this.CurShowBullet = make(map[int32]*Bullet)

	this.LastShowHalo = this.CurShowHalo
	this.CurShowHalo = make(map[int32]*Halo)

	this.LastShowSceneItem = this.CurShowSceneItem
	this.CurShowSceneItem = make(map[int32]*SceneItem)
	this.Msg = &protomsg.SC_Update{}

}
func (this *Player) SendNoticeWordToClient(typeid int32, param ...string) {
	msg := &protomsg.SC_NoticeWords{}
	msg.TypeID = typeid
	msg.P = param
	this.SendMsgToClient("SC_NoticeWords", msg)
}
func (this *Player) SendNoticeWordToClientP(typeid int32, param []string) {
	msg := &protomsg.SC_NoticeWords{}
	msg.TypeID = typeid
	msg.P = param
	this.SendMsgToClient("SC_NoticeWords", msg)
}

func (this *Player) SendMsgToClient(msgtype string, msg proto.Message) {
	data := &protomsg.MsgBase{}
	data.ConnectId = this.ConnectId
	data.ModeType = "Client"
	data.Uid = this.Uid
	data.MsgType = msgtype

	this.ServerAgent.WriteMsgBytes(datamsg.NewMsg1Bytes(data, msg))

}

//退出场景
func (this *Player) OutScene() {

	if this.CurScene != nil {
		this.CurScene.PlayerGoout(this)
		this.CurScene = nil
	}
	this.ReInit()

}

//进入场景
func (this *Player) GoInScene(scene *Scene, datas []byte) {
	if this.CurScene != nil {
		this.CurScene.PlayerGoout(this)
		this.CurScene = nil
	}
	this.CurScene = scene

	characterinfo := db.DB_CharacterInfo{}
	utils.Bytes2Struct(datas, &characterinfo)
	this.Characterid = characterinfo.Characterid
	this.LoadBagInfoFromDB(characterinfo.BagInfo)
	this.LoadItemSkillCDFromDB(characterinfo.ItemSkillCDInfo)
	//好友信息
	this.MyFriends = NewFriends(characterinfo.Friends, characterinfo.FriendsRequest, this)
	//邮件信息
	this.MyMails = NewMails(characterinfo.Mails, this)
	//公会信息
	if characterinfo.GuildId > 0 {
		this.MyGuild = NewGuildCharacterInfo(&characterinfo)
	} else {
		this.MyGuild = nil
	}

	this.CurScene.PlayerGoin(this, &characterinfo)
	//this.ReInit()
}

//玩家移动操作
func (this *Player) MoveCmd(data *protomsg.CS_PlayerMove) {
	this.CheckOtherUnit()
	for _, v := range data.IDs {
		if this.MainUnit.ID == v {
			this.MainUnit.PlayerControl_MoveCmd(data)

			this.CheckOtherUnit()
			items := this.OtherUnit.Items()
			for _, v1 := range items {
				v1.(*Unit).PlayerControl_MoveCmd(data)
			}
		}

	}
}

//SkillCmd
func (this *Player) SkillCmd(data *protomsg.CS_PlayerSkill) {
	this.CheckOtherUnit()
	if this.MainUnit.ID == data.ID {
		this.MainUnit.PlayerControl_SkillCmd(data)
	}
}

//玩家攻击操作
func (this *Player) AttackCmd(data *protomsg.CS_PlayerAttack) {
	this.CheckOtherUnit()
	for _, v := range data.IDs {
		if this.MainUnit.ID == v {
			this.MainUnit.AttackCmd(data)

			this.CheckOtherUnit()
			items := this.OtherUnit.Items()
			for _, v1 := range items {
				v1.(*Unit).PlayerControl_AttackCmd(data)
			}
		}
	}
}
