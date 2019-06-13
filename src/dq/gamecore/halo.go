package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
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
	CastUnit          *Unit      //施加buff的单位
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

func (this *Halo) DoHurtException(b *Bullet) {
	if this.Exception <= 0 {
		return
	}
	switch this.Exception {
	case 1: //小小山崩对投掷状态的单位造成3倍伤害
		{
			param := utils.GetFloat32FromString3(this.ExceptionParam, ":")
			if len(param) < 2 || b.DestUnit == nil || b.DestUnit.IsDisappear() {
				return
			}
			//投掷buff
			buff := b.DestUnit.GetBuff(int32(param[0]))
			if buff == nil {
				return
			}
			if len(b.OtherHurt) > 0 {
				b.OtherHurt[0].HurtValue = int32(float32(b.OtherHurt[0].HurtValue) * param[1])
			}

		}
	default:
		{

		}
	}
}

func (this *Halo) GetCastUnit() *Unit {
	if this.CastUnit != nil {
		return this.CastUnit
	}
	return this.Parent
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

			if this.Parent != nil && this.Parent.InScene != nil && this.Parent.IsDisappear() == false && this.GetCastUnit().IsDisappear() == false {
				//创建触发子弹 //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害 4:不造成伤害)
				//创建buff
				if this.HurtType != 4 || len(this.InitBuff) > 0 {
					//获取范围内的目标单位
					allunit := this.GetCastUnit().InScene.FindVisibleUnitsByPos(this.Position)
					count := int32(0)
					//log.Info("------------------len:%d", len(allunit))
					for _, v := range allunit {

						//创建子弹
						if count < this.UnitTargetMaxCount {
							if v.IsDisappear() {
								continue
							}
							//UnitTargetTeam      int32   //目标单位关系 1:友方  2:敌方 3:友方敌方都行 5:自己 10:除自己外的其他

							if this.UnitTargetTeam == 1 && this.GetCastUnit().CheckIsEnemy(v) == true {
								continue
							}
							if this.UnitTargetTeam == 2 && this.GetCastUnit().CheckIsEnemy(v) == false {
								continue
							}
							if this.UnitTargetTeam == 5 && this.GetCastUnit() != v {
								continue
							}
							if this.UnitTargetTeam == 10 && this.GetCastUnit() == v {
								continue
							}
							//检测是否在范围内
							if v.Body == nil || this.HaloRange <= 0 {
								continue
							}
							dis := float32(vec2d.Distanse(this.Position, v.Body.Position))
							//log.Info("-----------------dis:%f", dis)
							if dis <= this.HaloRange {

								count++

								//增加buff
								v.AddBuffFromStr(this.InitBuff, this.Level, this.GetCastUnit())
								//log.Info("-----------------InitBuff:%s", this.InitBuff)
								//BlinkToTarget
								if this.BlinkToTarget == 1 {

									this.GetCastUnit().Body.BlinkToPos(v.Body.Position, float64(utils.GetRandomFloat(180)))
									this.GetCastUnit().SetDirection(vec2d.Sub(v.Body.Position, this.GetCastUnit().Body.Position))
								}

								//技能免疫检测
								if this.HurtType == 2 && this.NoCareMagicImmune == 2 && v.MagicImmune == 1 {
									continue
								}
								//不造成伤害
								if this.HurtType == 4 {
									continue
								}

								b := NewBullet1(this.GetCastUnit(), v)
								b.SetNormalHurtRatio(this.NormalHurt)
								b.SetProjectileMode(this.BulletModeType, this.BulletSpeed)
								//技能增强
								if this.HurtType == 2 {
									hurtvalue := (this.HurtValue + int32(float32(this.HurtValue)*this.GetCastUnit().MagicScale))
									b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: hurtvalue})
								} else {
									b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: this.HurtValue})
								}
								//特殊情况处理
								this.DoHurtException(b)
								b.AddTargetBuff(this.TargetBuff, this.Level)
								if b != nil {
									if this.TriggerAttackEffect == 1 {
										this.GetCastUnit().CheckTriggerAttackSkill(b)
									}
									//log.Info("----------------add bullet")
									this.GetCastUnit().AddBullet(b)

								}
							}
						}

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

	if parent == nil {
		return
	}

	if this.Parent != nil {
		this.Parent = nil
	}
	this.Parent = parent
	if parent.Body != nil {
		this.Position = parent.Body.Position
	}

	if int32(this.Cooldown) == int32(-1) {
		this.Cooldown = parent.GetOneAttackTime()
		this.TriggerRemainTime = this.Cooldown
	}

}

//创建buf
func NewHalo(typeid int32, level int32) *Halo {

	log.Info("---new halo:%d   %d", typeid, level)

	halodata := conf.GetHaloData(typeid, level)
	if halodata == nil {
		log.Error("NewHalo %d  %d", typeid, level)
		return nil
	}
	halo := &Halo{}
	halo.HaloData = *halodata
	halo.Level = level
	halo.RemainTime = halodata.Time
	halo.TriggerRemainTime = halodata.Cooldown
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
