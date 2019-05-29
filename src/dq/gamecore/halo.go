package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/vec2d"
	"strings"
)

var HaloID int32 = 100

func GetHaloID() int32 {
	HaloID++
	if HaloID >= 100000000 {
		HaloID = 100
	}
	return HaloID
}

type Halo struct {
	conf.HaloData                //技能数据
	ID                int32      //光环ID
	Parent            *Unit      //载体
	Position          vec2d.Vec2 //位置
	PositionZ         float32    //z
	Level             int32      //当前等级
	RemainTime        float32    //剩余时间
	TriggerRemainTime float32    //触发剩余时间 造成伤害之类的
	IsEnd             bool       //是否已经结束
	IsActive          bool       //是否生效

	//发送数据部分
	ClientData    *protomsg.HaloDatas //客户端显示数据
	ClientDataSub *protomsg.HaloDatas //客户端显示差异数据
}

//更新
func (this *Halo) Update(dt float32) {
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
			this.TriggerRemainTime = this.Cooldown + this.TriggerRemainTime

			if this.Parent != nil && this.Parent.InScene != nil && this.Parent.IsDisappear() == false {
				//创建触发子弹 //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害 4:不造成伤害)
				//创建buff
				if this.HurtType != 4 || len(this.InitBuff) > 0 {
					//获取范围内的目标单位
					allunit := this.Parent.InScene.FindVisibleUnits(this.Parent)
					count := int32(0)
					//log.Info("------------------len:%d", len(allunit))
					for _, v := range allunit {

						//创建子弹
						if this.HurtType != 4 && count < this.UnitTargetMaxCount {
							//UnitTargetTeam      int32   //目标单位关系 1:友方  2:敌方 3:友方敌方都行
							if this.UnitTargetTeam == 1 && this.Parent.CheckIsEnemy(v) == true {
								continue
							}
							if this.UnitTargetTeam == 2 && this.Parent.CheckIsEnemy(v) == false {
								continue
							}
							//技能免疫检测
							if this.NoCareMagicImmune == 2 && v.MagicImmune == 1 {
								continue
							}
							//检测是否在范围内
							if v.Body == nil || this.HaloRange <= 0 {
								continue
							}
							dis := float32(vec2d.Distanse(this.Position, v.Body.Position))
							log.Info("-----------------dis:%f", dis)
							if dis <= this.HaloRange {
								b := NewBullet1(this.Parent, v)
								b.SetNormalHurtRatio(this.NormalHurt)
								b.SetProjectileMode(this.BulletModeType, this.BulletSpeed)
								//技能增强
								if this.HurtType == 2 {
									hurtvalue := (this.HurtValue + int32(float32(this.HurtValue)*this.Parent.MagicScale))
									b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: hurtvalue})
								} else {
									b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: this.HurtValue})
								}
								b.AddTargetBuff(this.TargetBuff, this.Level)
								if b != nil {
									if this.TriggerAttackEffect == 1 {
										this.Parent.CheckTriggerAttackSkill(b)
									}
									log.Info("----------------add bullet")
									this.Parent.AddBullet(b)
									count++
								}
							}
						}
						//增加buff
						v.AddBuffFromStr(this.InitBuff, this.Level, this.Parent)

					}
				}
			}

		}
		//
	} else {
		//		this.ActiveTime -= float32(dt)
		//		if this.ActiveTime <= 0 {
		//			this.ActiveTime = 0
		//			this.IsActive = true
		//		} else {
		//			this.IsActive = false
		//		}

	}

	//是否跟随主角
	if this.FollowParent == 1 && this.Parent != nil {
		if this.Parent.Body != nil {
			this.Position = this.Parent.Body.Position
		}
		if this.Parent.IsDisappear() {
			this.IsEnd = true
		}
	}

}

//设置载体
func (this *Halo) SetParent(parent *Unit) {
	if this.Parent != nil {
		this.Parent = nil
	}
	this.Parent = parent
	if parent.Body != nil {
		this.Position = parent.Body.Position
	}

}

//创建buf
func NewHalo(typeid int32, level int32) *Halo {

	halodata := conf.GetHaloData(typeid, level)
	if halodata == nil {
		log.Error("NewHalo %d  %d", typeid, level)
		return nil
	}
	halo := &Halo{}
	halo.HaloData = *halodata
	halo.Level = level
	halo.RemainTime = halodata.Time
	halo.TriggerRemainTime = 0
	halo.IsEnd = false
	halo.IsActive = true
	//	if halodata.ActiveTime <= 0 {
	//		halo.IsActive = true
	//	} else {
	//		halo.IsActive = false
	//	}

	//唯一ID处理
	halo.ID = GetHaloID()
	halo.PositionZ = 0.1

	return halo

}
func (this *Halo) IsDone() bool {
	if this.IsEnd == true {
		return true
	}
	return false
}

//客户端是否要显示
func (this *Halo) ClientIsShow() bool {
	if len(this.HaloModeType) > 0 {
		return true
	}
	return false
}

//刷新客户端显示数据
func (this *Halo) FreshClientData() {
	if this.ClientIsShow() == false {
		return
	}
	if this.ClientData == nil {
		this.ClientData = &protomsg.HaloDatas{}
	}

	this.ClientData.ID = this.ID

	this.ClientData.X = float32(this.Position.X)
	this.ClientData.Y = float32(this.Position.Y)
	this.ClientData.Z = float32(this.PositionZ)

	this.ClientData.ModeType = this.HaloModeType

}

//刷新客户端显示差异数据
func (this *Halo) FreshClientDataSub() {
	if this.ClientIsShow() == false {
		return
	}
	if this.ClientDataSub == nil {
		this.ClientDataSub = &protomsg.HaloDatas{}
	}
	if this.ClientData == nil {
		this.FreshClientData()
		*this.ClientDataSub = *this.ClientData
		return
	}
	//
	//字符串部分
	if strings.Compare(this.HaloModeType, this.ClientData.ModeType) != 0 {
		this.ClientDataSub.ModeType = this.HaloModeType
	} else {
		this.ClientDataSub.ModeType = ""
	}

	//当前数据与上一次数据对比 相减 数值部分
	this.ClientDataSub.X = float32(this.Position.X) - this.ClientData.X
	this.ClientDataSub.Y = float32(this.Position.Y) - this.ClientData.Y
	this.ClientDataSub.Z = float32(this.PositionZ) - this.ClientData.Z

}
