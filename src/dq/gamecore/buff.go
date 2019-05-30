package gamecore

import (
	"dq/conf"
	"dq/log"
)

type Buff struct {
	conf.BuffData //技能数据

	Parent            *Unit   //载体
	Level             int32   //当前等级
	RemainTime        float32 //剩余时间
	TagNum            int32   //标记数字 拍拍熊被动 大圣被动
	TriggerRemainTime float32 //触发剩余时间 造成伤害之类的

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

		this.TriggerRemainTime -= float32(dt)
		//检查是否触发
		if this.TriggerRemainTime <= 0 {
			//重置触发时间
			this.TriggerRemainTime = this.HurtTimeInterval + this.TriggerRemainTime

			if this.Parent != nil && this.Parent.InScene != nil && this.Parent.IsDisappear() == false {
				//创建触发子弹 //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害 4:不造成伤害)
				//创建buff
				if this.HurtType != 4 {
					b := NewBullet1(this.Parent, this.Parent)
					b.SetProjectileMode("", 0)
					b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: int32(this.HurtValue)})
					if b != nil {
						this.Parent.AddBullet(b)
					}
				}
			}
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
func NewBuff(typeid int32, level int32, parent *Unit) *Buff {

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
	buff.Parent = parent
	buff.TriggerRemainTime = 0

	return buff

}
