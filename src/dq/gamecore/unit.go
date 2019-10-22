package gamecore

import (
	"dq/conf"
	"dq/cyward"
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"math"
	//"math"
	"strconv"
	"strings"
	"time"
)

var UnitID int32 = 1000000
var UnitEquitCount int32 = 6

func (this *Unit) TestData() {
	this.AttackRange = 2
	this.AttackRangeBuffer = 5

	this.Camp = 1
	this.IsDeath = 2

	this.AutoAttackTraceRange = 5
	this.AutoAttackTraceOutRange = 10

	this.HP = 1000
	this.MP = 1000
	this.MAX_HP = 1000
	this.MAX_MP = 1000
}

//获取1次攻击所需的时间 (141点攻击速度等于 1.20秒一次) 170/攻击速度 等于攻击时间
func (this *Unit) GetOneAttackTime() float32 {

	return 170.0 / this.AttackSpeed

	//return 5
}

//目标离自动攻击范围的距离 小于0 表示在内
func (this *Unit) GetDistanseOfAutoAttackRange(target *Unit) float64 {
	if target == nil {
		return 10
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()

	return targetdis - float64(this.AutoAttackTraceRange)
}

//目标是否脱离 自动攻击取消追击范围
func (this *Unit) IsOutAutoAttackTraceOutRange(target *Unit) bool {

	if target == nil {
		return true
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()
	if targetdis <= float64(this.AutoAttackTraceOutRange) {
		return false
	}

	return true
}

//目标是否脱离 前摇中断范围
func (this *Unit) IsOutAttackRangeBuffer(target *Unit) bool {

	if target == nil {
		return true
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()
	if targetdis <= float64(this.AttackRange)+float64(this.AttackRangeBuffer) {
		return false
	}

	return true
}

//单位是否能被攻击 (各种BUF状态 无敌状态 免疫状态)
func (this *Unit) IsCanBeAttack() bool {
	//死亡后不能攻击
	if this.IsDeath == 1 {
		return false
	}
	return true
}

//是否能看见目标 目标可能隐身 本单位可能能看见隐身单位
func (this *Unit) CanSeeTarget(target *Unit) bool {
	if target == nil {
		return false
	}
	//目标隐身
	if target.Invisible == 1 {
		//目标是敌人
		if this.CheckIsEnemy(target) == true && this.CanSeeInvisible != 1 && target.InvisibleBeSee != 1 {
			//log.Info("----no see")
			return false
		}

		//log.Info("----Invisible")
	}
	//大师级隐身 不会被看见 (分身的无敌和其他的blink躲弹道) 1:是 2:否
	if target.MasterInvisible == 1 {
		if this.CheckIsEnemy(target) == true {
			return false
		}
	}
	return true
}

//单位是否消失 (单位离线 单位死亡 )
func (this *Unit) IsDisappear() bool {

	if this.IsDelete == true || this.IsDeath == 1 || this.InScene == nil {
		return true
	}

	return false
}

//目标是否在攻击范围内
func (this *Unit) IsInAttackRange(target *Unit) bool {
	//this.Body.Position.X
	if target == nil {
		return false
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()
	if targetdis <= float64(this.AttackRange) {
		return true
	}

	return false
}

//设置动画状态
func (this *Unit) SetAnimotorState(anistate int32) {
	this.AnimotorState = anistate

}

//设置单位状态
func (this *Unit) SetState(state UnitState) {
	this.State = state

}

//命令操作相关
type UnitCmd struct {
	//移动命令
	MoveCmdData *protomsg.CS_PlayerMove
	//攻击命令
	AttackCmdData *protomsg.CS_PlayerAttack
	//攻击目标
	AttackCmdDataTarget *Unit
	//创建攻击命令时的时间
	AttackCmdDataTime float64
	//技能命令
	SkillCmdData *protomsg.CS_PlayerSkill

	//升级技能命令
	UpgradeSkillData *protomsg.CS_PlayerUpgradeSkill

	//切换攻击技能命令
	ChangeAttackModeData *protomsg.CS_ChangeAttackMode
}

//-----------------------技能命令--------------------
//是否有技能命令
func (this *Unit) HaveSkillCmd() bool {

	if this.SkillCmdData != nil {
		return true
	}
	return false
}

//获取技能施法距离
func (this *Unit) GetSkillRange(skillid int32) float64 {
	skillrange := float64(0)
	skilldata, ok := this.GetSkillFromTypeID(skillid)
	if ok == false {
		return skillrange
	}

	skillrange = float64(skilldata.CastRange) + float64(this.AddedMagicRange)

	return skillrange
}

//检查目标释放在技能施法范围内
func (this *Unit) IsInSkillRange(data *protomsg.CS_PlayerSkill) bool {
	//检查本单位是否有这个技能
	skilldata, ok := this.GetSkillFromTypeID(data.SkillID)
	if ok == false {
		return false
	}
	//技能施法目标类型 为 单位
	if skilldata.CastTargetType == 2 {
		target := this.InScene.FindUnitByID(data.TargetUnitID)

		dir := vec2d.Sub(this.Body.Position, target.Body.Position)

		targetdis := dir.Length()
		if targetdis > float64(skilldata.CastRange)+float64(this.AddedMagicRange) {
			return false
		}
	} else if skilldata.CastTargetType == 3 { //以目的点为施法目标
		targetpos := vec2d.Vec2{X: float64(data.X), Y: float64(data.Y)}

		dir := vec2d.Sub(this.Body.Position, targetpos)

		targetdis := dir.Length()
		if targetdis > float64(skilldata.CastRange)+float64(this.AddedMagicRange) {
			return false
		}
	}

	return true

}

//检查阵营是否满足条件
func (this *Unit) CheckUnitTargetCamp(camp int32) bool {
	if camp != 5 {
		if this.UnitType != camp {
			return false
		}
		if camp == 1 && this.IsMirrorImage == 1 {
			return false
		}
	}

	return true
}

//检查目标单位关系是否满足条件
func (this *Unit) CheckUnitTargetTeam(target *Unit, team int32) bool {
	//与目标单位关系 1:友方  2:敌方 3:友方敌方都行包括自己  4:友方敌方都行不包括自己 5:自己 10:除自己外的其他 20 自己控制的单位(不包括自己)
	isEnemy := this.CheckIsEnemy(target)
	if team == 1 {
		if isEnemy == true {
			return false
		}
	} else if team == 2 {
		if isEnemy == false {
			return false
		}
	} else if team == 4 || team == 10 {
		if this == target {
			return false
		}
	} else if team == 5 {
		if this != target {
			return false
		}
	} else if team == 20 {
		if this.MyPlayer != target.MyPlayer || this == target {
			return false
		}
	}
	return true
}

//检查技能是否可以对目标释放
func (this *Unit) CheckCastSkillTarget(target *Unit, skilldata *Skill) bool {
	//int32 UnitTargetTeam = 8;//目标单位关系 1:友方  2:敌方 3:友方敌方都行
	//int32 UnitTargetCamp = 9;//目标单位阵营 (1:英雄 2:普通单位 3:远古 4:boss) 5:都行
	if target == nil {
		return false
	}
	//目标消失
	if target.IsDisappear() == true || this.CanSeeTarget(target) == false {
		return false
	}
	//
	if target.CheckUnitTargetCamp(skilldata.UnitTargetCamp) == false {
		return false
	}

	isEnemy := this.CheckIsEnemy(target)
	//目标技能免疫 且目标是敌人
	if isEnemy == true {
		if target.MagicImmune == 1 {
			if skilldata.NoCareMagicImmune == 2 {
				return false
			}
		}
	}
	if this.CheckUnitTargetTeam(target, skilldata.UnitTargetTeam) == false {
		return false
	}
	//	//与目标单位的关系
	//	if skilldata.UnitTargetTeam == 1 {
	//		if isEnemy == true {
	//			return false
	//		}
	//	} else if skilldata.UnitTargetTeam == 2 {
	//		if isEnemy == false {
	//			return false
	//		}
	//	}

	return true
	//if skilldata.UnitTargetCamp
}

//检查额外触发条件
func (this *Unit) CheckTriggerOtherRule(rule int32, param string) bool {
	if rule <= 0 {
		return true
	}
	switch rule {
	case 1, 2: //1:表示范围内地方英雄不超过几个
		{
			param := utils.GetFloat32FromString3(param, ":")
			if len(param) < 2 {
				return false
			}
			allunit := this.InScene.FindVisibleUnitsByPos(this.Body.Position)
			count := int32(0)
			for _, v := range allunit {
				if rule == 1 && v.UnitType != 1 {
					continue
				}
				if v.IsDisappear() == false && this.CheckIsEnemy(v) == true {
					dis := float32(vec2d.Distanse(this.Body.Position, v.Body.Position))
					if dis <= param[0] {
						count++
					}
				}

				if count >= int32(param[1]) {
					return false
				}
			}

			return true

		}
	case 3: //需要有特定buff
		{
			param := utils.GetInt32FromString3(param, ":")
			if len(param) < 1 {
				return false
			}

			buff := this.GetBuff(param[0])
			if buff == nil {
				return false
			}
			return true
		}
	default:
	}

	return true
}

//检查被动触发技能 从 有攻击动画的技能
func (this *Unit) GetTriggerAttackFromAttackAnim() []int32 {
	//AttackAnim
	var re = make([]int32, 0)
	for _, v := range this.Skills {
		//CastType              int32   // 施法类型:  1:主动技能  2:被动技能
		//TriggerTime int32 //触发时间 0:表示不触发 1:攻击时 2:被攻击时
		//主动技能
		if v.CastType == 2 && v.TriggerTime == 1 && v.AttackAnim > 0 {
			//检查cd 魔法消耗
			if v.CheckCDTime() {
				//检查 触发概率 和额外条件
				if utils.CheckRandom(v.TriggerProbability) && this.CheckTriggerOtherRule(v.TriggerOtherRule, v.TriggerOtherRuleParam) {
					re = append(re, v.TypeID)
				}
			}
		}
	}
	return re
}

//刷新技能CD 检查道具同CD技能
func (this *Unit) FreshCDTime(skill *Skill, time float32) {

	if skill == nil {
		return
	}
	skill.FreshCDTime(time)

	//刷新同种的道具技能CD
	for _, v := range this.ItemSkills {
		if skill.TypeID == v.TypeID && v != skill {
			v.SameCD(skill)
		}
	}
}

func (this *Unit) CheckTriggerAttackOneSkill(b *Bullet, animattack []int32, v *Skill) {
	if v == nil || v.Level <= 0 {
		return
	}
	//CastType              int32   // 施法类型:  1:主动技能  2:被动技能
	//TriggerTime int32 //触发时间 0:表示不触发 1:攻击时 2:被攻击时
	//主动技能
	if v.CastType == 2 && v.TriggerTime == 1 {
		//animattack
		isTrigger := false
		if v.AttackAnim > 0 {
			for _, v1 := range animattack {
				if v.TypeID == v1 {
					isTrigger = true
				}
			}
		} else {
			if v.CheckCDTime() {
				//检查 触发概率 和额外条件
				//log.Info("CheckRandom %f", v.TriggerProbability)
				if utils.CheckRandom(v.TriggerProbability) && this.CheckTriggerOtherRule(v.TriggerOtherRule, v.TriggerOtherRuleParam) {
					isTrigger = true
				}
			}
		}

		//检查cd 魔法消耗
		if isTrigger == true {
			//检查 触发概率 和额外条件
			//if utils.CheckRandom(v.TriggerProbability) && this.CheckTriggerOtherRule(v.TriggerOtherRule, v.TriggerOtherRuleParam) {
			//触发
			//添加自己的buff
			this.AddBuffFromStr(v.MyBuff, v.Level, this)
			//添加自己的halo
			this.AddHaloFromStr(v.MyHalo, v.Level, nil)
			//--
			b.SetProjectileMode(v.BulletModeType, v.BulletSpeed)
			//暴击
			b.SetCrit(v.TriggerCrit)
			b.AddNoCareDodge(v.NoCareDodge)
			b.AddDoHurtPhysicalAmaorCV(v.PhysicalAmaorCV)
			//额外伤害
			if v.HurtValue > 0 {
				//技能增强
				if v.HurtType == 2 {
					hurtvalue := (v.HurtValue + int32(float32(v.HurtValue)*this.MagicScale))
					b.AddOtherHurt(HurtInfo{HurtType: v.HurtType, HurtValue: hurtvalue})

					//log.Info("add hurtvalue %d", hurtvalue)
				} else {
					b.AddOtherHurt(HurtInfo{HurtType: v.HurtType, HurtValue: v.HurtValue})
				}
			}
			b.MagicValueHurt2PhisicHurtCR = v.MagicValueHurt2PhisicHurtCR
			//特殊情况处理
			this.DoSkillException(v, b.DestUnit, b)
			//弹射
			b.SetEjection(v.EjectionCount, v.EjectionRange, v.EjectionDecay, v.EjectionRepeat)

			//召唤信息
			b.BulletCallUnitInfo = BulletCallUnitInfo{v.CallUnitInfo, v.Level}
			//目标buff
			b.AddTargetBuff(v.TargetBuff, v.Level)
			b.AddTargetHalo(v.TargetHalo, v.Level)
			//强制移动
			//if v.ForceMoveType == 1 {
			b.SetForceMove(v.ForceMoveTime, v.ForceMoveSpeedSize, v.ForceMoveLevel, v.ForceMoveType, v.ForceMoveBuff)
			//}
			b.PhysicalHurtAddHP += v.PhysicalHurtAddHP
			b.MagicHurtAddHP += v.MagicHurtAddHP

			cdtime := v.Cooldown - this.MagicCD*v.Cooldown
			//v.FreshCDTime(cdtime)
			this.FreshCDTime(v, cdtime)

			//}
		}
	} else if v.CastType == 1 && v.CastTargetType == 4 && v.AttackAutoActive == 1 {
		//主动技能 攻击时自动释放的攻击特效
		if v.CheckCDTime() == false {
			return
		}
		//魔法不足
		shouldmp := v.GetManaCost() - int32(this.ManaCost*float32(v.GetManaCost()))
		if shouldmp > int32(this.MP) {
			return
		}
		if this.SkillEnable != 1 {
			return
		}
		if b.DestUnit == nil {
			return
		}
		if b.DestUnit.MagicImmune == 1 {
			if v.NoCareMagicImmune == 2 {
				return
			}
		}
		//目标buff
		b.AddTargetBuff(v.TargetBuff, v.Level)
		b.AddTargetHalo(v.TargetHalo, v.Level)

		b.SetProjectileMode(v.BulletModeType, v.BulletSpeed)
		//强制移动
		//if v.ForceMoveType == 1 {
		b.SetForceMove(v.ForceMoveTime, v.ForceMoveSpeedSize, v.ForceMoveLevel, v.ForceMoveType, v.ForceMoveBuff)
		//}

		b.PhysicalHurtAddHP += v.PhysicalHurtAddHP
		b.MagicHurtAddHP += v.MagicHurtAddHP

		//消耗 CD
		this.ChangeMP(float32(-shouldmp))

		cdtime := v.Cooldown - this.MagicCD*v.Cooldown
		//v.FreshCDTime(cdtime)
		this.FreshCDTime(v, cdtime)

	}
}

//检查攻击 触发攻击特效
func (this *Unit) CheckTriggerAttackSkill(b *Bullet, animattack []int32) {
	//溅射buff
	for _, v := range this.Buffs {
		if len(v) > 0 {
			for _, v1 := range v {
				if v1.SpurtingRadius > 0 {
					//溅射相关
					//				log.Info("--SpurtingRadius---:%f:%f:%f:%d",
					//					v.SpurtingHurtRatio, v.SpurtingRadius, v.SpurtingRadian, v.SpurtingNoCareMagicImmune)

					b.AddSputtering(BulletSputtering{
						HurtRatio:         v1.SpurtingHurtRatio,
						Radius:            v1.SpurtingRadius,
						Radian:            v1.SpurtingRadian,
						NoCareMagicImmune: v1.SpurtingNoCareMagicImmune})
					if len(v1.SpurtingBulletModeType) > 0 {
						b.SetProjectileMode(v1.SpurtingBulletModeType, 0)
					}

				}
			}

		}
	}
	checkindex := make(map[int32]int32)
	for _, v := range this.Skills {
		if v.NoReCheckTriggerIndex != 0 {
			if _, ok := checkindex[v.NoReCheckTriggerIndex]; ok {
				//log.Info("recheck")
				continue
			}
			checkindex[v.NoReCheckTriggerIndex] = v.NoReCheckTriggerIndex
		}

		this.CheckTriggerAttackOneSkill(b, animattack, v)
	}
	for _, v := range this.ItemSkills {
		if v.NoReCheckTriggerIndex != 0 {
			if _, ok := checkindex[v.NoReCheckTriggerIndex]; ok {
				//log.Info("recheck")
				continue
			}
			//log.Info("recheck----: %d", v.NoReCheckTriggerIndex)
			checkindex[v.NoReCheckTriggerIndex] = v.NoReCheckTriggerIndex
		}
		this.CheckTriggerAttackOneSkill(b, animattack, v)
	}
}

//技能特殊处理
func (this *Unit) DoSkillException(skilldata *Skill, targetunit *Unit, b *Bullet) {
	if skilldata.Exception == 0 {
		return
	}
	switch skilldata.Exception {
	case 1: //1:混沌间隙的目标和自己的瞬移
		{
			param := utils.GetFloat32FromString3(skilldata.ExceptionParam, ":")
			if len(param) <= 0 {
				return
			}
			//targetunit := this.InScene.FindUnitByID(data.TargetUnitID)
			if targetunit == nil || targetunit.IsDisappear() {
				return
			}
			dis := vec2d.Sub(this.Body.Position, targetunit.Body.Position)
			if dis.Length() <= float64(param[0]) {
				targetunit.Body.BlinkToPos(this.Body.Position, 0)
				this.Body.BlinkToPos(targetunit.Body.Position, 0)

				if this.MyPlayer != nil {
					otherunit := this.MyPlayer.OtherUnit.Items()
					for _, v := range otherunit {
						if v != nil && v.(*Unit).Body != nil {
							v.(*Unit).Body.BlinkToPos(targetunit.Body.Position, float64(utils.GetRandomFloat(float32(180))))
							acd := &protomsg.CS_PlayerAttack{}
							acd.TargetUnitID = targetunit.ID
							v.(*Unit).AttackCmd(acd)
						}

					}
				}

			} else {
				dis.Normalize()
				dis.MulToFloat64(float64(param[0]))
				dis.Collect(&targetunit.Body.Position)

				targetunit.Body.BlinkToPos(dis, 0)
				this.Body.BlinkToPos(dis, 0)

				if this.MyPlayer != nil {
					otherunit := this.MyPlayer.OtherUnit.Items()
					for _, v := range otherunit {
						if v != nil && v.(*Unit).Body != nil {
							v.(*Unit).Body.BlinkToPos(dis, float64(utils.GetRandomFloat(float32(180))))
							acd := &protomsg.CS_PlayerAttack{}
							acd.TargetUnitID = targetunit.ID
							v.(*Unit).AttackCmd(acd)
						}

					}
				}
			}
		}
	case 2: //熊战士的怒意狂击
		{
			if b == nil {
				return
			}
			param := utils.GetInt32FromString3(skilldata.ExceptionParam, ":")
			if len(param) <= 0 {
				return
			}
			if targetunit == nil || targetunit.IsDisappear() {
				return
			}
			//怒意狂击
			buff := targetunit.GetBuff(param[0])
			if buff == nil {
				b.OtherHurt[0].HurtValue *= 0
				return
			}
			//大招的倍率伤害
			beilv := float32(1)
			if len(param) >= 2 {
				mybigbuff := this.GetBuff(param[1])
				if mybigbuff != nil {
					param1 := utils.GetFloat32FromString3(mybigbuff.ExceptionParam, ":")
					if len(param1) >= 1 {
						beilv = param1[0]
					}
				}
			}

			if len(b.OtherHurt) > 0 {
				b.OtherHurt[0].HurtValue = int32(float32(b.OtherHurt[0].HurtValue) * float32(buff.TagNum) * beilv)

			}
		}
	case 4: //帕克幻象发球
		{
			if b == nil {
				return
			}
			//把子弹ID记录到指定技能里
			param := utils.GetInt32FromString3(skilldata.ExceptionParam, ":")
			if len(param) <= 0 {
				return
			}
			skilldata1, ok := this.Skills[param[0]]
			if ok == false {
				return
			}
			skilldata1.Param1 = b.ID

		}
	case 8: //瘟疫法师歇心光环
		{
			//对击杀英雄是 增加5个buff
			if targetunit == nil || targetunit.UnitType != 1 {
				return
			}
			param := utils.GetInt32FromString3(skilldata.ExceptionParam, ":")
			if len(param) < 2 {
				return
			}
			count := param[0]
			buffid := strconv.Itoa(int(param[1]))
			for i := int32(0); i < count; i++ {
				this.AddBuffFromStr(buffid, skilldata.Level, this)
			}
		}
	default:
	}
}

func (this *Unit) DoForceMove(skilldata *Skill, targetpos vec2d.Vec2) bool {
	//自己强制移动到指定位置
	if skilldata.ForceMoveType == 2 { //自己强制移动到指定位置
		dir := vec2d.Sub(targetpos, this.Body.Position)
		time := float32(dir.Length()) / skilldata.ForceMoveSpeedSize
		dir.Normalize()
		dir.MulToFloat64(float64(skilldata.ForceMoveSpeedSize))
		this.SetForceMove(time, dir, skilldata.ForceMoveLevel, float32(0))
		//更改buff时间
		if len(skilldata.ForceMoveBuff) > 0 {
			buffs := this.AddBuffFromStr(skilldata.ForceMoveBuff, skilldata.Level, this)
			for _, v := range buffs {
				v.RemainTime = time
				v.Time = time
			}
		}
	} else if skilldata.ForceMoveType == 3 { //小小的投掷
		//查找3米内的随机一个单位 如果找不到则失败
		var touzhiunit *Unit = nil
		allunit := this.InScene.FindVisibleUnits(this)

		minDis := float32(3.0)
		for _, v := range allunit {
			//魔免
			if v.MagicImmune == 1 || v.IsDisappear() || v == this {
				continue
			}
			dis := float32(vec2d.Distanse(this.Body.Position, v.Body.Position))
			if dis <= minDis {
				touzhiunit = v
				minDis = dis
				//break
			}
		}
		if touzhiunit == nil {
			return false
		}

		dir := vec2d.Sub(targetpos, touzhiunit.Body.Position)
		if dir.Length() <= 0 {
			dir.X = dir.X + 0.1
		}
		speedsize := float32(dir.Length()) / skilldata.ForceMoveTime
		dir.Normalize()
		dir.MulToFloat64(float64(speedsize))
		touzhiunit.SetForceMove(skilldata.ForceMoveTime, dir, skilldata.ForceMoveLevel, float32(10))
		//更改buff时间  添加投掷期间buff
		if len(skilldata.ForceMoveBuff) > 0 {
			buffs := touzhiunit.AddBuffFromStr(skilldata.ForceMoveBuff, skilldata.Level, this)
			for _, v := range buffs {
				v.RemainTime = skilldata.ForceMoveTime
				v.Time = skilldata.ForceMoveTime
			}
		}
		//添加额外buff和halo
		param := utils.GetStringFromString3(skilldata.ExceptionParam, ":")
		if len(param) >= 2 {
			touzhiunit.AddBuffFromStr(param[0], skilldata.Level, this)
			halos := touzhiunit.AddHaloFromStr(param[1], skilldata.Level, nil)
			for _, v := range halos {
				v.CastUnit = this
			}
		}

	}

	return true
}

//检查击杀 触发特效
func (this *Unit) CheckTriggerKillerSkill(target *Unit) {

	//禁止被动
	if this.NoPassiveSkill == 1 {
		return
	}

	for _, v := range this.Skills {
		//CastType              int32   // 施法类型:  1:主动技能  2:被动技能
		//TriggerTime int32 //触发时间 0:表示不触发 1:攻击时 2:被攻击时 3:击杀单位时
		//主动技能
		if v.CastType == 2 && v.TriggerTime == 3 {
			//检查cd 魔法消耗
			if v.CheckCDTime() {
				//检查 触发概率 和额外条件
				if utils.CheckRandom(v.TriggerProbability) && this.CheckTriggerOtherRule(v.TriggerOtherRule, v.TriggerOtherRuleParam) {

					skilldata := v
					data := &protomsg.CS_PlayerSkill{}
					data.TargetUnitID = target.ID
					data.SkillID = v.TypeID
					data.X = float32(target.Body.Position.X)
					data.Y = float32(target.Body.Position.Y)
					//驱散自己的buff
					this.ClearBuffForTarget(this, skilldata.MyClearLevel)

					//MyBuff
					buffs := this.AddBuffFromStr(skilldata.MyBuff, skilldata.Level, this)
					for _, v := range buffs {
						v.UseableUnitID = data.TargetUnitID
					}
					//MyHalo
					this.AddHaloFromStr(skilldata.MyHalo, skilldata.Level, nil)

					//创建子弹
					bullets := skilldata.CreateBullet(this, data)
					if len(bullets) > 0 {
						for _, v := range bullets {
							if skilldata.TriggerAttackEffect == 1 {
								this.CheckTriggerAttackSkill(v, make([]int32, 0))
							}
							this.AddBullet(v)
						}

					}

					//加血
					if skilldata.AddHPTarget == 1 {
						this.DoAddHP(skilldata.AddHPType, skilldata.AddHPValue)
					}
					//特殊处理
					targetunit := this.InScene.FindUnitByID(data.TargetUnitID)
					if len(bullets) > 0 {
						this.DoSkillException(skilldata, targetunit, bullets[0])
					} else {
						this.DoSkillException(skilldata, targetunit, nil)
					}

					//消耗 CD
					namacost := skilldata.GetManaCost() - int32(this.ManaCost*float32(skilldata.GetManaCost()))
					this.ChangeMP(float32(-namacost))

					cdtime := skilldata.Cooldown - this.MagicCD*skilldata.Cooldown
					//skilldata.FreshCDTime(cdtime)
					this.FreshCDTime(skilldata, cdtime)

				}
			}
		}
	}
}

//检查被攻击 触发攻击特效
func (this *Unit) CheckTriggerBeAttackSkill(target *Unit) {

	//道具技能 受伤打断
	for _, v := range this.ItemSkills {
		v.DoBeHurt()
	}

	//禁止被动
	if this.NoPassiveSkill == 1 {
		return
	}

	for _, v := range this.Skills {
		if v.Level <= 0 {
			continue
		}
		//CastType              int32   // 施法类型:  1:主动技能  2:被动技能
		//TriggerTime int32 //触发时间 0:表示不触发 1:攻击时 2:被攻击时
		//主动技能
		if v.CastType == 2 && v.TriggerTime == 2 {
			//检查cd 魔法消耗
			if v.CheckCDTime() {
				//检查 触发概率 和额外条件
				if utils.CheckRandom(v.TriggerProbability) && this.CheckTriggerOtherRule(v.TriggerOtherRule, v.TriggerOtherRuleParam) {

					skilldata := v
					data := &protomsg.CS_PlayerSkill{}
					data.TargetUnitID = target.ID
					data.SkillID = v.TypeID
					data.X = float32(target.Body.Position.X)
					data.Y = float32(target.Body.Position.Y)
					//驱散自己的buff
					this.ClearBuffForTarget(this, skilldata.MyClearLevel)

					//MyBuff
					buffs := this.AddBuffFromStr(skilldata.MyBuff, skilldata.Level, this)
					for _, v := range buffs {
						v.UseableUnitID = data.TargetUnitID
					}
					//MyHalo
					this.AddHaloFromStr(skilldata.MyHalo, skilldata.Level, nil)

					//创建子弹
					bullets := skilldata.CreateBullet(this, data)
					if len(bullets) > 0 {
						for _, v := range bullets {
							if skilldata.TriggerAttackEffect == 1 {
								this.CheckTriggerAttackSkill(v, make([]int32, 0))
							}
							this.AddBullet(v)
						}

					}

					//加血
					if skilldata.AddHPTarget == 1 {
						this.DoAddHP(skilldata.AddHPType, skilldata.AddHPValue)
					}
					//特殊处理
					targetunit := this.InScene.FindUnitByID(data.TargetUnitID)
					if len(bullets) > 0 {
						this.DoSkillException(skilldata, targetunit, bullets[0])
					} else {
						this.DoSkillException(skilldata, targetunit, nil)
					}

					//消耗 CD
					namacost := skilldata.GetManaCost() - int32(this.ManaCost*float32(skilldata.GetManaCost()))
					this.ChangeMP(float32(-namacost))

					cdtime := skilldata.Cooldown - this.MagicCD*skilldata.Cooldown
					//skilldata.FreshCDTime(cdtime)
					this.FreshCDTime(skilldata, cdtime)

				}
			}
		}
	}
}

//使用技能 创建子弹
func (this *Unit) DoSkill(data *protomsg.CS_PlayerSkill, targetpos vec2d.Vec2) {
	if data == nil {
		return
	}
	//检查本单位是否有这个技能
	skilldata, ok := this.GetSkillFromTypeID(data.SkillID)
	if ok == false {
		return
	}
	if this.DoForceMove(skilldata, targetpos) == false {
		//删除技能命令
		this.StopSkillCmd()
		return
	}

	//驱散自己的buff
	this.ClearBuffForTarget(this, skilldata.MyClearLevel)

	//MyBuff
	buffs := this.AddBuffFromStr(skilldata.MyBuff, skilldata.Level, this)
	for _, v := range buffs {
		v.UseableUnitID = data.TargetUnitID
	}
	//MyHalo
	this.AddHaloFromStr(skilldata.MyHalo, skilldata.Level, nil)

	//BlinkToTarget
	if skilldata.BlinkToTarget == 1 {
		if skilldata.CastTargetType == 1 {
			//对自己施法 分身的时候使用
			this.Body.BlinkToPos(vec2d.Vec2{this.Body.Position.X + float64(utils.GetRandomFloat(2)),
				this.Body.Position.Y + float64(utils.GetRandomFloat(2))}, 0)
		} else {
			log.Info("blinkpos:%v", targetpos)
			this.Body.BlinkToPos(targetpos, 0)
		}

	} else if skilldata.BlinkToTarget == 3 { //blink到子弹参数位置
		if skilldata.Param1 > 0 {
			oldbullet := this.InScene.GetBulletByID(skilldata.Param1)
			if oldbullet != nil {
				this.Body.BlinkToPos(vec2d.Vec2{X: oldbullet.Position.X, Y: oldbullet.Position.Y}, 0)
				oldbullet.DestPos = oldbullet.Position
			}
		}
	}

	//加血
	if skilldata.AddHPTarget == 1 {
		this.DoAddHP(skilldata.AddHPType, skilldata.AddHPValue)
	}
	targetunit := this.InScene.FindUnitByID(data.TargetUnitID)

	//创建子弹
	bullets := skilldata.CreateBullet(this, data)
	if len(bullets) > 0 {
		for _, v := range bullets {
			if skilldata.TriggerAttackEffect == 1 {
				this.CheckTriggerAttackSkill(v, make([]int32, 0))
			}
			this.AddBullet(v)
		}

		this.DoSkillException(skilldata, targetunit, bullets[0])

	} else {

	}
	//特殊处理

	//检查关联
	if skilldata.VisibleRelationSkillID > 0 && skilldata.UseToHide == 1 {
		skilldata1, ok1 := this.Skills[skilldata.VisibleRelationSkillID]
		if ok1 {
			skilldata.SetVisible(2)
			skilldata1.SetVisible(1)
		}
	}

	//消耗 CD
	namacost := skilldata.GetManaCost() - int32(this.ManaCost*float32(skilldata.GetManaCost()))
	this.ChangeMP(float32(-namacost))

	cdtime := skilldata.Cooldown - this.MagicCD*skilldata.Cooldown
	//skilldata.FreshCDTime(cdtime)
	this.FreshCDTime(skilldata, cdtime)

	//删除技能命令
	this.StopSkillCmd()

	//如果目标是敌人 则自动攻击
	//targetunit := this.InScene.FindUnitByID(data.TargetUnitID)
	if targetunit != nil {
		if this.CheckIsEnemy(targetunit) == true {
			acd := &protomsg.CS_PlayerAttack{}
			acd.TargetUnitID = data.TargetUnitID
			this.AttackCmd(acd)
		}
	}

}
func (this *Unit) SendNoticeWord(typeid int32) {
	if this.MyPlayer != nil {
		this.MyPlayer.SendNoticeWordToClient(typeid)
	}
}

//检查是否能使用技能
func (this *Unit) UseSkillEnable(data *protomsg.CS_PlayerSkill) bool {

	if this.IsDisappear() == true {
		return false
	}

	//检查本单位是否有这个技能
	skilldata, ok := this.GetSkillFromTypeID(data.SkillID)
	if ok == false {
		return false
	}
	//英雄技能禁用
	if this.SkillEnable == 2 && skilldata.IsItemSkill != 1 {
		return false
	}
	//道具技能禁用
	if this.ItemEnable == 2 && skilldata.IsItemSkill == 1 {
		return false
	}

	//技能等级
	if skilldata.Level <= 0 {
		return false
	}

	//不是主动技能
	if skilldata.CastType != 1 {
		return false
	}
	//cd中
	if skilldata.CheckCDTime() == false {
		this.SendNoticeWord(2)
		return false
	}
	//魔法不足
	shouldmp := skilldata.GetManaCost() - int32(this.ManaCost*float32(skilldata.GetManaCost()))
	if shouldmp > int32(this.MP) {
		this.SendNoticeWord(1)
		return false
	}

	//技能施法目标类型 为 单位
	if skilldata.CastTargetType == 2 {
		target := this.InScene.FindUnitByID(data.TargetUnitID)
		if this.CheckCastSkillTarget(target, skilldata) == false {
			return false
		}
	}
	//施法距离就不判断了 如果距离不够单位自己移动过去

	return true
}

//玩家操作技能行为命令
func (this *Unit) PlayerControl_SkillCmd(data *protomsg.CS_PlayerSkill) {
	if this.PlayerControlEnable != 1 {
		return
	}
	this.SkillCmd(data)
}

//切换攻击模式命令
func (this *Unit) ChangeAttackMode(data *protomsg.CS_ChangeAttackMode) {
	this.ChangeAttackModeData = data
}
func (this *Unit) DoChangeAttackMode() {
	if this.ChangeAttackModeData != nil {

		this.AttackMode = this.ChangeAttackModeData.AttackMode

		this.ChangeAttackModeData = nil
	}
}

//玩家操作技能行为命令
func (this *Unit) UpgradeSkill(data *protomsg.CS_PlayerUpgradeSkill) {
	this.UpgradeSkillData = data
}

func (this *Unit) DoUpgradeSkill() {
	if this.UpgradeSkillData != nil {

		skilldata, ok := this.Skills[this.UpgradeSkillData.TypeID]
		if !ok {
			this.UpgradeSkillData = nil
			return
		}

		allLevel := int32(0)
		for _, v := range this.Skills {
			allLevel += v.Level - v.InitLevel
		}
		if allLevel < this.Level {
			nextlevel_needunitlevel := skilldata.RequiredLevel + skilldata.LevelsBetweenUpgrades*skilldata.Level
			if nextlevel_needunitlevel <= this.Level && skilldata.Level < skilldata.MaxLevel {
				//升级技能
				this.AddSkill(this.UpgradeSkillData.TypeID, skilldata.Level+1)
			}
		}

		this.UpgradeSkillData = nil
	}
}

//通过typid获取技能 包括道具技能
func (this *Unit) GetSkillFromTypeID(typeid int32) (*Skill, bool) {
	skilldata, ok := this.Skills[typeid]
	if ok == true {
		return skilldata, ok
	}
	for _, v := range this.ItemSkills {
		if v.TypeID == typeid {
			return v, true
		}
	}

	return nil, false
}

//技能行为命令
func (this *Unit) SkillCmd(data *protomsg.CS_PlayerSkill) {

	//如果是攻击特效技能(比如小黑的冰箭)
	//检查本单位是否有这个技能
	skilldata, ok := this.GetSkillFromTypeID(data.SkillID)
	if ok == false {
		return
	}
	if skilldata.CastType == 1 && skilldata.CastTargetType == 4 {
		skilldata.DoActive()
		return
	}

	//判断阵营 攻击模式 是否允许本次攻击
	if this.UseSkillEnable(data) == true {
		this.SkillCmdData = data
		//this.AttackCmdDataTarget = at
		log.Info("---------SkillCmd")
		//处理指定方向的技能的目标位置
		if skilldata.CastTargetType == 5 {
			dir := vec2d.Sub(vec2d.Vec2{float64(data.X), float64(data.Y)}, this.Body.Position)
			dir.Normalize()
			dir.MulToFloat64(float64(skilldata.CastRange + this.AddedMagicRange))
			dir.Collect(&this.Body.Position)
			this.SkillCmdData.X = float32(dir.X)
			this.SkillCmdData.Y = float32(dir.Y)
		}
	}

}

//检查技能指令的有效性
func (this *Unit) CheckSkillCmd() {

	if this.HaveSkillCmd() == false {
		return
	}

	///at := this.InScene.FindUnitByID(this.AttackCmdDataTarget.ID)
	if this.UseSkillEnable(this.SkillCmdData) == false {
		this.StopSkillCmd()
	}
}

//中断技能命令
func (this *Unit) StopSkillCmd() {
	this.SkillCmdData = nil
}

//-----------------------攻击命令--------------------
//是否有攻击命令
func (this *Unit) HaveAttackCmd() bool {

	if this.AttackCmdData != nil {
		return true
	}
	return false
}
func (this *Unit) CheckIsEnemy(target *Unit) bool {
	if target == nil {
		return false
	}
	//攻击模式(1:和平模式 2:组队模式 3:全体模式 4:阵营模式(玩家,NPC) 5:行会模式)
	if this.AttackMode == 1 {
		//阵营不同 //阵营(1:玩家 2:NPC)
		if this.Camp != target.Camp {
			return true
		} else {
			//如果目标是全体模式 也是敌人
			if target.AttackMode == 3 {
				return true
			}
		}
		return false
	} else if this.AttackMode == 3 {
		//全体模式下 不是同一个玩家控制的单位可以攻击
		if this.ControlID != target.ControlID {
			return true
		} else {
			if this.ControlID == -1 && this.ID != target.ID {
				return true
			}
			return false
		}

	}

	return false

}

//能否攻击(根据阵营,攻击模式判断与是否死亡)
func (this *Unit) CheckAttackEnable2Target(target *Unit) bool {

	if this.IsDisappear() || target.IsDisappear() || this.CanSeeTarget(target) == false || target.PhisicImmune == 1 {
		return false
	}

	//不能攻击
	if this.AttackEnable == 2 {
		//log.Info("------AttackEnable---")
		return false
	}
	//检查是否是敌人
	if this.CheckIsEnemy(target) == false {
		return false
	}

	return true

}

//玩家操作攻击行为命令
func (this *Unit) PlayerControl_AttackCmd(data *protomsg.CS_PlayerAttack) {
	if this.PlayerControlEnable != 1 {
		return
	}
	this.AttackCmd(data)
}

//攻击行为命令
func (this *Unit) AttackCmd(data *protomsg.CS_PlayerAttack) {

	at := this.InScene.FindUnitByID(data.TargetUnitID)
	if at == nil {
		return
	}
	//判断阵营 攻击模式 是否允许本次攻击
	if this.CheckAttackEnable2Target(at) == true {
		this.AttackCmdData = data
		this.AttackCmdDataTarget = at
		this.AttackCmdDataTime = utils.GetCurTimeOfSecond()
		//		if this.AttackMode == 3 {
		//			log.Info("---------AttackCmd")
		//		}

	}

}

//检查攻击指令的有效性 如果目标单位被场景删除 则无效
func (this *Unit) CheckAttackCmd() {

	if this.HaveAttackCmd() == false {
		return
	}
	if this.AttackCmdDataTarget == nil {
		return
	}
	///at := this.InScene.FindUnitByID(this.AttackCmdDataTarget.ID)
	if this.CheckAttackEnable2Target(this.AttackCmdDataTarget) == false {
		//log.Info("---------tttt:%d", this.AttackCmdDataTarget.ID)
		this.StopAttackCmd()
	}
	if utils.GetCurTimeOfSecond()-this.AttackCmdDataTime >= 0.5 && this.HaveMoveCmd() {
		this.StopAttackCmd()
		log.Info("---------StopAttackCmd:time 2")
	}
}

//中断攻击命令
func (this *Unit) StopAttackCmd() {
	this.AttackCmdData = nil
	this.AttackCmdDataTarget = nil
	if this.UnitType == 1 {
		//log.Info("---------StopAttackCmd")
	}

}

//-----------------------移动命令--------------------
//是否有移动命令
func (this *Unit) HaveMoveCmd() bool {

	if this.MoveCmdData != nil && this.MoveCmdData.IsStart == true {
		return true
	}

	return false
}

//是否可以移动
func (this *Unit) GetCanMove() bool {

	//	if this.IsDizzy == 1 || this.IsTwine == 1 || this.IsForceMove == 1 {
	//		return false
	//	}
	if this.MoveEnable == 2 {
		return false
	}

	return true
}

//设置单位朝向
func (this *Unit) SetDirection(dir vec2d.Vec2) {
	this.Body.Direction = dir
}

//玩家操作行为命令
func (this *Unit) PlayerControl_MoveCmd(data *protomsg.CS_PlayerMove) {

	//移动结束命令可执行
	if data.IsStart == false {
		this.MoveCmdData = data
		return
	}

	if this.PlayerControlEnable != 1 {
		return
	}
	this.MoveCmd(data)
}

//移动行为命令
func (this *Unit) MoveCmd(data *protomsg.CS_PlayerMove) {
	if this.MoveEnable == 2 {
		//当不能移动的时候 检查是否可以转向
		if this.TurnEnable == 1 {
			v1 := vec2d.Vec2{X: float64(this.MoveCmdData.X), Y: float64(this.MoveCmdData.Y)}
			this.SetDirection(v1)
		}
		return
	}

	if this.AttackCmdData == nil {
		this.MoveCmdData = data
		return
	}
	//检测是否要中断攻击
	if this.MoveCmdData == nil {
		log.Info("---------111111")
		this.StopAttackCmd()
		//this.MoveCmdData = data
	} else {
		if this.MoveCmdData.IsStart == false && data.IsStart == true {
			log.Info("---------2222")
			this.StopAttackCmd()
			//this.MoveCmdData = data
		}
		if this.MoveCmdData.IsStart == true && data.IsStart == true {
			//一直在移动中 如果超过2秒则中断攻击

			//			v1 := vec2d.Vec2{X: float64(this.MoveCmdData.X), Y: float64(this.MoveCmdData.Y)}
			//			v2 := vec2d.Vec2{X: float64(data.X), Y: float64(data.Y)}

			//			angle := vec2d.Angle(v1, v2)

			//			if math.Abs(angle) >= 0.4 {

			//				this.StopAttackCmd()
			//				this.MoveCmdData = data
			//				log.Info("---------angle:%f", angle)
			//			}
		}
	}

	this.MoveCmdData = data

	log.Info("---------MoveCmd")
}

//创建子弹的时候需要使用
type UnitProjectilePos struct {
	//弹道起始位置
	ProjectileStartPosition vec2d.Vector3
	//弹道结束位置
	ProjectileEndPosition vec2d.Vector3
	//	//弹道起始位置距离
	//	ProjectileStartPosDis float32
	//	//弹道起始位置高度
	//	ProjectileStartPosHeight float32
	//	//弹道结束位置距离
	//	ProjectileEndPosDis float32
	//	//弹道结束位置高度
	//	ProjectileEndPosHeight float32
}

//获取弹道起始位置
func (this *Unit) GetProjectileStartPos() vec2d.Vector3 {
	if this.Body == nil {
		return vec2d.NewVector3(0, 0, 0.5)
	}
	v3 := vec2d.Vec2{X: this.ProjectileStartPosition.X, Y: this.ProjectileStartPosition.Z}
	v3.Rotate(this.Body.Direction.Angle() - 90)

	//pos := vec2d.Add(this.Body.Position, vec2d.Mul(this.Body.Direction.GetNormalized(), float64(this.ProjectileStartPosDis)))
	pos := vec2d.Add(this.Body.Position, v3)

	//log.Info("GetProjectileStartPos---:%f---:%f", this.ProjectileStartPosDis, this.ProjectileStartPosHeight)

	//后期可能需要单位的z坐标参与计算
	return vec2d.NewVector3(pos.X, pos.Y, float64(this.ProjectileStartPosition.Y))
}

//获取弹道结束位置
func (this *Unit) GetProjectileEndPos() vec2d.Vector3 {
	if this.Body == nil {
		return vec2d.NewVector3(0, 0, 0.5)
	}
	v3 := vec2d.Vec2{X: this.ProjectileEndPosition.X, Y: this.ProjectileEndPosition.Z}
	v3.Rotate(this.Body.Direction.Angle() - 90)
	//log.Info("------GetProjectileEndPos:%v", v3)

	//pos := vec2d.Add(this.Body.Position, vec2d.Mul(this.Body.Direction.GetNormalized(), float64(this.ProjectileStartPosDis)))
	pos := vec2d.Add(this.Body.Position, v3)

	//pos := vec2d.Add(this.Body.Position, vec2d.Mul(this.Body.Direction.GetNormalized(), float64(this.ProjectileEndPosDis)))

	//后期可能需要单位的z坐标参与计算
	return vec2d.NewVector3(pos.X, pos.Y, float64(this.ProjectileEndPosition.Y))
}

//初始化hp和mp
func (this *Unit) InitHPandMP(hp float32, mp float32) {
	//满血 满蓝
	this.MAX_HP = this.BaseHP
	this.HP = int32(float32(this.MAX_HP) * hp)
	this.MAX_MP = this.BaseMP
	this.MP = (float32(this.MAX_MP) * mp)
	//log.Info("---hp:%d---mp:%d", this.HP, this.MP)
}

//--------------单位面板数据-----------
type UnitBaseProperty struct {
	//-
	AttributeStrength     float32 //当前力量属性
	AttributeIntelligence float32 //当前智力属性
	AttributeAgility      float32 //当前敏捷属性
	//------攻击---------
	AttackSpeed float32 //攻击速度
	Attack      int32   //攻击力 (基础攻击力+属性影响+buff影响)
	AttackRange float32 //攻击范围 攻击距离
	MoveSpeed   float64 //移动速度
	MagicScale  float32 //技能增强
	MPRegain    float32 //魔法恢复
	//------防御---------
	PhysicalAmaor  float32 //物理护甲(-1)
	PhysicalResist float32 //物理伤害抵挡
	MagicAmaor     float32 //魔法抗性(0.25)
	StatusAmaor    float32 //状态抗性(0)
	Dodge          float32 //闪避(0)
	HPRegain       float32 //生命恢复
	AddHPEffect    float32 //加血效果变化

	//hp---mp---
	AddHP int32 //增加血量
	AddMP int32 //增加蓝量

	NoCareDodge     float32 //无视闪避几率
	AddedMagicRange float32 //额外施法距离
	ManaCost        float32 //魔法消耗降低 (0.1)表示降低 10%
	MagicCD         float32 //技能CD降低 (0.1)表示降低 10%

	PhysicalHurtAddHP float32 //物理伤害吸血 0.1表示 增加攻击造成伤害的10%的HP
	MagicHurtAddHP    float32 //魔法伤害吸血 0.1表示 增加攻击造成伤害的10%的HP

	AllHurtCV   float32 //受到总伤害变化率 0.1表示 增加10%的总伤害 -0.1表示减少10%总伤害
	DoAllHurtCV float32 //造成总伤害变化率 0.1表示 增加10%的总伤害 -0.1表示减少10%总伤害

	RecoverHurt float32 //反弹伤害系数 0表示不反弹 1表示反弹100%
}

//------------------单位本体------------------
type UnitProperty struct {
	conf.UnitFileData //单位配置文件数据
	UnitProjectilePos

	// 当前数据
	ControlID int32 //控制者ID
	IsMain    int32 //是否是主单位 1:是  2:不是

	AnimotorState int32 //动画状态 1:idle 2:walk 3:attack 4:skill 5:death
	AttackAnim    int32 //攻击动画
	//-------------新加----------
	AttackMode int32 //攻击模式(1:和平模式 2:组队模式 3:全体模式 4:阵营模式(玩家,NPC) 5:行会模式)

	IsDeath int32 //是否死亡(1:死亡 2:没死)
	Name    string

	NextAttackRemainTime float32 //下次攻击剩余时间

	//复合数据 会随时变动的数据 比如受buff影响攻击力降低  (每帧动态计算)
	HP               int32
	MAX_HP           int32
	MP               float32
	MAX_MP           int32
	Level            int32 //等级 会影响属性
	Experience       int32
	MaxExperience    int32
	Gold             int32   //金币
	Diamond          int32   //钻石
	GetExperienceDay string  //获取经验的日期
	RemainExperience int32   //今天还能获取的经验值
	RemainReviveTime float32 //剩余复活时间

	UnitBaseProperty

	PlayerControlEnable int32 //玩家能否操作 1:能 2:不能
	MoveEnable          int32 //能否移动 (比如 被缠绕不能移动) 1:可以 2:不可以
	TurnEnable          int32 //能否转向 (比如 被眩晕不能转向) 1:可以 2:不可以
	AttackEnable        int32 //能否攻击 (比如 被眩晕和缴械不能攻击) 1:可以 2:不可以
	SkillEnable         int32 //能否使用主动技能 (比如 被眩晕和沉默不能使用主动技能) 1:可以 2:不可以
	NoPassiveSkill      int32 //禁止被动技能 1:是 2:非
	ItemEnable          int32 //能否使用主动道具 (比如 被眩晕和禁用道具不能使用主动道具) 1:可以 2:不可以
	MagicImmune         int32 //是否技能免疫 1：是 2:不是
	PhisicImmune        int32 //是否物理攻击免疫 1:是 2:否
	MagicCDStop         int32 //技能冷却停止 1:是 2:非
	AnimotorPause       int32 //是否暂停动画 1:是 2:非

	Invisible       int32 //隐身 1:是 2:否
	InvisibleBeSee  int32 //隐身可以被看见 1:是 2:否
	CanSeeInvisible int32 //可以看见隐身 1:是 2:否
	MasterInvisible int32 //大师级隐身 不会被看见 (分身的无敌和其他的blink躲弹道) 1:是 2:否

	IsMirrorImage int32 //是否是镜像 1:是 2:不是

	//强制移动相关
	ForceMoveTime       float32    //强制移动总时间
	ForceMoveMaxHeight  float32    //强制移动最大高度
	ForceMoveRemainTime float32    //强制移动剩余时间
	ForceMoveSpeed      vec2d.Vec2 //强制移动速度 包括方向和大小
	ForceMoveLevel      int32      //强制移动等级
	Z                   float32    //z坐标

}

//时间点的伤害
type TimeAndHurt struct {
	Time      float64
	HurtValue int32
}

//存储时间点受到的伤害
func (this *Unit) SaveTimeAndHurt(hurt int32) {
	tah := TimeAndHurt{}
	tah.Time = utils.GetCurTimeOfSecond()
	tah.HurtValue = hurt
	this.TimeHurts = append(this.TimeHurts, tah)
}

//自动删除8秒前 存储的时间点受到的伤害
func (this *Unit) AutoRemoveTimeAndHurt() {
	curtime := utils.GetCurTimeOfSecond()
	deleteindex := -1 //删除的索引
	for k, v := range this.TimeHurts {
		if curtime-v.Time > 8 {
			deleteindex = k
		} else {
			break
		}
	}
	if deleteindex >= 0 {
		this.TimeHurts = this.TimeHurts[deleteindex+1:]
	}

}

//获取time之内受到的伤害
func (this *Unit) GetTimeAndHurt(t float32) int32 {
	re := int32(0)
	curtime := utils.GetCurTimeOfSecond()
	for _, v := range this.TimeHurts {
		if curtime-v.Time < float64(t) {
			re += v.HurtValue
		}
	}
	return re
}

type Unit struct {
	UnitProperty
	UnitCmd
	InScene  *Scene
	MyPlayer *Player
	ID       int32        //单位唯一ID
	Body     *cyward.Body //移动相关(位置信息) 需要设置移动速度
	State    UnitState    //逻辑状态
	AI       UnitAI

	IsDelete bool //是否被删除

	InitPosition vec2d.Vec2 //初始位置

	ItemSkills []*Skill //所有道具技能

	Skills map[int32]*Skill  //所有技能
	Buffs  map[int32][]*Buff //所有buff 同typeID下可能有多个buff

	Items []*Item //所有道具

	HaloInSkills map[int32][]int32 //来自被动技能的光环

	//记录时间点的伤害
	TimeHurts []TimeAndHurt

	//每秒钟干事 剩余时间
	EveryTimeDoRemainTime float32 //每秒钟干事 的剩余时间

	//场景中的NPC 死亡后重新创建信息
	ReCreateInfo *conf.Unit
	//场景中的NPC 死亡后掉落道具

	//发送数据部分
	ClientData    *protomsg.UnitDatas //客户端显示数据
	ClientDataSub *protomsg.UnitDatas //客户端显示差异数据
}

func (this *Unit) SetReCreateInfo(recreateinfo *conf.Unit) {
	this.ReCreateInfo = recreateinfo
}

func (this *Unit) SetAI(ai UnitAI) {
	this.AI = ai
	ai.OnStart()

}
func (this *Unit) FreshHaloInSkills() {
	for _, v := range this.Skills {
		if len(v.InitHalo) > 0 {
			//已经有此halo 就要删除之前的
			if _, ok := this.HaloInSkills[v.TypeID]; ok {
				for _, v1 := range this.HaloInSkills[v.TypeID] {
					this.InScene.RemoveHalo(v1)
				}
			}
			//---------------
			//log.Info("111111111111")
			if v.Level <= 0 {
				continue
			}

			halos := utils.GetInt32FromString2(v.InitHalo)
			re := make([]int32, 0)
			for _, v1 := range halos {
				halo := NewHalo(v1, v.Level)
				halo.SetParent(this)
				if halo != nil {
					this.InScene.AddHalo(halo)
					//log.Info("2222222222222")
				}
				re = append(re, halo.ID)
			}
			this.HaloInSkills[v.TypeID] = re

		}
	}
}

//删除道具 从装备栏
func (this *Unit) RemoveItem(index int32) {
	if int(index) >= len(this.Items) || int(index) < 0 {
		return
	}
	item := this.Items[index]

	if item != nil {
		item.Clear()
		this.Items[index] = nil
	}
}

//添加道具在装备栏
func (this *Unit) AddItem(index int32, item *Item) bool {

	if int(index) >= len(this.Items) || int(index) < 0 {
		for k, v := range this.Items {
			if v == nil {
				index = int32(k)
				break
			}
		}
	}
	if this.Items[index] == nil {
		this.Items[index] = item
		if item != nil {
			item.Clear()
			item.Add2Unit(this, index)
		}
		return true
	}

	return false

}

//添加道具技能
func (this *Unit) AddItemSkill(skill *Skill) bool {
	//this.ItemSkills[skill] = skill //所有道具技能
	if skill.ActiveUnitAcpabilities == 1 && this.AttackAcpabilities != 1 {
		log.Info("no AddItemSkill Acpabilities %d", skill.TypeID)
		return false
	}
	if skill.ActiveUnitAcpabilities == 2 && this.AttackAcpabilities != 2 {
		log.Info("no AddItemSkill Acpabilities %d", skill.TypeID)
		return false
	}

	//如果已经有此道具在身上 就同步身上道具的CD 否则同步数据库中的cd
	isInBody := false
	for _, v := range this.ItemSkills {
		if v.TypeID == skill.TypeID {
			skill.SameCD(v)
			isInBody = true
			break
		}
	}
	if isInBody == false {
		if this.MyPlayer != nil {
			iscdinfo := this.MyPlayer.GetItemSkillCDInfo(skill.TypeID)

			if iscdinfo != nil {
				log.Info("---AddItemSkill-%d  %f", skill.TypeID, iscdinfo.RemainCDTime)
				skill.ResetCDTime(iscdinfo.RemainCDTime)
			}
		}
	}

	this.ItemSkills = append(this.ItemSkills, skill)

	return true
}

//删除道具技能
func (this *Unit) RemoveItemSkill(skill *Skill) {

	if skill == nil {
		return
	}
	//log.Info("---RemoveItemSkill-%d", skill.TypeID)
	//保存CD
	if this.MyPlayer != nil {
		this.MyPlayer.SaveItemSkillCDInfo(skill)
	}

	for k, v := range this.ItemSkills {
		if v == skill {
			this.ItemSkills = append(this.ItemSkills[:k], this.ItemSkills[k+1:]...)
			return
		}
	}

}

//刷新道具的作用
//func (this *Unit) FreshUseableItem() {
//	for _, v := range this.Items {
//		if v != nil {
//			v.Clear()
//			v.Add2Unit(this)
//		}
//	}
//}

//掉落道具
func (this *Unit) DropItem() {

	drop := make([]int32, 0)
	if len(this.NPCItemDropInfo) > 0 {
		dropitems := strings.Split(this.NPCItemDropInfo, ";")
		for _, v := range dropitems {
			param := utils.GetFloat32FromString3(v, ",")
			if len(param) != 2 {
				continue
			}
			//
			if utils.CheckRandom(param[1]) {
				drop = append(drop, int32(param[0]))
			}

		}
	}
	if len(drop) > 0 {
		this.InScene.CreateSceneItems(drop, this.Body.Position)
	}
	//CreateSceneItems
}

func CreateUnit(scene *Scene, typeid int32) *Unit {

	if scene == nil {
		return nil
	}

	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	//	文件数据
	unitre.UnitFileData = *(conf.GetUnitFileData(typeid))
	unitre.Name = unitre.UnitName
	unitre.Level = 1
	leveldata := conf.GetLevelFileData(unitre.Level)
	if leveldata != nil {
		unitre.MaxExperience = leveldata.UpgradeExperience
	}

	//初始化技能被动光环
	unitre.ItemSkills = make([]*Skill, 0)  //所有道具技能
	unitre.Skills = make(map[int32]*Skill) //所有技能
	unitre.HaloInSkills = make(map[int32][]int32)
	unitre.FreshHaloInSkills()

	unitre.Init()
	//创建道具
	unitre.Items = make([]*Item, UnitEquitCount)
	unitre.InitHPandMP(1.0, 1.0)
	unitre.IsMain = 2
	//unitre.UnitType = 2 //单位类型(1:英雄 2:普通单位 3:远古 4:boss)
	unitre.ControlID = -1

	return unitre
}
func (this *Unit) AddSkill(id int32, level int32) {
	skill := NewOneSkill(id, level, this)
	if skill != nil {
		this.Skills[id] = skill
	}
	this.FreshHaloInSkills()
}

//幻象复制主体道具
func (this *Unit) CopyItem(unit *Unit) {
	if unit == nil {
		return
	}
	this.Items = make([]*Item, UnitEquitCount)
	for _, v := range unit.Items {
		if v != nil {
			this.AddItem(v.Index, NewItem(v.TypeID))
		}
	}

}
func CreateUnitByCopyUnit(unit *Unit, controlplayer *Player) *Unit {
	if unit == nil || unit.InScene == nil {
		return nil
	}

	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = unit.InScene
	unitre.MyPlayer = controlplayer

	//	文件数据
	unitre.UnitFileData = *(conf.GetUnitFileData(unit.TypeID))
	unitre.InitHPandMP(float32(unit.HP)/float32(unit.MAX_HP), float32(unit.MP)/float32(unit.MAX_MP))

	//名字 等级 经验 创建时的位置
	unitre.Name = unit.Name
	unitre.Level = unit.Level
	unitre.Experience = unit.Experience
	leveldata := conf.GetLevelFileData(unitre.Level)
	if leveldata != nil {
		unitre.MaxExperience = leveldata.UpgradeExperience
	}

	//继承被动技能
	unitre.ItemSkills = make([]*Skill, 0)  //所有道具技能
	unitre.Skills = make(map[int32]*Skill) //所有技能
	for _, v := range unit.Skills {
		if v.CastType == 2 {
			skill := NewOneSkill(v.TypeID, v.Level, unitre)
			if skill != nil {
				unitre.Skills[v.TypeID] = skill
			}
		}
	}
	//初始化技能被动光环
	unitre.HaloInSkills = make(map[int32][]int32)

	unitre.Death2RemoveTime = 0

	//初始化
	unitre.Init()
	//创建道具
	unitre.Items = make([]*Item, UnitEquitCount)
	//拷贝道具

	unitre.FreshHaloInSkills()
	unitre.IsMain = 0
	unitre.IsMirrorImage = 1
	//复制主体道具
	unitre.CopyItem(unit)

	controlplayer.AddOtherUnit(unitre)

	return unitre

}

func CreateUnitByPlayer(scene *Scene, player *Player, datas []byte) *Unit {
	if scene == nil || player == nil {
		return nil
	}
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	unitre.MyPlayer = player

	//---------db.DB_CharacterInfo
	characterinfo := db.DB_CharacterInfo{}
	utils.Bytes2Struct(datas, &characterinfo)
	player.Characterid = characterinfo.Characterid
	player.LoadBagInfoFromDB(characterinfo.BagInfo)
	player.LoadItemSkillCDFromDB(characterinfo.ItemSkillCDInfo)

	log.Info("---DB_CharacterInfo---%v", characterinfo)
	//	文件数据
	unitre.UnitFileData = *(conf.GetUnitFileData(characterinfo.Typeid))
	unitre.InitHPandMP(characterinfo.HP, characterinfo.MP)

	//名字 等级 经验 创建时的位置
	unitre.Name = characterinfo.Name
	unitre.Level = characterinfo.Level
	unitre.Experience = characterinfo.Experience
	unitre.Gold = characterinfo.Gold
	unitre.Diamond = characterinfo.Diamond
	unitre.GetExperienceDay = characterinfo.GetExperienceDay //获取经验的日期
	unitre.RemainExperience = characterinfo.RemainExperience //今天还能获取的经验值
	unitre.RemainReviveTime = characterinfo.RemainReviveTime
	unitre.InitPosition = vec2d.Vec2{float64(characterinfo.X), float64(characterinfo.Y)}

	leveldata := conf.GetLevelFileData(unitre.Level)
	if leveldata != nil {
		unitre.MaxExperience = leveldata.UpgradeExperience
	}

	//创建技能
	unitre.ItemSkills = make([]*Skill, 0) //所有道具技能
	skilldbdata := strings.Split(characterinfo.Skill, ";")
	unitre.Skills = NewUnitSkills(skilldbdata, unitre.InitSkillsInfo, unitre) //所有技能
	for _, v := range unitre.Skills {
		log.Info("-------new skill:%v", v)
	}
	//初始化技能被动光环
	unitre.HaloInSkills = make(map[int32][]int32)

	//初始化
	unitre.Init()

	//创建道具
	unitre.Items = make([]*Item, UnitEquitCount)
	//	unitre.Items[0] = NewItemFromDB(characterinfo.Item1)
	//	unitre.Items[1] = NewItemFromDB(characterinfo.Item2)
	//	unitre.Items[2] = NewItemFromDB(characterinfo.Item3)
	//	unitre.Items[3] = NewItemFromDB(characterinfo.Item4)
	//	unitre.Items[4] = NewItemFromDB(characterinfo.Item5)
	//	unitre.Items[5] = NewItemFromDB(characterinfo.Item6)
	unitre.AddItem(0, NewItemFromDB(characterinfo.Item1))
	unitre.AddItem(1, NewItemFromDB(characterinfo.Item2))
	unitre.AddItem(2, NewItemFromDB(characterinfo.Item3))
	unitre.AddItem(3, NewItemFromDB(characterinfo.Item4))
	unitre.AddItem(4, NewItemFromDB(characterinfo.Item5))
	unitre.AddItem(5, NewItemFromDB(characterinfo.Item6))
	//unitre.FreshUseableItem()

	unitre.FreshHaloInSkills()

	unitre.IsMain = 1
	unitre.ControlID = player.Uid

	log.Info("---DB_CharacterInfo---over")

	return unitre
}

//初始化
func (this *Unit) Init() {
	this.State = NewIdleState(this)

	this.AttackMode = 1 //和平攻击模式
	this.EveryTimeDoRemainTime = 1

	this.IsDeath = 2
	this.IsMirrorImage = 2

	this.ClearBuff()

	//弹道位置计算

	utils.GetFloat64FromString(this.ProjectileStartPos, &this.ProjectileStartPosition.X, &this.ProjectileStartPosition.Z, &this.ProjectileStartPosition.Y)
	utils.GetFloat64FromString(this.ProjectileEndPos, &this.ProjectileEndPosition.X, &this.ProjectileEndPosition.Z, &this.ProjectileEndPosition.Y)

	//this.TestData()
	this.TimeHurts = make([]TimeAndHurt, 0)
}

//设置强制移动相关
func (this *Unit) SetForceMove(time float32, speed vec2d.Vec2, level int32, height float32) {

	//direction.Normalize()
	log.Info("SetForceMove--%f  %v  %d %f  %d  %f", time, speed, level, height, this.ForceMoveLevel, this.ForceMoveRemainTime)

	if this.ForceMoveRemainTime <= 0 {
		this.ForceMoveRemainTime = time
		this.ForceMoveSpeed = speed
		this.ForceMoveLevel = level
		this.ForceMoveTime = time
		this.ForceMoveMaxHeight = height
	} else {
		if level >= this.ForceMoveLevel {
			this.ForceMoveRemainTime = time
			this.ForceMoveSpeed = speed
			this.ForceMoveLevel = level
			this.ForceMoveTime = time
			this.ForceMoveMaxHeight = height
		}
	}
}

//更新强制移动
func (this *Unit) UpdateForceMove(dt float64) {
	if this.ForceMoveRemainTime > 0 {
		this.ForceMoveRemainTime -= float32(dt)
		//movedelta := vec2d.Mul(this.ForceMoveSpeed.GetNormalized(),dt)
		this.Body.SpeedSize = this.ForceMoveSpeed.Length()
		this.Body.IsCollisoin = false
		this.Body.TurnDirection = false
		this.Body.CollisoinLevel = 4
		this.Body.SetMoveDir(this.ForceMoveSpeed)

		//设置高度
		bilv := this.ForceMoveRemainTime / this.ForceMoveTime
		if bilv <= 0.5 {
			this.Z = this.ForceMoveMaxHeight * bilv
		} else {
			this.Z = this.ForceMoveMaxHeight * (1 - bilv)
		}

	} else {
		//this.Z = 0
	}
}

//目前1秒钟更新一次
func (this *Unit) UpdateSkillAddBuff() {
	//------
	//技能携带的buf
	for _, v := range this.Skills {
		if v.Level <= 0 {
			continue
		}
		//被动技能
		if v.CastType == 2 {

			//没有禁止被动
			if this.NoPassiveSkill == 2 {
				buffs := this.AddBuffFromStr(v.InitBuff, v.Level, this)
				for _, v := range buffs {
					v.RemainTime = 1
				}
				//log.Info("NoPassiveSkill")

				if len(v.InitHalo) > 0 {
					if _, ok := this.HaloInSkills[v.TypeID]; ok {
						for _, v1 := range this.HaloInSkills[v.TypeID] {
							this.InScene.ForbiddenHalo(v1, false)
						}
					}
				}
			} else { //禁止被动
				if len(v.InitHalo) > 0 {
					if _, ok := this.HaloInSkills[v.TypeID]; ok {
						for _, v1 := range this.HaloInSkills[v.TypeID] {
							this.InScene.ForbiddenHalo(v1, true)
						}
					}
				}
			}
		} else {
			buffs := this.AddBuffFromStr(v.InitBuff, v.Level, this)
			for _, v := range buffs {
				v.RemainTime = 1
			}
		}

	}
}

//检查剩余获取经验值日期
func (this *Unit) CheckGetExperienceDay() {
	today := time.Now().Format("2006-01-02")
	if today != this.GetExperienceDay {
		this.GetExperienceDay = today
		leveldata := conf.GetLevelFileData(this.Level)
		if leveldata != nil {
			this.RemainExperience = leveldata.MaxExperienceOneDay
		} else {
			this.RemainExperience = 1000
		}

	}
}

//处理复活
func (this *Unit) CheckAvive(dt float32) {

	if this.RemainReviveTime <= 0 {
		this.RemainReviveTime = 0
		return
	}
	if this.HP > 0 {
		return
	}
	//处理复活
	if this.MyPlayer != nil && this == this.MyPlayer.MainUnit {
		this.RemainReviveTime -= float32(dt)
		if this.RemainReviveTime <= 0 {
			//复活
			this.DoAvive(2, 0.5, 0.5)
		}
	}
}

//复活 postype复活位置类型 1表示原地复活 2表示当前地图随机位置复活 hpmp复活状态百分比
func (this *Unit) DoAvive(postype int32, hp float32, mp float32) {
	mmp := mp
	if mmp < this.MP/float32(this.MAX_MP) {
		mmp = this.MP / float32(this.MAX_MP)
	}
	this.InitHPandMP(hp, mmp)

	if this.InScene == nil {
		return
	}
	if postype == 2 {
		////StartX	StartY	EndX	EndY
		x := utils.GetRandomFloatTwoNum(this.InScene.StartX, this.InScene.EndX)
		y := utils.GetRandomFloatTwoNum(this.InScene.StartY, this.InScene.EndY)
		targetpos := vec2d.Vec2{float64(x), float64(y)}
		this.Body.BlinkToPos(targetpos, 0)
	}
}

//
func (this *Unit) EveryTimeDo(dt float64) {

	this.EveryTimeDoRemainTime -= float32(dt)
	if this.EveryTimeDoRemainTime <= 0 {
		//do
		this.EveryTimeDoRemainTime += 1
		this.CheckAvive(1)
		if this.IsDisappear() {
			return
		}
		//每秒回血
		this.ChangeHP(int32(this.HPRegain))
		this.ChangeMP((this.MPRegain))

		this.UpdateSkillAddBuff()

		this.AutoRemoveTimeAndHurt()

		this.CheckGetExperienceDay()

		//
	}
}

func (this *Unit) ShowMiss(isshow bool) {
	if this.ClientData != nil {
		this.ClientData.IsMiss = isshow

	}
	if this.ClientDataSub != nil {
		this.ClientDataSub.IsMiss = isshow
		//		if isshow {
		//			log.Info("ShowMiss")
		//		}

	}
}

//
//更新 范围影响的buff 被动技能
func (this *Unit) PreUpdate(dt float64) {
	if this.IsDisappear() {
		return
	}
	this.NextAttackRemainTime -= float32(dt)

	this.DoUpgradeSkill()
	this.DoChangeAttackMode()

	this.ShowMiss(false)
}

//更新
func (this *Unit) Update(dt float64) {
	//设置是否有碰撞  设置移动速度 和逻辑状态

	//技能更新
	if this.MagicCDStop != 1 {
		for _, v := range this.Skills {
			v.Update(dt)
		}
	}
	//道具技能更新
	for _, v := range this.ItemSkills {
		v.Update(dt)
	}

	//更新buff
	for k, _ := range this.Buffs {
		//log.Info("----buff-----id:%d", k)
		for i := 0; i < len(this.Buffs[k]); {
			//log.Info("----buff22-----id:%d  %d %d", k, i, len(v))
			this.Buffs[k][i].Update(dt)

			if this.Buffs[k][i].IsEnd == true {
				this.Buffs[k] = append(this.Buffs[k][:i], this.Buffs[k][i+1:]...)
			} else {
				i++
			}

		}
		if len(this.Buffs[k]) <= 0 {
			delete(this.Buffs, k)

		}
		//log.Info("----buff11-----id:%d  %f  %f", k, utils.GetCurTimeOfSecond())

	}

	//计算属性值
	this.CalProperty()

	//移动核心
	//移动行为逻辑速度设置
	this.Body.SpeedSize = this.MoveSpeed

	//AI
	if this.AI != nil {
		this.AI.Update(dt)
	}
	this.CheckSkillCmd()
	this.CheckAttackCmd()

	//
	this.EveryTimeDo(dt)

	//逻辑状态更新
	this.State.OnTransform()
	this.State.Update(dt)

	//强制移动更新
	this.UpdateForceMove(dt)

	//设置碰撞等级
	if this.IsDeath == 1 {
		this.Body.IsCollisoin = false
	}
}

//AttributeStrength float32//当前力量属性
//	AttributeIntelligence float32//当前智力属性
//	AttributeAgility float32//当前敏捷属性
//计算属性(力量 智力 敏捷)
func (this *Unit) CalAttribute() {
	//基础力量+等级带来的力量成长
	this.AttributeStrength = this.AttributeBaseStrength + float32(this.Level-1)*this.AttributeStrengthGain

	this.AttributeIntelligence = this.AttributeBaseIntelligence + float32(this.Level-1)*this.AttributeIntelligenceGain

	this.AttributeAgility = this.AttributeBaseAgility + float32(this.Level-1)*this.AttributeAgilityGain

	//装备
	//技能

}

//加血
func (this *Unit) DoAddHP(addType int32, addval float32) int32 {
	re := int32(0)
	//1:以AddHPValue为固定值 2:以AddHPValue为时间 加单位在此时间内受到的伤害值
	//3:以AddHPValue为最大魔法值的比例加血（血晶石）
	if addType == 1 {
		re = this.ChangeHP(int32(addval))
	} else if addType == 2 {
		re = this.ChangeHP(0 - this.GetTimeAndHurt(addval))
	} else if addType == 3 {

		addhp := int32(float32(this.MAX_MP) * addval)
		log.Info("AddHPType---------------------:%d", addhp)
		re = this.ChangeHP(addhp)
	}
	//客户端显示
	if this.MyPlayer != nil {
		mph := &protomsg.MsgPlayerHurt{HurtUnitID: this.ID, HurtAllValue: re}
		this.MyPlayer.AddHurtValue(mph)
	}

	return re
}

//改变血量
func (this *Unit) ChangeHP(hp int32) int32 {

	//log.Info("---ChangeHP11---:%d   :%d", hp, this.HP)
	//加血增强
	if hp > 0 {
		hp = hp + int32(float32(hp)*this.AddHPEffect)
	}

	lasthp := this.HP
	this.HP += hp
	if this.HP <= 0 {
		//死亡处理
		this.HP = 0
		this.IsDeath = 1
	} else {
		//死亡处理
		this.IsDeath = 2
	}
	if this.HP >= this.MAX_HP {
		this.HP = this.MAX_HP
	}
	//log.Info("---ChangeHP---:%d   :%d", hp, this.HP)
	return this.HP - lasthp
	//log.Info("---ChangeHP---:%d   :%d", hp, this.HP)
}

//改变MP
func (this *Unit) ChangeMP(mp float32) int32 {
	lastmp := int32(this.MP)
	this.MP += mp
	if this.MP <= 0 {
		this.MP = 0
	}
	if this.MP >= float32(this.MAX_MP) {
		this.MP = float32(this.MAX_MP)
	}
	return int32(this.MP) - lastmp
}

//计算MAX_HP和MAX_MP
func (this *Unit) CalMaxHP_MaxHP(add *UnitBaseProperty) {
	AddHP := int32(0) //增加血量
	AddMP := int32(0) //增加蓝量
	if add != nil {
		AddHP = add.AddHP //增加血量
		AddMP = add.AddMP //增加蓝量
	}
	maxhp := this.BaseHP + AddHP + int32(this.AttributeStrength*conf.StrengthAddHP)
	//装备
	//技能
	//buff

	if maxhp != this.MAX_HP {

		//按百分比增减当前血量
		changehp := float32(maxhp)/float32(this.MAX_HP)*float32(this.HP) - float32(this.HP)
		//log.Info("change hp:%d-----%d---%d----%d", int32(changehp), maxhp, this.MAX_HP, this.HP)
		this.MAX_HP = maxhp
		this.ChangeHP(int32(changehp))

	}

	//MP
	maxmp := this.BaseMP + AddMP + int32(this.AttributeIntelligence*conf.IntelligenceAddMP)
	if maxmp != this.MAX_MP {

		changemp := float32(maxmp)/float32(this.MAX_MP)*float32(this.MP) - float32(this.MP)
		this.MAX_MP = maxmp
		this.ChangeMP((changemp))

	}
}

//计算攻击速度
func (this *Unit) CalAttackSpeed() {
	//基础攻击加上敏捷增加的攻击
	this.AttackSpeed = float32(this.BaseAttackSpeed) + float32(this.AttributeAgility*conf.AgilityAddAttackSpeed)
	//装备
	//技能
	//buff

}

//计算攻击力
func (this *Unit) CalAttack() {
	//主属性(1:力量 2:敏捷 3:智力)
	//基础攻击力+主属性增减攻击力
	switch this.AttributePrimary {
	case 1:
		this.Attack = this.BaseAttack + int32(this.AttributeStrength*conf.AttributePrimaryAddAttack)
		break
	case 2:
		this.Attack = this.BaseAttack + int32(this.AttributeAgility*conf.AttributePrimaryAddAttack)
		break
	case 3:
		this.Attack = this.BaseAttack + int32(this.AttributeIntelligence*conf.AttributePrimaryAddAttack)
		break
	}

	//装备
	//技能
	//buff

}

//计算攻击距离
func (this *Unit) CalAttackRange() {
	//攻击范围
	this.AttackRange = this.BaseAttackRange

	//装备
	//技能
	//buff
}

//计算移动速度
func (this *Unit) CalMoveSpeed() {
	//基础移动速度+敏捷对移动速度的提升
	agilityaddspeed := float64(this.BaseMoveSpeed) * float64(this.AttributeAgility*conf.AgilityAddMoveSpeed)
	this.MoveSpeed = float64(this.BaseMoveSpeed) + agilityaddspeed

	//装备
	//技能
	//buff

}

//计算技能增强IntelligenceAddMagicScale
func (this *Unit) CalMagicScale() {
	//通过智力计算
	this.MagicScale = float32(this.AttributeIntelligence * conf.IntelligenceAddMagicScale)
	//装备
	//技能
	//buff
}

//计算魔法回复
func (this *Unit) CalMPRegain() {
	this.MPRegain = this.BaseMPRegain + float32(this.AttributeIntelligence*conf.IntelligenceAddMPRegain)
	//装备
	//技能
	//buff
}

////------防御---------

func (this *Unit) GetBasePhysicalAmaor() float32 {
	return this.BasePhysicalAmaor + float32(this.AttributeAgility*conf.AgilityAddPhysicalAmaor)
}

//计算护甲和物理抵抗
func (this *Unit) CalPhysicalAmaor() {
	//基础护甲+敏捷增减的护甲
	this.PhysicalAmaor = this.GetBasePhysicalAmaor()

	//装备
	//技能
	//buff

	//计算物理伤害抵挡
	this.PhysicalResist = utils.UnitPhysicalAmaor2PhysicalResist(this.PhysicalAmaor)

	//	if this.UnitType == 1 {
	//		log.Info("CalPhysicalAmaor %f   %f", this.PhysicalAmaor, this.PhysicalResist)
	//	}

}

//计算魔抗
func (this *Unit) CalMagicAmaor() {
	//非线性叠加
	//基础魔抗叠加力量带来的魔抗

	strenth := float32(this.AttributeStrength * conf.StrengthAddMagicAmaor)

	//装备
	//技能
	//buff

	this.MagicAmaor = utils.NoLinerAdd(this.BaseMagicAmaor, strenth)
}

//计算状态抗性
func (this *Unit) CalStatusAmaor() {
	//非线性叠加
	statusamaor := (1 - this.BaseStatusAmaor)

	//装备
	//技能
	//buff

	this.StatusAmaor = 1 - statusamaor
}

//计算闪避
func (this *Unit) CalDodge() {
	//非线性叠加
	dodge := (1 - this.BaseDodge)

	//装备
	//技能
	//buff

	this.Dodge = (1 - dodge)

	//计算无视闪避几率
	//装备
	//技能
	//buff
	this.NoCareDodge = 0
}

//计算生命回复
func (this *Unit) CalHPRegain() {
	this.HPRegain = this.BaseHPRegain + float32(this.AttributeStrength*conf.StrengthAddHPRegain)
	//装备
	//技能
	//buff
}

//计算异常状态
func (this *Unit) CalControlState() {
	this.MoveEnable = 1   //能否移动 (比如 被缠绕不能移动) 1:可以 2:不可以
	this.TurnEnable = 1   //能否转向 (比如 被眩晕不能转向) 1:可以 2:不可以
	this.AttackEnable = 1 //能否攻击 (比如 被眩晕和缴械不能攻击) 1:可以 2:不可以
	this.SkillEnable = 1  //能否使用主动技能 (比如 被眩晕和沉默不能使用主动技能) 1:可以 2:不可以
	this.NoPassiveSkill = 2
	this.ItemEnable = 1 //能否使用主动道具 (比如 被眩晕和禁用道具不能使用主动道具) 1:可以 2:不可以
	this.MagicImmune = 2
	this.PhisicImmune = 2
	this.MagicCDStop = 2
	this.AnimotorPause = 2
	this.PlayerControlEnable = 1

	this.AddedMagicRange = 0 //额外施法距离
	this.ManaCost = 0        //魔法消耗降低
	this.AddHPEffect = 0     //
	this.MagicCD = 0         //技能CD降低
	this.Invisible = 2       //隐身   否
	this.InvisibleBeSee = 2
	this.CanSeeInvisible = 2
	this.MasterInvisible = 2

	this.PhysicalHurtAddHP = 0
	this.MagicHurtAddHP = 0

	this.AllHurtCV = 0
	this.DoAllHurtCV = 0

	this.RecoverHurt = 0

	this.Body.IsCollisoin = true
	this.Body.TurnDirection = true
	this.Body.CollisoinLevel = this.CollisoinLevel
	this.Body.MoveDir = vec2d.Vec2{}

	this.Z = 0
}

//计算单个buff对属性的影响
func (this *Unit) CalPropertyByBuffFirst(v1 *Buff, add *UnitBaseProperty) {
	if v1 == nil || v1.IsActive == false || v1.IsUseable == false {
		return
	}
	if v1.AttributePrimaryCV > 0 {
		switch this.AttributePrimary {
		case 1:
			{
				add.AttributeStrength += v1.AttributePrimaryCV
			}
		case 2:
			{
				add.AttributeAgility += v1.AttributePrimaryCV
			}
		case 3:
			{
				add.AttributeIntelligence += v1.AttributePrimaryCV
			}
		}
	}

	add.AttributeStrength += v1.AttributeStrengthCV
	add.AttributeIntelligence += v1.AttributeIntelligenceCV
	add.AttributeAgility += v1.AttributeAgilityCV

}

func (this *Unit) AddBuffPropertyFirst(add *UnitBaseProperty) {

	this.AttributeStrength += add.AttributeStrength
	this.AttributeIntelligence += add.AttributeIntelligence
	this.AttributeAgility += add.AttributeAgility

}

//计算单个buff对属性的影响
func (this *Unit) CalPropertyByBuff(v1 *Buff, add *UnitBaseProperty) {
	if v1 == nil || v1.IsActive == false || v1.IsUseable == false {
		return
	}

	//add.Attack += int32(float32(this.Attack) * v1.AttackSpeedCR)
	add.AttackSpeed += float32(this.AttackSpeed) * v1.AttackSpeedCR
	add.MoveSpeed += this.MoveSpeed * float64(v1.MoveSpeedCR)
	add.MPRegain += this.MPRegain * v1.MPRegainCR
	add.PhysicalAmaor += this.PhysicalAmaor * v1.PhysicalAmaorCR
	add.HPRegain += this.HPRegain * v1.HPRegainCR
	add.AttributeStrength += v1.AttributeStrengthCV
	add.AttributeIntelligence += v1.AttributeIntelligenceCV
	add.AttributeAgility += v1.AttributeAgilityCV
	add.AttackSpeed += v1.AttackSpeedCV
	add.Attack += int32(v1.AttackCV)
	add.Attack += int32(float32(this.Attack) * v1.AttackCR)
	add.MoveSpeed += float64(v1.MoveSpeedCV)
	add.MPRegain += v1.MPRegainCV
	add.PhysicalAmaor += v1.PhysicalAmaorCV
	add.HPRegain += v1.HPRegainCV
	add.HPRegain += v1.HPRegainCVOfMaxHP * float32(this.MAX_HP)
	add.AddedMagicRange += v1.AddedMagicRangeCV
	add.PhysicalHurtAddHP += v1.PhysicalHurtAddHP
	add.MagicHurtAddHP += v1.MagicHurtAddHP
	add.AllHurtCV += v1.AllHurtCV
	add.DoAllHurtCV += v1.DoAllHurtCV
	add.AttackRange += v1.AttackRangeCV
	add.AddHPEffect += v1.AddHPEffectCV
	add.RecoverHurt += v1.RecoverHurt

	add.AddHP += v1.AddHP
	add.AddMP += v1.AddMP //增加血量//增加蓝量

	this.MagicScale = utils.NoLinerAdd(this.MagicScale, v1.MagicScaleCV)
	this.MagicAmaor = utils.NoLinerAdd(this.MagicAmaor, v1.MagicAmaorCV)
	this.StatusAmaor = utils.NoLinerAdd(this.StatusAmaor, v1.StatusAmaorCV)

	this.Dodge = utils.NoLinerAdd(this.Dodge, v1.DodgeCV)
	this.NoCareDodge = utils.NoLinerAdd(this.NoCareDodge, v1.NoCareDodgeCV)
	this.ManaCost = utils.NoLinerAdd(this.ManaCost, v1.ManaCostCV)
	//this.AddHPEffect = utils.NoLinerAdd(this.AddHPEffect, v1.AddHPEffectCV)
	this.MagicCD = utils.NoLinerAdd(this.MagicCD, v1.MagicCDCV)

	if v1.NoMove == 1 {
		this.MoveEnable = 2
	}
	if v1.NoTurn == 1 {
		this.TurnEnable = 2
	}
	if v1.NoAttack == 1 {
		this.AttackEnable = 2
	}
	if v1.NoPassiveSkill == 1 {
		this.NoPassiveSkill = 1
	}
	if v1.NoSkill == 1 {
		this.SkillEnable = 2
	}
	if v1.NoItem == 1 {
		this.ItemEnable = 2
	}
	if v1.MagicImmune == 1 {
		this.MagicImmune = 1
	}
	if v1.Invisible == 1 {
		this.Invisible = 1
	}
	if v1.InvisibleBeSee == 1 {
		this.InvisibleBeSee = 1
	}
	if v1.CanSeeInvisible == 1 {
		this.CanSeeInvisible = 1
	}
	if v1.MasterInvisible == 1 {
		this.MasterInvisible = 1
	}

	if v1.PhisicImmune == 1 {
		this.PhisicImmune = 1
	}
	if v1.MagicCDStop == 1 {
		this.MagicCDStop = 1
	}
	if v1.AnimotorPause == 1 {
		this.AnimotorPause = 1
	}
	if v1.IsCollisoin == 2 {
		this.Body.IsCollisoin = false
	}
	if v1.CollisoinLevel >= 0 {
		this.Body.CollisoinLevel = v1.CollisoinLevel
	}
	if v1.NoPlayerControl == 1 {
		this.PlayerControlEnable = 2
	}
	if v1.ChangeZ > 0 {
		this.Z = v1.ChangeZ
	}

}

func (this *Unit) AddBuffProperty(add *UnitBaseProperty) {
	//属性的增加需要在最开始计算
	//this.AttributeStrength += add.AttributeStrength
	//this.AttributeIntelligence += add.AttributeIntelligence
	//this.AttributeAgility += add.AttributeAgility
	this.AttackSpeed += add.AttackSpeed
	this.Attack += add.Attack
	this.MoveSpeed += add.MoveSpeed
	this.MPRegain += add.MPRegain
	this.PhysicalAmaor += add.PhysicalAmaor
	this.HPRegain += add.HPRegain
	this.AddedMagicRange += add.AddedMagicRange
	this.PhysicalHurtAddHP += add.PhysicalHurtAddHP
	this.MagicHurtAddHP += add.MagicHurtAddHP
	this.AllHurtCV += add.AllHurtCV
	this.DoAllHurtCV += add.DoAllHurtCV
	this.AttackRange += add.AttackRange
	this.AddHPEffect += add.AddHPEffect
	this.RecoverHurt += add.RecoverHurt

}

//刷新攻击时对特定目标起作用的buff useable
func (this *Unit) FreshBuffsUseable(target *Unit) {
	for _, v := range this.Buffs {
		for _, v1 := range v {
			v1.FreshUseable(target)
		}

	}
}

//计算所有buff对1级属性的影响
func (this *Unit) CalPropertyByBuffsFirst() {
	add := &UnitBaseProperty{}

	//buff
	for _, v := range this.Buffs {

		if len(v) <= 0 {
			continue
		}
		if v[0].OverlyingType == 4 {
			this.CalPropertyByBuffFirst(v[0], add)
		} else {
			for _, v1 := range v {
				this.CalPropertyByBuffFirst(v1, add)
			}
		}

	}
	this.AddBuffPropertyFirst(add)

}

//计算所有buff对2级属性的影响
func (this *Unit) CalPropertyByBuffs() {
	add := &UnitBaseProperty{}

	//buff
	for _, v := range this.Buffs {
		if len(v) <= 0 {
			continue
		}
		if v[0].OverlyingType == 4 {
			this.CalPropertyByBuff(v[0], add)
		} else {
			for _, v1 := range v {
				this.CalPropertyByBuff(v1, add)
			}
		}

	}
	this.AddBuffProperty(add)

	////hp---mp---
	this.CalMaxHP_MaxHP(add)

	//物理伤害抵挡
	this.PhysicalResist = utils.UnitPhysicalAmaor2PhysicalResist(this.PhysicalAmaor)

	//攻击速度取值范围
	if this.AttackSpeed <= 10 {
		this.AttackSpeed = 10
	} else if this.AttackSpeed >= float32(this.BaseMaxAttackSpeed) {
		this.AttackSpeed = float32(this.BaseMaxAttackSpeed)
	}
}

//计算属性 (每一帧 都可能会变)
func (this *Unit) CalProperty() {
	//计算属性
	this.CalAttribute()
	//计算buff对属性的影响
	this.CalPropertyByBuffsFirst()
	//计算MAXHP MP
	//this.CalMaxHP_MaxHP()
	//计算攻击速度
	this.CalAttackSpeed()
	//计算攻击力
	this.CalAttack()
	//计算攻击距离
	this.CalAttackRange()
	//计算移动速度
	this.CalMoveSpeed()
	//计算技能增强
	this.CalMagicScale()
	//计算魔法回复
	this.CalMPRegain()
	//计算护甲 物理伤害抵挡
	this.CalPhysicalAmaor()
	//计算魔抗
	this.CalMagicAmaor()
	//计算状态抗性
	this.CalStatusAmaor()
	//计算闪避
	this.CalDodge()
	//计算生命回复
	this.CalHPRegain()
	//计算异常状态
	this.CalControlState()

	//计算buff对属性的影响
	this.CalPropertyByBuffs()
}

//清空buff
func (this *Unit) ClearBuff() {
	this.Buffs = make(map[int32][]*Buff)
}

//删除buff 删除使用技能后失效的buff
func (this *Unit) RemoveBuffForDoSkilled() {
	//buff
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			if v1.DoSkilledInvalid == 1 {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}
		}
	}
}

//移动后失效
func (this *Unit) RemoveBuffForMoved() {
	//buff
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			if v1.MoveInvalid == 1 {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}
		}
	}
}

