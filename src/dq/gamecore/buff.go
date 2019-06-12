package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/utils"
	"dq/vec2d"
)

type Buff struct {
	conf.BuffData //技能数据

	Parent            *Unit   //载体
	CastUnit          *Unit   //施加buff的单位
	Level             int32   //当前等级
	RemainTime        float32 //剩余时间
	TagNum            int32   //标记数字 拍拍熊被动 大圣被动
	TriggerRemainTime float32 //触发剩余时间 造成伤害之类的

	IsEnd bool //是否已经结束

	IsActive bool //是否生效

	LastPos vec2d.Vec2 //载体上次计算的位置
}

func (this *Buff) HurtTrigger() {
	if this.Exception <= 0 {
		return
	}
	switch this.Exception {
	case 3: //血魔的大招
		{
			this.HurtValue = 0
			if this.Parent == nil || this.Parent.Body == nil {
				return
			}
			param := utils.GetFloat32FromString3(this.ExceptionParam, ":")
			if len(param) < 1 {
				return
			}

			dis := vec2d.Sub(this.Parent.Body.Position, this.LastPos)
			this.HurtValue = param[0] * float32(dis.Length())
			log.Info("----hurt:%f", this.HurtValue)
			this.LastPos = this.Parent.Body.Position
		}
	default:
		{

		}
	}
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
					castunit := this.Parent
					if this.CastUnit != nil {
						castunit = this.CastUnit
					}
					b := NewBullet1(castunit, this.Parent)
					b.SetProjectileMode("", 0)
					this.HurtTrigger()
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
	buff.TagNum = buffdata.InitTagNum
	buff.IsEnd = false

	if buffdata.ActiveTime <= 0 {
		buff.IsActive = true
	} else {
		buff.IsActive = false
	}
	buff.Parent = parent
	if parent != nil && parent.Body != nil {
		buff.LastPos = parent.Body.Position
	}
	buff.TriggerRemainTime = 0

	return buff

}
