package gamecore

import (
	"dq/conf"
	"dq/log"
)

type Buff struct {
	conf.BuffData //技能数据

	Level      int32   //当前等级
	RemainTime float32 //剩余时间
	TagNum     int32   //标记数字 拍拍熊被动 大圣被动

	IsEnd bool //是否已经结束

	IsActive bool //是否生效
}

//更新
func (this *Buff) Update(dt float64) {
	//CD时间减少
	if this.IsActive {
		this.RemainTime -= float32(dt)
		if this.RemainTime <= 0 {
			this.RemainTime = 0
			this.IsEnd = true
		}
	} else {
		this.ActiveTime -= float32(dt)
		if this.ActiveTime <= 0 {
			this.ActiveTime = 0
			this.IsActive = true
		} else {
			this.IsActive = false
		}
	}

}

//创建buf
func NewBuff(typeid int32, level int32) *Buff {

	buffdata := conf.GetBuffData(typeid, level)
	if buffdata == nil {
		log.Error("NewBuff %d  %d", typeid, level)
		return nil
	}
	buff := &Buff{}
	buff.BuffData = *buffdata
	buff.Level = level
	buff.RemainTime = buffdata.Time
	buff.TagNum = 0
	buff.IsEnd = false

	if buffdata.ActiveTime <= 0 {
		buff.IsActive = true
	} else {
		buff.IsActive = false
	}

	return buff

}