//击杀单位后失效
func (this *Unit) RemoveBuffForKilled() {
	//buff
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			if v1.KilledInvalid == 1 {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}
		}
	}
}

//击杀单位后失效
func (this *Unit) RemoveHaloForKilled() {
	this.InScene.RemoveHaloForKilled(this)

}

//删除buff 删除攻击后失效的buff
func (this *Unit) RemoveBuffForAttacked() {
	//buff
	for k, v := range this.Buffs {
		for k1, v1 := range v {

			//攻击时减少标记
			if v1.SubTagNumRule == 1 {
				v1.TagNum -= 1
			}

			if v1.AttackedInvalid == 1 || v1.TagNum == 0 {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}

		}

	}
}

//被目标 驱散自己的buff
func (this *Unit) ClearBuffForTarget(target *Unit, clearlevel int32) {
	if target == nil || clearlevel <= 0 || target.IsDisappear() {
		return
	}

	//buff
	isenemy := target.CheckIsEnemy(this)
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			//BuffType         int32 //buff类型 1:表示良性 2:表示恶性  队友只能驱散我的恶性buff 敌人只能驱散我的良性buff
			//ClearLevel       int32 //驱散等级 1 表示需要驱散等级大于等于1的 驱散效果才能驱散此buff
			if isenemy == true && v1.BuffType == 1 && clearlevel >= v1.ClearLevel {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}
			if isenemy == false && v1.BuffType == 2 && clearlevel >= v1.ClearLevel {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}

		}

	}

}

