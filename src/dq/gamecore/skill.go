package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/utils"
	"strconv"
)

type Skill struct {
	conf.SkillData //技能数据

	Level        int32   //技能当前等级
	RemainCDTime float32 //技能cd 剩余时间
}

func (this *Skill) Update(dt float64) {
	//CD时间减少
	this.RemainCDTime -= float32(dt)
	if this.RemainCDTime <= 0 {
		this.RemainCDTime = 0
	}
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
		sk.RemainCDTime = 10.0
		if _, ok := re[sk.TypeID]; ok {
			re[sk.TypeID] = sk
		}

	}

	return re
}
