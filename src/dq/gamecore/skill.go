package gamecore

import (
	"dq/conf"
)

type Skill struct {
	conf.SkillData //技能数据

	Level        int32   //技能当前等级
	RemainCDTime float32 //技能cd 剩余时间
}

//通过数据库数据创建技能 (1,2,0) ID,LEVEL,CD剩余时间
func NewSkill(dbdata string) *Skill {

}