//通过 buff 添加buff
func (this *Unit) AddBuffFromBuff(buff *Buff, castunit *Unit) *Buff {

	if castunit == nil || castunit.IsDisappear() || this.IsDisappear() {
		//log.Info("aaaaaaaaaaaaaa")
		return nil
	}
	//攻击距离类型
	if buff.ActiveUnitAcpabilities == 1 && this.AttackAcpabilities != 1 {
		//log.Info("eeeeeeeeeee")
		return nil
	}
	if buff.ActiveUnitAcpabilities == 2 && this.AttackAcpabilities != 2 {
		//log.Info("fffffffffffff")
		return nil
	}
	//BuffType         int32 //buff类型 1:表示良性 2:表示恶性  队友只能驱散我的恶性buff 敌人只能驱散我的良性buff
	isenemy := castunit.CheckIsEnemy(this)
	//如果是敌人 且 是良性buff 就不添加
	if isenemy == true && buff.BuffType == 1 {
		//log.Info("bbbbbbbbbb")
		return nil
	}
	//如果不是敌人 且 是恶性buff 就不添加
	if isenemy == false && buff.BuffType == 2 {
		//log.Info("ccccccccccc")
		return nil
	}

	//如果恶性buff 单位魔法免疫 buff没有无视技能免疫
	if buff.BuffType == 2 && this.MagicImmune == 1 && buff.NoCareMagicImmuneAddBuff == 2 {
		//log.Info("dddddddddddddd")
		return nil
	}
	buff.CastUnit = castunit

	//状态抗性 减状态时间
	if buff.BuffType == 2 && buff.IsCalStatusAmaor == 1 && this.StatusAmaor != 0 {
		buff.Time = buff.Time - buff.Time*this.StatusAmaor
		buff.RemainTime = buff.Time
	}

	bf, ok := this.Buffs[buff.TypeID]

	//叠加机制
	//		OverlyingType          int32 //叠加类型 1:只更新最大时间 2:完美叠加(小鱼的偷属性) 4:添加时叠加 发生作用时不叠加
	//	OverlyingAddTag        int32 //叠加时是否增加标记数字 1:表示增加 2:表示不增加 3:最大标记覆盖原值
	if ok == true && len(bf) > 0 {
		if buff.OverlyingType == 1 {

			if bf[0].RemainTime < buff.Time {
				bf[0].RemainTime = buff.Time
			}
			if buff.OverlyingAddTag == 1 {
				bf[0].TagNum += buff.TagNum
			} else if buff.OverlyingAddTag == 3 {
				log.Info("--11111111---%d   %d", buff.TagNum, bf[0].TagNum)
				if buff.TagNum > bf[0].TagNum {
					bf[0].TagNum = buff.TagNum
				}
			}
			//log.Info("bb--111111122:%d  %f  %f  %f  ", buff.TypeID, bf[0].RemainTime, buff.Time, utils.GetCurTimeOfSecond())
			return bf[0]

		} else if buff.OverlyingType == 2 || buff.OverlyingType == 4 {
			this.Buffs[buff.TypeID] = append(bf, buff)
			//log.Info("--111111133:%d", buff.TypeID)
			this.CheckTriggerCreateBuff(buff)
			return buff
		} else if buff.OverlyingType == 3 {
			bf[0] = buff
			log.Info("--111111122:%d", buff.TypeID)
			this.CheckTriggerCreateBuff(buff)
			return buff
		}
	} else {
		bfs := make([]*Buff, 0)
		bfs = append(bfs, buff)
		this.Buffs[buff.TypeID] = bfs
		//log.Info("--111111144")
		//log.Info("aa--111111122:%d  %f  %f", buff.TypeID, utils.GetCurTimeOfSecond(), bfs[0].RemainTime)
		this.CheckTriggerCreateBuff(buff)

		//给单位计算buff效果
		//this.CalPropertyByBuffCV(buff)

		return buff
	}
	//log.Info("ggggggggggggg")
	return nil
}

