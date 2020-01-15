package gamecore

import (
	"dq/conf"
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/timer"
	"dq/utils"
	"sync"
	"time"
)

var (
	ExchangeManagerObj = &ExchangeManager{}
)

type ExchangeManager struct {
	Commoditys     *utils.BeeMap //当前服务器交易所的商品
	CommodityCount *utils.BeeMap //记录相同种类道具的售卖数量
	OperateLock    *sync.RWMutex //同步操作锁
	Server         ServerInterface
	//时间到 倒计时
	UpdateTimer *timer.Timer
}

//从数据库载入数据
func (this *ExchangeManager) LoadDataFromDB() {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	commoditys := make([]db.DB_PlayerItemTransactionInfo, 0)
	db.DbOne.GetExchanges(&commoditys)
	for k, v := range commoditys {
		//log.Info("----------ExchangeManager load %d %v", v.Id, &commoditys[k])
		this.Commoditys.Set(v.Id, &commoditys[k])
		this.CommodityCount.AddInt2(v.ItemID, 1)
	}

	//	teams := this.Commoditys.Items()
	//	log.Info("11---count: %d", len(teams))
	//	for _, v := range teams {
	//		if v == nil {
	//			continue
	//		}
	//		one := v.(*db.DB_PlayerItemTransactionInfo)
	//		log.Info("11---one: %d", one.Id)
	//	}
}

//初始化
func (this *ExchangeManager) Init(server ServerInterface) {
	log.Info("----------ExchangeManager Init---------")
	this.Server = server
	this.Commoditys = utils.NewBeeMap()
	this.CommodityCount = utils.NewBeeMap()
	this.OperateLock = new(sync.RWMutex)

	this.LoadDataFromDB()

	this.UpdateTimer = timer.AddRepeatCallback(time.Second*10, this.Update)
}
func (this *ExchangeManager) Close() {
	if this.UpdateTimer != nil {
		this.UpdateTimer.Cancel()
		this.UpdateTimer = nil
	}
}

////返回正在售卖的道具数据short SC_GetExchangeShortCommoditys
func (this *ExchangeManager) GetExchangeShortCommoditys() *protomsg.SC_GetExchangeShortCommoditys {
	this.OperateLock.RLock()
	defer this.OperateLock.RUnlock()
	data := &protomsg.SC_GetExchangeShortCommoditys{}
	data.Commoditys = make([]*protomsg.ExchangeShortCommodityData, 0)
	items := this.CommodityCount.Items()
	for k, v := range items {
		if v.(int) <= 0 {
			continue
		}
		d1 := &protomsg.ExchangeShortCommodityData{}
		d1.ItemID = k.(int32)
		d1.SellCount = int32(v.(int))
		data.Commoditys = append(data.Commoditys, d1)
	}
	return data
}

//返回正在售卖的道具数据
func (this *ExchangeManager) GetExchangeDetailedCommoditys(itemid int32) *protomsg.SC_GetExchangeDetailedCommoditys {
	this.OperateLock.RLock()
	defer this.OperateLock.RUnlock()
	data := &protomsg.SC_GetExchangeDetailedCommoditys{}
	data.Commoditys = make([]*protomsg.ExchangeDetailedCommodityData, 0)
	items := this.Commoditys.Items()
	for _, v := range items {
		if v == nil {
			continue
		}
		itemv := v.(*db.DB_PlayerItemTransactionInfo)
		if itemv.ItemID != itemid {
			continue
		}
		remaindtime := int32(24*60*60) - int32(utils.GetCurTimeOfSecond()-float64(itemv.ShelfTime))
		if remaindtime <= 0 {
			continue
		}

		d1 := &protomsg.ExchangeDetailedCommodityData{}
		d1.CommodityData = &protomsg.CommodityDataProto{}
		d1.CommodityData.TypeID = itemv.Id
		d1.CommodityData.ItemID = itemv.ItemID
		d1.CommodityData.PriceType = itemv.PriceType
		d1.CommodityData.Price = itemv.Price
		d1.CommodityData.Level = itemv.Level
		d1.RemaindTime = remaindtime
		data.Commoditys = append(data.Commoditys, d1)
	}
	return data
}

//上架商品(本函数未删除玩家背包里的道具)
func (this *ExchangeManager) ShelfExchangeCommodity(data *db.DB_PlayerItemTransactionInfo) {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()

	data.ShelfTime = int32(utils.GetCurTimeOfSecond())
	db.DbOne.CreateAndSaveCommodity(data)

	this.Commoditys.Set(data.Id, data)
	this.CommodityCount.AddInt2(data.ItemID, 1)
}

