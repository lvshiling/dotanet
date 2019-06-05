package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"strconv"
)

type Skill struct {
	conf.SkillData //技能数据

	Level            int32   //技能当前等级
	RemainCDTime     float32 //技能cd 剩余时间
	AttackAutoActive int32   //攻击时自动释放 是否激活 1:激活 2:否
}

//激活与不激活
func (this *Skill) DoActive() {
	if this.AttackAutoActive == 1 {
		this.AttackAutoActive = 2
	} else {
		this.AttackAutoActive = 1
	}
}

//创建子弹
func (this *Skill) CreateBullet(unit *Unit, data *protomsg.CS_PlayerSkill) *Bullet {
	if unit == nil || data == nil {
		return nil
	}
	//
	//自身
	var b *Bullet = nil
	if this.CastTargetType == 1 {
		b = NewBullet1(unit, unit)
	} else if this.CastTargetType == 2 { //目标单位

		targetunit := unit.InScene.FindUnitByID(data.TargetUnitID)
		b = NewBullet1(unit, targetunit)

	} else if this.CastTargetType == 3 || this.CastTargetType == 5 { //目的点
		b = NewBullet2(unit, vec2d.Vec2{float64(data.X), float64(data.Y)})
	}

	b.SetNormalHurtRatio(this.NormalHurt)
	b.SetProjectileMode(this.BulletModeType, this.BulletSpeed)
	//技能增强
	if this.HurtType == 2 {
		hurtvalue := (this.HurtValue + int32(float32(this.HurtValue)*unit.MagicScale))
		b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: hurtvalue})
	} else {
		b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: this.HurtValue})
	}
	b.AddTargetBuff(this.TargetBuff, this.Level)
	b.SkillID = this.TypeID
	b.SkillLevel = this.Level
	//召唤信息
	b.BulletCallUnitInfo = BulletCallUnitInfo{this.CallUnitInfo, this.Level}
	if this.AwaysHurt == 1 {
		b.IsDoHurtOnMove = 1
	}
	//伤害范围 和目标关系
	b.SetRange(this.HurtRange)
	b.UnitTargetTeam = this.UnitTargetTeam
	//强制移动
	b.SetForceMove(this.ForceMoveTime, this.ForceMoveSpeedSize, this.ForceMoveLevel)

	return b
}

//更新
func (this *Skill) Update(dt float64) {
	//CD时间减少
	this.RemainCDTime -= float32(dt)
	if this.RemainCDTime <= 0 {
		this.RemainCDTime = 0
	}
}

//刷新CD
func (this *Skill) FreshCDTime(time float32) {
	this.RemainCDTime = time
}

//返回数据库字符串
func (this *Skill) ToDBString() string {
	return strconv.Itoa(int(this.TypeID)) + "," + strconv.Itoa(int(this.Level)) + "," + strconv.FormatFloat(float64(this.RemainCDTime), 'f', 4, 32)
}

//通过数据库数据和单位基本数据创建技能 (1,2,0) ID,LEVEL,CD剩余时间
func NewUnitSkills(dbdata []string, unitskilldata string) map[int32]*Skill {
	re := make(map[int32]*Skill)

	//单位基本技能
	skillids := utils.GetInt32FromString2(unitskilldata)
	for k, v := range skillids {
		sk := &Skill{}
		skdata := conf.GetSkillData(v, 1)
		if skdata == nil {
			log.Error("NewUnitSkills %d  %d", v, 1)
			continue
		}
		sk.SkillData = *skdata
		sk.SkillData.Index = int32(k)

		log.Info("skill index:%d", sk.SkillData.Index)
		sk.Level = 0
		sk.RemainCDTime = 0
		re[sk.TypeID] = sk
	}
	//数据库技能
	for _, v := range dbdata {

		oneskilldbdata := utils.GetFloat32FromString2(v)
		if len(oneskilldbdata) != 3 {
			continue
		}
		skillid := int32(oneskilldbdata[0])
		skilllevel := int32(oneskilldbdata[1])
		skillcd := oneskilldbdata[2]

		sk := &Skill{}
		skdata := conf.GetSkillData(skillid, skilllevel)
		if skdata == nil {
			log.Error("NewUnitSkills %d  %d", skillid, skilllevel)
			continue

		}
		sk.SkillData = *skdata
		sk.Level = skilllevel
		sk.RemainCDTime = skillcd
		sk.AttackAutoActive = 1
		//sk.RemainCDTime = 10.0
		if initskill, ok := re[sk.TypeID]; ok {
			sk.Index = initskill.Index
			re[sk.TypeID] = sk
		}

	}

	return re
}