func (this *Unit) GetBuff(typeid int32) *Buff {
	//var re *Buff = nil
	bf, ok := this.Buffs[typeid]
	if ok == true && len(bf) > 0 {
		return bf[0]
	}

	return nil
}

//通过bufftypeid string 添加buff  castunit给我添加
func (this *Unit) AddBuffFromStr(buffsstr string, level int32, castunit *Unit) []*Buff {
	//log.Info("---------------------:%s", buffsstr)
	buffs := utils.GetInt32FromString2(buffsstr)
	re := make([]*Buff, 0)
	for _, v := range buffs {
		buff := NewBuff(v, level, this)
		//log.Info("----------buff:%d", buff.TypeID)
		if buff != nil {
			buff = this.AddBuffFromBuff(buff, castunit)
			if buff != nil {
				re = append(re, buff)
			}

		}
	}
	return re
}

//通过bufftypeid string 添加halo
func (this *Unit) AddHaloFromStr(halosstr string, level int32, pos *vec2d.Vec2) []*Halo {
	halos := utils.GetInt32FromString2(halosstr)
	re := make([]*Halo, 0)
	for _, v := range halos {
		halo := NewHalo(v, level)
		halo.SetParent(this)
		if pos != nil {
			halo.Position = *pos
		}
		if halo != nil {
			this.InScene.AddHalo(halo)
		}
		re = append(re, halo)
	}
	return re
}

