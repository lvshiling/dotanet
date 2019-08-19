package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/utils"
)

type Item struct {
	conf.ItemData         //技能数据
	Parent        *Unit   //载体
	UnitBuffs     []*Buff //单位身上的buf
}

//删除道具的属性到单位身上
func (this *Item) Clear() {
	if this.Parent == nil {
		return
	}
	for _, v := range this.UnitBuffs {
		v.IsEnd = true
	}

	this.Parent = nil
}

//添加道具的属性到单位身上
func (this *Item) Add2Unit(unit *Unit) {
	this.Parent = unit
	this.UnitBuffs = unit.AddBuffFromStr(this.Buffs, 1, unit)

	log.Error("NewItembuf %s ", this.Buffs)
}

//(dbdata []string,
//创建buf
func NewItemFromDB(dbdata string) *Item {
	if len(dbdata) <= 0 {
		return nil
	}

	param := utils.GetFloat32FromString3(dbdata, ",")
	if len(param) < 1 {
		return nil
	}
	typeid := int32(param[0])

	itemdata := conf.GetItemData(typeid)
	if itemdata == nil {
		log.Error("NewItem %d ", typeid)
		return nil
	}
	item := &Item{}
	item.ItemData = *itemdata
	//item.Parent = parent
	return item
}

//创建buf
func NewItem(typeid int32) *Item {

	itemdata := conf.GetItemData(typeid)
	if itemdata == nil {
		log.Error("NewItem %d ", typeid)
		return nil
	}
	item := &Item{}
	item.ItemData = *itemdata
	//item.Parent = parent
	return item
}
