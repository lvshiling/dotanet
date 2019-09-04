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

	//ScenePosition vec2d.Vec2 //场景里的位置
	UnitSkills []*Skill //

	UnitHalos []*Halo //单位身上的halo
	Index     int32   //位置索引
}

//删除道具的属性到单位身上
func (this *Item) Clear() {
	if this.Parent == nil {
		return
	}
	//清除buf
	for _, v := range this.UnitBuffs {
		v.IsEnd = true
	}
	this.UnitBuffs = make([]*Buff, 0)
	//清除技能
	for _, v := range this.UnitSkills {
		this.Parent.RemoveItemSkill(v)
	}
	this.UnitSkills = make([]*Skill, 0)

	//消除halo this.InScene.RemoveHalo(v1)
	for _, v := range this.UnitHalos {
		if this.Parent.InScene != nil {
			this.Parent.InScene.RemoveHalo(v.ID)
		}
	}
	this.UnitHalos = make([]*Halo, 0)

	this.Parent = nil
}

//设置技能图标显示位置索引
func (this *Item) SetIndex(index int32) {
	this.Index = index
	//技能
	for _, v := range this.UnitSkills {
		v.Index = index
	}
}

//添加道具的属性到单位身上
func (this *Item) Add2Unit(unit *Unit, index int32) {
	this.Parent = unit
	this.UnitBuffs = unit.AddBuffFromStr(this.Buffs, 1, unit)
	this.Index = index

	//技能
	skills := utils.GetInt32FromString3(this.Skills, ",")
	for _, v := range skills {
		skill := NewOneSkill(v, 1, unit)
		if skill != nil {
			skill.Index = index
			ok := unit.AddItemSkill(skill)
			if ok == true {
				this.UnitSkills = append(this.UnitSkills, skill)
			}
		}
	}

	//光环
	this.UnitHalos = unit.AddHaloFromStr(this.Halos, 1, nil)

	//

	log.Info("NewItembuf %s ", this.Buffs)
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
	item.UnitBuffs = make([]*Buff, 0)
	item.UnitSkills = make([]*Skill, 0)
	item.ItemData = *itemdata
	//item.ScenePosition = vec2d.Vec2{X: 0, Y: 0}
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
	item.UnitBuffs = make([]*Buff, 0)
	item.UnitSkills = make([]*Skill, 0)
	item.ItemData = *itemdata
	//item.ScenePosition = vec2d.Vec2{X: 0, Y: 0}
	//item.Parent = parent
	return item
}