//客户端 显示自己受到的伤害
func (this *Unit) CreateShowHurt(hurtvalue int32) {
	if this.MyPlayer == nil {
		return
	}
	//为了显示 玩家造成的伤害
	mph := &protomsg.MsgPlayerHurt{HurtUnitID: this.ID, HurtAllValue: hurtvalue}
	this.MyPlayer.AddHurtValue(mph)
}

//直接造成伤害
func (this *Unit) BeAttackedFromValue(value int32, attackunit *Unit) {

	if value >= 0 {
		return
	}

	lasthp := this.HP
	this.ChangeHP(value)

	//
	this.CheckTriggerBeAttackAfter(attackunit, 1, value)

	//客户端显示自己受伤数字
	this.CreateShowHurt(value)

	//被攻击死亡
	if this.HP <= 0 && lasthp > 0 {
		this.CheckTriggerDie(attackunit)
	}

	this.SaveTimeAndHurt(value)
}

//检测伤害反弹
func (this *Unit) CheckRecover(bullet *Bullet) {
	if bullet == nil || bullet.SrcUnit == nil {
		return
	}
	//RecoverHurt
	if this.RecoverHurt > 0.0001 && bullet.IsRecoverHurt != 1 {
		physicAttack := bullet.GetAttackOfType(1) //物理攻击
		magicAttack := bullet.GetAttackOfType(2)
		pureAttack := bullet.GetAttackOfType(3) //纯粹攻击

		b := NewBullet1(this, bullet.SrcUnit)
		//(1:物理伤害 2:魔法伤害 3:纯粹伤害)
		b.AddOtherHurt(HurtInfo{HurtType: 1, HurtValue: physicAttack})
		b.AddOtherHurt(HurtInfo{HurtType: 2, HurtValue: magicAttack})
		b.AddOtherHurt(HurtInfo{HurtType: 3, HurtValue: pureAttack})
		b.IsRecoverHurt = 1
		//特殊情况处理
		this.AddBullet(b)
	}

}

