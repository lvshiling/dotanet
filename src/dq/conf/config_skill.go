// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conf

import (
	"dq/log"
	"dq/utils"
)

var (
	SkillFileDatas = make(map[interface{}]interface{})

	//key 为 id_level
	SkillDatas = make(map[interface{}]interface{})
)

//场景配置文件
func LoadSkillFileData() {
	_, SkillFileDatas = utils.ReadXlsxData("bin/conf/skills.xlsx", (*SkillFileData)(nil))

	InitSkillDatas()

}

//初始化具体技能数据
func InitSkillDatas() {

	for _, v := range SkillFileDatas {
		ssd := make([]SkillData, 0)
		v.(*SkillFileData).Trans2SkillData(&ssd)

		for k, v1 := range ssd {
			test := v1
			SkillDatas[string(v1.TypeID)+"_"+string(k+1)] = &test
		}
	}

	//log.Info("----------1---------")

	//log.Info("-:%v", SkillDatas)
	for i := 1; i < 5; i++ {
		t := GetSkillData(1, int32(i))
		if t != nil {
			log.Info("%d:%v", i, *t)
		}

	}

	//log.Info("----------2---------")
}

//获取技能数据 通过技能ID和等级
func GetSkillData(typeid int32, level int32) *SkillData {
	//log.Info("find unitfile:%d", typeid)
	if level <= 0 {
		level = 1
	}
	key := string(typeid) + "_" + string(level)

	re := (SkillDatas[key])
	if re == nil {
		log.Info("not find skill unitfile:%d", typeid)
		return nil
	}
	return (re).(*SkillData)
}

//技能基本数据
type SkillBaseData struct {
	TypeID                int32   //类型ID
	CastType              int32   // 施法类型:  1:主动技能  2:被动技能
	CastTargetType        int32   //施法目标类型 1:自身为目标 2:以单位为目标 3:以地面1点为目标
	CastTargetRange       float32 //施法目标范围 小于等于0表示单体 以施法目标点为中心的范围内的多个目标为 最终弹道目标
	UnitTargetTeam        int32   //目标单位关系 1:友方  2:敌方 3:友方敌方都行
	UnitTargetCamp        int32   //目标单位阵营 (1:玩家 2:NPC) 3:玩家NPC都行
	NoCareMagicImmune     int32   //无视技能免疫 (1:无视技能免疫 2:非)
	BulletModeType        string  //子弹模型
	BulletSpeed           float32 //子弹速度
	HurtType              int32   //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害)
	TriggerAttackEffect   int32   //能否触发普通攻击特效 (1:触发 2:不触发)
	CastPoint             float32 //施法前摇(以施法时间为比列 0.5表示 施法的中间时间点触发)
	CastTime              float32 //施法时间(以秒为单位的时间 比如1秒)
	RequiredLevel         int32   //初始等级需求 1级 需要玩家多少级才能学习
	LevelsBetweenUpgrades int32   //等级需求步长 2
	MaxLevel              int32   //最高升级到的等级 5 表示能升级到5级
	Index                 int32   //技能索引 按升序排列  在屏幕右下角的显示位置
	TargetBuff            string  //对目标造成的buff 比如 1,2 表示对目标造成typeid为 1和2的buff
	BlinkToTarget         int32   //是否瞬间移动到目的地 1:是 2:否
	MyBuff                string  //对自己造成的buff 比如 1,2 表示对目标造成typeid为 1和2的buff
}

//技能数据 (会根据等级变化的数据)
type SkillData struct {
	SkillBaseData
	CastRange  float32 //施法距离
	Cooldown   float32 //技能冷却时间
	HurtValue  int32   //技能伤害
	HurtRange  float32 //伤害范围 小于等于0表示单体
	NormalHurt float32 //附带普通攻击百分比 (0.5 为 50%的普通攻击伤害) 一般为0
	ManaCost   int32   //技能魔法消耗

}

//单位配置文件数据
type SkillFileData struct {
	//配置文件数据
	SkillBaseData
	//跟等级相关的数值 逗号分隔
	CastRange  string //施法距离
	Cooldown   string //技能冷却时间
	HurtValue  string //技能伤害
	HurtRange  string //伤害范围 小于等于0表示单体
	NormalHurt string //附带普通攻击百分比 (0.5 为 50%的普通攻击伤害) 一般为0
	ManaCost   string //技能魔法消耗
}

//把等级相关的字符串 转成具体类型数据
func (this *SkillFileData) Trans2SkillData(re *[]SkillData) {
	if this.MaxLevel <= 0 {
		this.MaxLevel = 1
	}

	CastRange := utils.GetFloat32FromString2(this.CastRange)
	Cooldown := utils.GetFloat32FromString2(this.Cooldown)
	HurtValue := utils.GetInt32FromString2(this.HurtValue)
	HurtRange := utils.GetFloat32FromString2(this.HurtRange)
	NormalHurt := utils.GetFloat32FromString2(this.NormalHurt)
	ManaCost := utils.GetInt32FromString2(this.ManaCost)

	for i := int32(0); i < this.MaxLevel; i++ {
		ssd := SkillData{}
		ssd.SkillBaseData = this.SkillBaseData
		if int32(len(CastRange)) <= i {
			ssd.CastRange = CastRange[len(CastRange)-1]
		} else {
			ssd.CastRange = CastRange[i]
		}
		if int32(len(Cooldown)) <= i {
			ssd.Cooldown = Cooldown[len(Cooldown)-1]
		} else {
			ssd.Cooldown = Cooldown[i]
		}
		if int32(len(HurtValue)) <= i {
			ssd.HurtValue = HurtValue[len(HurtValue)-1]
		} else {
			ssd.HurtValue = HurtValue[i]
		}
		if int32(len(HurtRange)) <= i {
			ssd.HurtRange = HurtRange[len(HurtRange)-1]
		} else {
			ssd.HurtRange = HurtRange[i]
		}
		if int32(len(NormalHurt)) <= i {
			ssd.NormalHurt = NormalHurt[len(NormalHurt)-1]
		} else {
			ssd.NormalHurt = NormalHurt[i]
		}
		if int32(len(ManaCost)) <= i {
			ssd.ManaCost = ManaCost[len(ManaCost)-1]
		} else {
			ssd.ManaCost = ManaCost[i]
		}
		//log.Info("111-:%v--%d", ssd)
		*re = append(*re, ssd)

	}
}