//购买商品
func (this *ExchangeManager) BuyExchangeCommodity(data *protomsg.CS_BuyExchangeCommodity, buyplayer *Player) bool {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	if data == nil || buyplayer == nil {
		return false
	}
	if buyplayer.MyMails == nil {
		return false
	}
	commodityv := this.Commoditys.Get(data.ID)
	if commodityv == nil {
		//商品不存在了
		buyplayer.SendNoticeWordToClient(24)
		return false
	}
	commodity := commodityv.(*db.DB_PlayerItemTransactionInfo)
	//买家扣钱
	if buyplayer.BuyItemSubMoneyLock(commodity.PriceType, commodity.Price) == false {
		//货币不足
		buyplayer.SendNoticeWordToClient(commodity.PriceType)
		return false
	}
	//给买家发货(邮件)
	buyplayer.MyMails.BuyCommodityMail(commodity.ItemID, commodity.Level)
	//买家购买成功
	buyplayer.SendNoticeWordToClient(8)
	//给卖家钱
	sellplayer := this.Server.GetPlayerByID(commodity.SellerUid)
	//卖家在线
	//扣除税收后的收入系数
	ratio := 1 - float32(conf.Conf.NormalInfo.SellExchangeTax)
	getprice := float32(commodity.Price) * ratio
	if sellplayer != nil && sellplayer.Characterid == commodity.SellerCharacterid && sellplayer.MyMails != nil {
		sellplayer.MyMails.SellCommodityMail(commodity.PriceType, int32(getprice))
	} else {
		mi := Create_SellCommodityMail_Mail(commodity.PriceType, int32(getprice))
		mi.RecUid = commodity.SellerUid
		mi.RecCharacterid = commodity.SellerCharacterid
		db.DbOne.CreateAndSaveMail(&mi.DB_MailInfo)
		db.DbOne.AddMail(commodity.SellerCharacterid, mi.DB_MailInfo.Id)
	}

	//删除商品
	this.Commoditys.Delete(commodity.Id)
	this.CommodityCount.AddInt2(commodity.ItemID, -1)
	//删除数据库中的商品commodity
	db.DbOne.DeleteCommodity(commodity.Id)
	return true

}

//获取玩家正在售卖的装备数量
func (this *ExchangeManager) GetPlayerSellingCount(chaid int32) int32 {
	this.OperateLock.RLock()
	defer this.OperateLock.RUnlock()
	count := int32(0)
	items := this.Commoditys.Items()
	for _, v := range items {
		if v == nil {
			continue
		}
		itemv := v.(*db.DB_PlayerItemTransactionInfo)
		if itemv.SellerCharacterid == chaid {
			count++
		}
	}

	return count
}

//获取玩家正在卖的装备
func (this *ExchangeManager) GetPlayerSelling(chaid int32) []*protomsg.ExchangeDetailedCommodityData {
	this.OperateLock.RLock()
	defer this.OperateLock.RUnlock()
	commoditys := make([]*protomsg.ExchangeDetailedCommodityData, 0)
	items := this.Commoditys.Items()
	for _, v := range items {
		if v == nil {
			continue
		}
		itemv := v.(*db.DB_PlayerItemTransactionInfo)
		if itemv.SellerCharacterid == chaid {
			remaindtime := int32(24*60*60) - int32(utils.GetCurTimeOfSecond()-float64(itemv.ShelfTime))
			if remaindtime <= 0 {
				continue
			}

			d1 := &protomsg.ExchangeDetailedCommodityData{}
			d1.CommodityData = &protomsg.CommodityDataProto{}
			d1.CommodityData.TypeID = itemv.Id
			d1.CommodityData.ItemID = itemv.ItemID
			d1.CommodityData.PriceType = itemv.PriceType
			d1.CommodityData.Price = itemv.Price
			d1.CommodityData.Level = itemv.Level
			d1.RemaindTime = remaindtime
			commoditys = append(commoditys, d1)
		}
	}

	return commoditys
}

//下架道具
func (this *ExchangeManager) UnShelfExchangeCommodity_Lock(id int32) bool {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	return this.UnShelfExchangeCommodity_NoLock(id)
}

//下架道具
func (this *ExchangeManager) UnShelfExchangeCommodity_NoLock(id int32) bool {

	log.Info("----------UnShelfExchangeCommodity_NoLock id %d", id)

	commodityv := this.Commoditys.Get(id)
	if commodityv == nil {
		//商品不存在了
		return false
	}
	commodity := commodityv.(*db.DB_PlayerItemTransactionInfo)

	sellplayer := this.Server.GetPlayerByID(commodity.SellerUid)

	//返还道具给卖家
	//卖家在线
	if sellplayer != nil && sellplayer.Characterid == commodity.SellerCharacterid && sellplayer.MyMails != nil {
		sellplayer.MyMails.UnShelfCommodityMail(commodity.ItemID, commodity.Level)
	} else {
		mi := Create_UnShelfCommodityMail_Mail(commodity.ItemID, commodity.Level)
		mi.RecUid = commodity.SellerUid
		mi.RecCharacterid = commodity.SellerCharacterid
		db.DbOne.CreateAndSaveMail(&mi.DB_MailInfo)
		db.DbOne.AddMail(commodity.SellerCharacterid, mi.DB_MailInfo.Id)
	}
	//删除商品
	this.Commoditys.Delete(commodity.Id)
	this.CommodityCount.AddInt2(commodity.ItemID, -1)
	//删除数据库中的商品commodity
	db.DbOne.DeleteCommodity(commodity.Id)
	return true
}

//更新
func (this *ExchangeManager) Update() {
	this.OperateLock.Lock()
	defer this.OperateLock.Unlock()
	alltime := int32(conf.Conf.NormalInfo.AutoUnShelfTime)
	//检查玩家上架的物品是否过期
	teams := this.Commoditys.Items()
	//log.Info("---count: %d", len(teams))
	for _, v := range teams {
		if v == nil {
			continue
		}
		one := v.(*db.DB_PlayerItemTransactionInfo)
		//log.Info("---one: %d", one.Id)

		remaindtime := alltime - int32(utils.GetCurTimeOfSecond()-float64(one.ShelfTime))
		//log.Info("ExchangeManager:%d  %d  %d", alltime, remaindtime, int32(one.ShelfTime))
		if remaindtime <= 0 {
			//超时了就下架
			this.UnShelfExchangeCommodity_NoLock(one.Id)
			continue
		}

	}
}