//检测攻击命中后 技能触发(小电锤)
func (this *Unit) CheckAttackSucAllSkillTrigger(bullet *Bullet) {

	if bullet == nil || bullet.DestUnit == nil || bullet.DestUnit.IsDisappear() {
		return
	}

	for _, v := range this.ItemSkills {
		this.CheckAttackSucOneSkillTrigger(v, bullet)
	}
}

//检测攻击命中后 技能触发(小电锤)
func (this *Unit) CheckAttackSucOneSkillTrigger(v *Skill, bullet *Bullet) {
	//CastType              int32   // 施法类型:  1:主动技能  2:被动技能
	//TriggerTime int32 //触发时间 0:表示不触发 1:攻击时 2:被攻击时 5:命中后触发

	//主动技能
	if v.CastType == 2 && v.TriggerTime == 5 {
		//animattack
		isTrigger := false

		if v.CheckCDTime() {
			//检查 触发概率 和额外条件
			//log.Info("CheckRandom %f", v.TriggerProbability)
			if utils.CheckRandom(v.TriggerProbability) && this.CheckTriggerOtherRule(v.TriggerOtherRule, v.TriggerOtherRuleParam) {
				isTrigger = true
			}
		}

		//检查cd 魔法消耗
		if isTrigger == true {
			skilldata := v
			target := bullet.DestUnit

			data := &protomsg.CS_PlayerSkill{}
			data.TargetUnitID = bullet.DestUnit.ID
			data.SkillID = v.TypeID
			data.X = float32(target.Body.Position.X)
			data.Y = float32(target.Body.Position.Y)
			targetpos := vec2d.Vec2{X: float64(data.X), Y: float64(data.Y)}

			//驱散自己的buff
			this.ClearBuffForTarget(this, skilldata.MyClearLevel)

			//MyBuff
			buffs := this.AddBuffFromStr(skilldata.MyBuff, skilldata.Level, this)
			for _, v := range buffs {
				v.UseableUnitID = data.TargetUnitID
			}
			//MyHalo
			this.AddHaloFromStr(skilldata.MyHalo, skilldata.Level, nil)

			//BlinkToTarget
			if skilldata.BlinkToTarget == 1 {
				if skilldata.CastTargetType == 1 {
					//对自己施法 分身的时候使用
					this.Body.BlinkToPos(vec2d.Vec2{this.Body.Position.X + float64(utils.GetRandomFloat(2)),
						this.Body.Position.Y + float64(utils.GetRandomFloat(2))}, 0)
				} else {
					log.Info("blinkpos:%v", targetpos)
					this.Body.BlinkToPos(targetpos, 0)
				}
			}

			//创建子弹
			bullets := skilldata.CreateBullet(this, data)
			if len(bullets) > 0 {
				for _, v := range bullets {
					if skilldata.TriggerAttackEffect == 1 {
						this.CheckTriggerAttackSkill(v, make([]int32, 0))
					}
					this.AddBullet(v)
				}

			}

			//加血
			if skilldata.AddHPTarget == 1 {
				this.DoAddHP(skilldata.AddHPType, skilldata.AddHPValue)
			}
			//特殊处理
			targetunit := this.InScene.FindUnitByID(data.TargetUnitID)
			this.DoSkillException(skilldata, targetunit, bullets[0])

			//消耗 CD
			namacost := skilldata.GetManaCost() - int32(this.ManaCost*float32(skilldata.GetManaCost()))
			this.ChangeMP(float32(-namacost))

			cdtime := skilldata.Cooldown - this.MagicCD*skilldata.Cooldown
			//skilldata.FreshCDTime(cdtime)
			this.FreshCDTime(skilldata, cdtime)

			//}
		}
	}
}

//受到来自子弹的伤害 calcmiss是否计算miss 溅射不会miss
func (this *Unit) BeAttacked(bullet *Bullet) (bool, int32, int32, int32) {
	//计算闪避
	if bullet.SkillID == -1 {
		//普通攻击
		isDodge := false //闪避
		//无视闪避
		if utils.CheckRandom(bullet.NoCareDodge) {
			isDodge = false
		} else {
			if utils.CheckRandom(this.Dodge) {
				isDodge = true
			} else {
				isDodge = false
			}
		}

		//闪避了
		if isDodge {
			//本单位显示miss
			this.ShowMiss(true)
			return true, 0, 0, 0
		}

		//触发攻击命中
		if bullet.SrcUnit != nil && bullet.SrcUnit.IsDisappear() == false && bullet.IsRecoverHurt != 1 {
			bullet.SrcUnit.CheckAttackSucAllSkillTrigger(bullet)
		}
	}
	//检测伤害反弹
	this.CheckRecover(bullet)

	//计算魔法消除
	magicdeletephisichurt := int32(0)
	magicdelete := bullet.GetAttackOfType(10) //魔法消除
	if magicdelete > 0 {

		magicdelete = this.ChangeMP(float32(-magicdelete))
		//魔法消除造成的物理伤害
		if bullet.MagicValueHurt2PhisicHurtCR > 0 {
			magicdeletephisichurt = int32(float32(-magicdelete) * bullet.MagicValueHurt2PhisicHurtCR)
		}

	}

	//计算伤害
	physicAttack := int32(0)
	if this.PhisicImmune != 1 {
		physicAttack = bullet.GetAttackOfType(1) + magicdeletephisichurt //物理攻击
		//log.Info("  physicAttack:%d    %d", physicAttack, magicdeletephisichurt)

		physicAttack = physicAttack - this.CheckPhisicHurtBlock(physicAttack) //伤害格挡
		//计算护甲抵消后伤害
		if bullet.DoHurtPhysicalAmaorCV == 0 {
			physicAttack = int32(utils.SetValueGreaterE(float32(physicAttack)*(1-this.PhysicalResist), 0))
		} else {
			physicalAmaor := this.PhysicalAmaor + bullet.DoHurtPhysicalAmaorCV
			physicalResist := utils.UnitPhysicalAmaor2PhysicalResist(physicalAmaor)
			physicAttack = int32(utils.SetValueGreaterE(float32(physicAttack)*(1-physicalResist), 0))
		}

	}

	magicAttack := bullet.GetAttackOfType(2)            //魔法攻击CheckMagicHurtBlock
	magicAttack = this.CheckMagicHurtBlock(magicAttack) //伤害格挡
	//log.Info("---------magicAttack:%d", magicAttack)
	pureAttack := bullet.GetAttackOfType(3) //纯粹攻击

	//计算魔抗抵消后伤害
	magicAttack = int32(utils.SetValueGreaterE(float32(magicAttack)*(1-this.MagicAmaor), 0))
	//magicAttack = utils.SetValueGreaterE(magicAttack,0)

	//-----扣血--
	hurtvalue := (physicAttack + magicAttack + pureAttack)

	//伤害加深或者减免
	doallhurtcv := float32(0)
	if bullet.SrcUnit != nil && bullet.SrcUnit.IsDisappear() == false {
		doallhurtcv = bullet.SrcUnit.DoAllHurtCV
	}
	hurtvalue += int32(float32(hurtvalue) * (this.AllHurtCV + doallhurtcv))
	if hurtvalue < 0 {
		hurtvalue = 0
	}

	//乘以伤害系数
	//hurtvalue *= hurtratio

	lasthp := this.HP
	this.ChangeHP(-hurtvalue)

	//客户端显示自己受伤数字
	this.CreateShowHurt(-hurtvalue)

	maxhurttype := int32(1)
	if magicAttack > physicAttack && magicAttack > pureAttack {
		maxhurttype = 2
	}
	if pureAttack > physicAttack && pureAttack > magicAttack {
		maxhurttype = 3
	}

	this.CheckTriggerBeAttackAfter(bullet.SrcUnit, maxhurttype, hurtvalue)
	//被攻击死亡
	if this.HP <= 0 && lasthp > 0 {
		this.CheckTriggerDie(bullet.SrcUnit)
	}

	this.SaveTimeAndHurt(-hurtvalue)
	//log.Info("---hurtvalue---:%d   %f", hurtvalue, this.PhysicalResist)
	return false, -hurtvalue, -physicAttack, -magicAttack
}
func (this *Unit) CheckTriggerCreateBuff(v1 *Buff) {
	if v1 == nil {
		return
	}
	//
	if v1.Exception <= 0 {
		return
	}
	switch v1.Exception {
	case 2: //血魔的焦渴
		{
			//log.Info("血魔的焦渴")
			v1.MoveSpeedCR = 0
			v1.AttackSpeedCV = 0
			if this.Body == nil {
				return
			}
			param := utils.GetFloat32FromString3(v1.ExceptionParam, ":")
			if len(param) < 6 {
				return
			}
			maxdis := param[0]
			maxhp := param[1]
			movespeed := param[2]
			attackspeed := param[3]
			minhp := param[4]

			buffstr := strconv.Itoa(int(param[5]))
			//获取范围内的目标单位
			allunit := this.InScene.FindVisibleUnits(this)
			for _, v := range allunit {

				if v.IsDisappear() || this.CheckIsEnemy(v) == false || v.UnitType != 1 || v.Body == nil {
					continue
				}
				//检测是否在范围内
				dis := float32(vec2d.Distanse(this.Body.Position, v.Body.Position))
				if dis > maxdis {
					continue
				}

				hp := float32(v.HP) / float32(v.MAX_HP)
				if hp <= maxhp {
					t1 := 1 - (hp-minhp)/(maxhp-minhp)
					if t1 > 1 { //hp小于最低值
						t1 = 1
						v.AddBuffFromStr(buffstr, 1, this)
					}
					v1.MoveSpeedCR += movespeed * t1
					v1.AttackSpeedCV += attackspeed * t1
				}

			}

			//log.Info("move:%f    attackspeed:%f", v1.MoveSpeedCR, v1.AttackSpeedCV)

		}
	default:
		{

		}
	}
}

//被攻击时 格挡处理 返回格挡了的伤害值
//伤害格挡
//	HurtBlockProbability float32 //伤害格挡几率
//	HurtBlockPhisicValue int32   //物理伤害格挡值
func (this *Unit) CheckPhisicHurtBlock(physicAttack int32) int32 {
	//遍历 可以优化  本体的buff
	for _, v := range this.Buffs {
		for _, v1 := range v {
			//攻击时减少标记
			if v1.HurtBlockProbability <= 0 || v1.HurtBlockPhisicValue <= 0 {
				continue
			}
			if utils.CheckRandom(v1.HurtBlockProbability) {
				//log.Info("格挡:%d  -- %d", v1.HurtBlockPhisicValue, physicAttack)
				if v1.HurtBlockPhisicValue >= physicAttack {

					return physicAttack
				}
				return v1.HurtBlockPhisicValue
			}
		}
	}

	return 0
}

//魔法伤害吸收 返回吸收后剩下的魔法伤害值
func (this *Unit) CheckMagicHurtBlock(magicAttack int32) int32 {
	//遍历 可以优化  本体的buff
	for _, v := range this.Buffs {
		for _, v1 := range v {
			//log.Info("CheckMagicHurtBlock:%d", v1.TypeID)
			magicblock := v1.BlockMagicHurt(magicAttack)
			if magicblock != magicAttack {
				return magicblock
			}
			//return v1.BlockMagicHurt(magicAttack)
		}
	}

	return magicAttack
}

//被攻击时 buff异常处理
func (this *Unit) CheckTriggerBeAttackAfter(attacker *Unit, hurttype int32, hurtvalue int32) {
	if attacker == nil || attacker.IsDisappear() || hurtvalue <= 0 {
		return
	}

	//AI增加仇恨值
	if this.AI != nil {
		this.AI.AddEnemies(attacker, hurtvalue)
	}

	//物理攻击
	//if hurttype == 1 {
	this.CheckTriggerBeAttackSkill(attacker)
	//}

	//-----------处理异常------------------
	//遍历 可以优化  本体的buff
	for _, v := range this.Buffs {
		for _, v1 := range v {
			//攻击时减少标记
			if v1.Exception <= 0 {
				continue
			}
			switch v1.Exception {
			case 5: //幽鬼折射
				{
					param := utils.GetFloat32FromString3(v1.ExceptionParam, ":")
					if len(param) < 3 {
						return
					}
					mindis := param[0]
					maxdis := param[1]
					maxHurt := param[2] * float32(hurtvalue)

					allunit := this.InScene.FindVisibleUnitsByPos(this.Body.Position)
					for _, v := range allunit {
						if v.IsDisappear() {
							continue
						}
						if this.CheckIsEnemy(v) == false {
							continue
						}
						//检测是否在范围内
						if v.Body == nil || maxdis <= 0 {
							continue
						}
						dis := float32(vec2d.Distanse(this.Body.Position, v.Body.Position))
						//log.Info("-----------------dis:%f", dis)
						hv := maxHurt
						if dis <= maxdis {
							if dis > mindis {
								hv = (1 - (dis-mindis)/(maxdis-mindis)) * maxHurt
							}
							b := NewBullet1(this, v)
							//b.SetProjectileMode(this.BulletModeType, this.BulletSpeed)
							b.AddOtherHurt(HurtInfo{HurtType: hurttype, HurtValue: int32(hv)})
							//特殊情况处理
							this.AddBullet(b)
						}

					}
				}
			default:
				{

				}
			}

		}

	}
}

//获得经验
func (this *Unit) AddExperience(add int32) {

	if this.Level >= conf.MaxLevel {
		return
	}

	//检查今天是否还能获取超过add的经验值
	if this.RemainExperience < add {
		add = this.RemainExperience

	}
	this.RemainExperience -= add

	this.Experience += add
	if this.Experience >= this.MaxExperience {
		//升级
		this.Level += 1
		this.Experience -= this.MaxExperience
		//满血满蓝
		this.InitHPandMP(1, 1)

		//下一级需要的经验值
		leveldata := conf.GetLevelFileData(this.Level)
		if leveldata != nil {
			this.MaxExperience = leveldata.UpgradeExperience
		}

		//-----------
		//today := time.Now().Format("2006-01-02")
	}
}

//击杀单位获得经验 金币奖励
func (this *Unit) GetRewardForKill(deathunit *Unit) {
	if deathunit == nil {
		return
	}
	//阵营(1:玩家 2:NPC)  玩家的召唤物和幻象camp也是 玩家
	if deathunit.Camp == 2 {

		if deathunit.InScene == nil || this.MyPlayer == nil || this.MyPlayer.MainUnit == nil {
			return
		}
		addgold := float64(deathunit.InScene.UnitGold)
		addExp := float64(deathunit.InScene.UnitExperience)
		var team interface{}
		if this.MyPlayer.TeamID > 0 {
			team = TeamManagerObj.Teams.Get(this.MyPlayer.TeamID)
		}
		//组队时 需要平分
		if team != nil {
			teamplayers := team.(*TeamInfo).Players.Items()
			addgold = math.Ceil(addgold / float64(len(teamplayers)))
			addExp = math.Ceil(addExp / float64(len(teamplayers)))
			for _, v := range teamplayers {
				player := v.(*Player)
				if player != nil && player.MainUnit != nil {
					player.MainUnit.Gold += int32(addgold)
					player.MainUnit.AddExperience(int32(addExp))
					//显示奖励的金币
					mph := &protomsg.MsgPlayerHurt{HurtUnitID: deathunit.ID, GetGold: int32(addgold)}
					player.AddHurtValue(mph)
				}
			}
		} else {
			this.MyPlayer.MainUnit.Gold += int32(addgold)
			this.MyPlayer.MainUnit.AddExperience(int32(addExp))
			//显示奖励的金币
			mph := &protomsg.MsgPlayerHurt{HurtUnitID: deathunit.ID, GetGold: int32(addgold)}
			if this.MyPlayer != nil {
				this.MyPlayer.AddHurtValue(mph)
			}
		}

	}
}

//被杀死时 buff异常处理
func (this *Unit) CheckTriggerDie(killer *Unit) {

	//处理死亡单位
	//重置死亡复活时间
	leveldata := conf.GetLevelFileData(this.Level)
	if leveldata != nil {
		this.RemainReviveTime = leveldata.ReviveTime
	} else {
		this.RemainReviveTime = 100
	}

	//处理击杀者相关逻辑
	if killer == nil || killer.IsDisappear() {
		return
	}
	//击杀者获得奖励
	killer.GetRewardForKill(this)

	killer.RemoveBuffForKilled()
	killer.RemoveHaloForKilled()
	killer.CheckTriggerKillerSkill(this)

	//--------------处理buff Exception-------------

	//遍历 可以优化  本体的buff
	for _, v := range this.Buffs {
		for _, v1 := range v {
			//攻击时减少标记
			if v1.Exception <= 0 {
				continue
			}
			switch v1.Exception {
			case 1: //血魔的血怒 死亡后加血
				{
					param := utils.GetFloat32FromString3(v1.ExceptionParam, ":")
					if len(param) <= 0 {
						return
					}
					addhp := float32(this.MAX_HP) * param[0]
					if addhp > 0 {
						killer.ChangeHP(int32(addhp))
					}
				}
			default:
				{

				}
			}

		}

	}
	//遍历 可以优化  killer的buff
	for _, v := range killer.Buffs {
		for _, v1 := range v {
			//攻击时减少标记
			if v1.Exception <= 0 {
				continue
			}
			switch v1.Exception {
			case 1: //血魔的血怒 死亡后加血
				{
					param := utils.GetFloat32FromString3(v1.ExceptionParam, ":")
					if len(param) <= 0 {
						return
					}
					addhp := float32(this.MAX_HP) * param[0]
					if addhp > 0 {
						killer.ChangeHP(int32(addhp))
					}
				}
			default:
				{

				}
			}

		}

	}
}

//初始刷新技能CD  不能刷新自己的CD
func (this *Unit) DoFreshSkillTime(skillid int32) {
	for _, v := range this.Skills {
		if v.TypeID == skillid {
			continue
		}
		v.FreshSkill()
	}
	for _, v := range this.ItemSkills {
		if v.TypeID == skillid {
			continue
		}
		v.FreshSkill()
	}

}

//创建子弹
func (this *Unit) AddBullet(bullet *Bullet) {
	if this.InScene != nil {
		this.InScene.AddBullet(bullet)
	}
}

//刷新客户端显示数据
func (this *Unit) FreshClientData() {

	if this.ClientData == nil {
		this.ClientData = &protomsg.UnitDatas{}
	}

	this.ClientData.HP = this.HP
	this.ClientData.MaxHP = this.MAX_HP
	this.ClientData.MP = int32(this.MP)
	this.ClientData.MaxMP = this.MAX_MP
	this.ClientData.Name = this.Name
	this.ClientData.Level = this.Level
	this.ClientData.ID = this.ID
	this.ClientData.ModeType = this.ModeType
	this.ClientData.Experience = this.Experience
	this.ClientData.MaxExperience = this.MaxExperience
	this.ClientData.ControlID = this.ControlID
	this.ClientData.AnimotorState = this.AnimotorState
	this.ClientData.AttackTime = this.GetOneAttackTime()

	this.ClientData.X = float32(this.Body.Position.X)
	this.ClientData.Y = float32(this.Body.Position.Y)

	this.ClientData.DirectionX = float32(this.Body.Direction.X)
	this.ClientData.DirectionY = float32(this.Body.Direction.Y)

	this.ClientData.UnitType = this.UnitType
	this.ClientData.AttackAcpabilities = this.AttackAcpabilities
	this.ClientData.AttackMode = this.AttackMode
	this.ClientData.IsMain = this.IsMain
	this.ClientData.IsDeath = this.IsDeath
	this.ClientData.Invisible = this.Invisible
	this.ClientData.Camp = this.Camp
	this.ClientData.AnimotorPause = this.AnimotorPause
	this.ClientData.SkillEnable = this.SkillEnable
	this.ClientData.ItemEnable = this.ItemEnable
	this.ClientData.Z = this.Z
	this.ClientData.IsMirrorImage = this.IsMirrorImage
	this.ClientData.AttackRange = this.AttackRange
	this.ClientData.AttackAnim = this.AttackAnim
	this.ClientData.TypeID = this.TypeID
	this.ClientData.RemainReviveTime = this.RemainReviveTime
	this.ClientData.Gold = this.Gold
	this.ClientData.Diamond = this.Diamond
	if this.MyPlayer != nil {
		this.ClientData.TeamID = this.MyPlayer.TeamID
	}

	//道具技能
	isds := make(map[int32]int32)
	this.ClientData.ISD = make([]*protomsg.SkillDatas, 0)
	for _, v := range this.ItemSkills {

		if v.CastType != 1 {
			continue
		}

		if _, ok := isds[v.TypeID]; ok {
			continue
		} else {
			isds[v.TypeID] = v.TypeID
		}

		skdata := &protomsg.SkillDatas{}
		skdata.TypeID = v.TypeID
		skdata.Level = v.Level
		skdata.RemainCDTime = v.RemainCDTime
		skdata.CanUpgrade = int32(2) //v.CanUpgrade
		skdata.Index = v.Index
		skdata.CastType = v.CastType
		skdata.CastTargetType = v.CastTargetType
		skdata.UnitTargetTeam = v.UnitTargetTeam
		skdata.UnitTargetCamp = v.UnitTargetCamp
		skdata.NoCareMagicImmune = v.NoCareMagicImmune
		skdata.CastRange = v.CastRange + this.AddedMagicRange
		skdata.Cooldown = v.Cooldown
		skdata.HurtRange = v.HurtRange
		skdata.ManaCost = v.GetManaCost()
		skdata.AttackAutoActive = v.AttackAutoActive
		skdata.Visible = v.Visible
		skdata.RemainSkillCount = v.RemainSkillCount
		skdata.MaxLevel = v.MaxLevel
		this.ClientData.ISD = append(this.ClientData.ISD, skdata)
	}

	//技能AttackAnim
	this.ClientData.SD = make([]*protomsg.SkillDatas, 0)
	for _, v := range this.Skills {
		skdata := &protomsg.SkillDatas{}
		skdata.TypeID = v.TypeID
		skdata.Level = v.Level
		skdata.RemainCDTime = v.RemainCDTime
		skdata.CanUpgrade = int32(2) //v.CanUpgrade
		skdata.Index = v.Index
		skdata.CastType = v.CastType
		skdata.CastTargetType = v.CastTargetType
		skdata.UnitTargetTeam = v.UnitTargetTeam
		skdata.UnitTargetCamp = v.UnitTargetCamp
		skdata.NoCareMagicImmune = v.NoCareMagicImmune
		skdata.CastRange = v.CastRange + this.AddedMagicRange
		skdata.Cooldown = v.Cooldown
		skdata.HurtRange = v.HurtRange
		skdata.ManaCost = v.GetManaCost()
		skdata.AttackAutoActive = v.AttackAutoActive
		skdata.Visible = v.Visible
		skdata.RemainSkillCount = v.RemainSkillCount
		skdata.MaxLevel = v.MaxLevel
		skdata.RequiredLevel = v.RequiredLevel
		skdata.LevelsBetweenUpgrades = v.LevelsBetweenUpgrades
		skdata.InitLevel = v.InitLevel
		this.ClientData.SD = append(this.ClientData.SD, skdata)
	}
	//Buffs  map[int32][]*Buff //所有buff 同typeID下可能有多个buff
	this.ClientData.BD = make([]*protomsg.BuffDatas, 0)
	for _, v := range this.Buffs {
		if len(v) <= 0 {
			continue
		}
		buffdata := &protomsg.BuffDatas{}
		buffdata.TypeID = v[0].TypeID
		buffdata.RemainTime = v[0].RemainTime
		buffdata.Time = v[0].Time
		buffdata.ConnectionType = v[0].ConnectionType
		buffdata.ConnectionX = float32(v[0].ConnectionPoint.X)
		buffdata.ConnectionY = float32(v[0].ConnectionPoint.Y)
		buffdata.ConnectionZ = float32(0)

		if len(v) > 1 {
			buffdata.TagNum = int32(len(v))
		} else {
			buffdata.TagNum = v[0].TagNum
		}
		this.ClientData.BD = append(this.ClientData.BD, buffdata)
	}

	//Skills map[int32]*Skill //所有技能

}

//func (this *Unit) OnePropSub(prop interface{}){

//}

//刷新客户端显示差异数据
func (this *Unit) FreshClientDataSub() {

	if this.ClientDataSub == nil {
		this.ClientDataSub = &protomsg.UnitDatas{}
	}
	if this.ClientData == nil {
		this.FreshClientData()
		*this.ClientDataSub = *this.ClientData
		return
	}

	//
	//字符串部分
	if strings.Compare(this.Name, this.ClientData.Name) != 0 {
		this.ClientDataSub.Name = this.Name
	} else {
		this.ClientDataSub.Name = ""
	}
	if strings.Compare(this.ModeType, this.ClientData.ModeType) != 0 {
		this.ClientDataSub.ModeType = this.ModeType
	} else {
		this.ClientDataSub.ModeType = ""
	}

	//当前数据与上一次数据对比 相减 数值部分
	this.ClientDataSub.HP = this.HP - this.ClientData.HP
	this.ClientDataSub.MaxHP = this.MAX_HP - this.ClientData.MaxHP
	this.ClientDataSub.MP = int32(this.MP) - this.ClientData.MP
	this.ClientDataSub.MaxMP = this.MAX_MP - this.ClientData.MaxMP
	this.ClientDataSub.Level = this.Level - this.ClientData.Level
	this.ClientDataSub.Experience = this.Experience - this.ClientData.Experience
	this.ClientDataSub.MaxExperience = this.MaxExperience - this.ClientData.MaxExperience
	this.ClientDataSub.ControlID = this.ControlID - this.ClientData.ControlID

	this.ClientDataSub.AnimotorState = this.AnimotorState - this.ClientData.AnimotorState
	this.ClientDataSub.AttackTime = this.GetOneAttackTime() - this.ClientData.AttackTime

	this.ClientDataSub.X = float32(this.Body.Position.X) - this.ClientData.X
	this.ClientDataSub.Y = float32(this.Body.Position.Y) - this.ClientData.Y

	this.ClientDataSub.DirectionX = float32(this.Body.Direction.X) - this.ClientData.DirectionX
	this.ClientDataSub.DirectionY = float32(this.Body.Direction.Y) - this.ClientData.DirectionY

	this.ClientDataSub.UnitType = this.UnitType - this.ClientData.UnitType
	this.ClientDataSub.AttackAcpabilities = this.AttackAcpabilities - this.ClientData.AttackAcpabilities
	this.ClientDataSub.AttackMode = this.AttackMode - this.ClientData.AttackMode

	this.ClientDataSub.IsMain = this.IsMain - this.ClientData.IsMain
	this.ClientDataSub.IsDeath = this.IsDeath - this.ClientData.IsDeath
	this.ClientDataSub.Invisible = this.Invisible - this.ClientData.Invisible
	this.ClientDataSub.Camp = this.Camp - this.ClientData.Camp
	this.ClientDataSub.AnimotorPause = this.AnimotorPause - this.ClientData.AnimotorPause
	this.ClientDataSub.SkillEnable = this.SkillEnable - this.ClientData.SkillEnable
	this.ClientDataSub.ItemEnable = this.ItemEnable - this.ClientData.ItemEnable
	this.ClientDataSub.Z = this.Z - this.ClientData.Z
	this.ClientDataSub.IsMirrorImage = this.IsMirrorImage - this.ClientData.IsMirrorImage
	this.ClientDataSub.AttackRange = this.AttackRange - this.ClientData.AttackRange
	this.ClientDataSub.AttackAnim = this.AttackAnim - this.ClientData.AttackAnim
	this.ClientDataSub.TypeID = this.TypeID - this.ClientData.TypeID
	this.ClientDataSub.RemainReviveTime = this.RemainReviveTime - this.ClientData.RemainReviveTime
	this.ClientDataSub.Gold = this.Gold - this.ClientData.Gold
	this.ClientDataSub.Diamond = this.Diamond - this.ClientData.Diamond

	if this.MyPlayer != nil {
		this.ClientDataSub.TeamID = this.MyPlayer.TeamID - this.ClientData.TeamID
	}
	//道具技能
	isds := make(map[int32]int32)
	this.ClientDataSub.ISD = make([]*protomsg.SkillDatas, 0)
	for _, v := range this.ItemSkills {

		if v.CastType != 1 {
			continue
		}
		if _, ok := isds[v.TypeID]; ok {
			continue
		} else {
			isds[v.TypeID] = v.TypeID
		}

		skdata := &protomsg.SkillDatas{}
		skdata.TypeID = v.TypeID
		//上次发送的数据
		lastdata := &protomsg.SkillDatas{}
		for _, v1 := range this.ClientData.ISD {
			if v1.TypeID == v.TypeID {
				lastdata = v1
				break
			}
		}

		skdata.Level = v.Level - lastdata.Level
		skdata.RemainCDTime = v.RemainCDTime - lastdata.RemainCDTime
		skdata.CanUpgrade = int32(2) - lastdata.CanUpgrade //v.CanUpgrade
		skdata.Index = v.Index - lastdata.Index
		skdata.CastType = v.CastType - lastdata.CastType
		skdata.CastTargetType = v.CastTargetType - lastdata.CastTargetType
		skdata.UnitTargetTeam = v.UnitTargetTeam - lastdata.UnitTargetTeam
		skdata.UnitTargetCamp = v.UnitTargetCamp - lastdata.UnitTargetCamp
		skdata.NoCareMagicImmune = v.NoCareMagicImmune - lastdata.NoCareMagicImmune
		skdata.CastRange = v.CastRange + this.AddedMagicRange - lastdata.CastRange
		skdata.Cooldown = v.Cooldown - lastdata.Cooldown
		skdata.HurtRange = v.HurtRange - lastdata.HurtRange
		skdata.ManaCost = v.GetManaCost() - lastdata.ManaCost
		skdata.AttackAutoActive = v.AttackAutoActive - lastdata.AttackAutoActive
		skdata.Visible = v.Visible - lastdata.Visible
		skdata.RemainSkillCount = v.RemainSkillCount - lastdata.RemainSkillCount
		skdata.MaxLevel = v.MaxLevel - lastdata.MaxLevel
		this.ClientDataSub.ISD = append(this.ClientDataSub.ISD, skdata)
	}

	//技能
	this.ClientDataSub.SD = make([]*protomsg.SkillDatas, 0)
	for _, v := range this.Skills {
		skdata := &protomsg.SkillDatas{}
		skdata.TypeID = v.TypeID
		//上次发送的数据
		lastdata := &protomsg.SkillDatas{}
		for _, v1 := range this.ClientData.SD {
			if v1.TypeID == v.TypeID {
				lastdata = v1
				break
			}
		}

		skdata.Level = v.Level - lastdata.Level
		skdata.RemainCDTime = v.RemainCDTime - lastdata.RemainCDTime
		skdata.CanUpgrade = int32(2) - lastdata.CanUpgrade //v.CanUpgrade
		skdata.Index = v.Index - lastdata.Index
		skdata.CastType = v.CastType - lastdata.CastType
		skdata.CastTargetType = v.CastTargetType - lastdata.CastTargetType
		skdata.UnitTargetTeam = v.UnitTargetTeam - lastdata.UnitTargetTeam
		skdata.UnitTargetCamp = v.UnitTargetCamp - lastdata.UnitTargetCamp
		skdata.NoCareMagicImmune = v.NoCareMagicImmune - lastdata.NoCareMagicImmune
		skdata.CastRange = v.CastRange + this.AddedMagicRange - lastdata.CastRange
		skdata.Cooldown = v.Cooldown - lastdata.Cooldown
		skdata.HurtRange = v.HurtRange - lastdata.HurtRange
		skdata.ManaCost = v.GetManaCost() - lastdata.ManaCost
		skdata.AttackAutoActive = v.AttackAutoActive - lastdata.AttackAutoActive
		skdata.Visible = v.Visible - lastdata.Visible
		skdata.RemainSkillCount = v.RemainSkillCount - lastdata.RemainSkillCount
		skdata.MaxLevel = v.MaxLevel - lastdata.MaxLevel
		skdata.RequiredLevel = v.RequiredLevel - lastdata.RequiredLevel
		skdata.LevelsBetweenUpgrades = v.LevelsBetweenUpgrades - lastdata.LevelsBetweenUpgrades
		skdata.InitLevel = v.InitLevel - lastdata.InitLevel
		this.ClientDataSub.SD = append(this.ClientDataSub.SD, skdata)
	}

	this.ClientDataSub.BD = make([]*protomsg.BuffDatas, 0)
	for _, v := range this.Buffs {
		if len(v) <= 0 {
			continue
		}
		buffdata := &protomsg.BuffDatas{}
		buffdata.TypeID = v[0].TypeID
		//上次发送的数据
		lastdata := &protomsg.BuffDatas{}
		for _, v1 := range this.ClientData.BD {
			if v1.TypeID == v[0].TypeID {
				lastdata = v1
				break
			}
		}

		buffdata.RemainTime = v[0].RemainTime - lastdata.RemainTime
		buffdata.Time = v[0].Time - lastdata.Time
		if len(v) > 1 {
			buffdata.TagNum = int32(len(v)) - lastdata.TagNum
		} else {
			buffdata.TagNum = v[0].TagNum - lastdata.TagNum
		}

		buffdata.ConnectionType = v[0].ConnectionType - lastdata.ConnectionType
		buffdata.ConnectionX = float32(v[0].ConnectionPoint.X) - lastdata.ConnectionX
		buffdata.ConnectionY = float32(v[0].ConnectionPoint.Y) - lastdata.ConnectionY
		buffdata.ConnectionZ = float32(0) - lastdata.ConnectionZ

		this.ClientDataSub.BD = append(this.ClientDataSub.BD, buffdata)
	}

}

//被删除的时候
func (this *Unit) OnDestroy() {
	this.IsDelete = true
}

//即时属性获取
