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
	BuffFileDatas = make(map[interface{}]interface{})

	//key 为 id_level
	BuffDatas = make(map[interface{}]interface{})
)

//场景配置文件
func LoadBuffFileData() {
	_, BuffFileDatas = utils.ReadXlsxData("bin/conf/buff.xlsx", (*BuffFileData)(nil))

	InitBuffDatas()

}

//初始化具体技能数据
func InitBuffDatas() {

	for _, v := range BuffFileDatas {
		ssd := make([]BuffData, 0)
		v.(*BuffFileData).Trans2BuffData(&ssd)

		for k, v1 := range ssd {
			test := v1
			BuffDatas[string(v1.TypeID)+"_"+string(k+1)] = &test
		}
	}

	//log.Info("----------1---------")

	//log.Info("-:%v", SkillDatas)
	for i := 1; i < 5; i++ {
		t := GetBuffData(1, int32(i))
		if t != nil {
			log.Info("buff %d:%v", i, *t)
		}

	}

	//log.Info("----------2---------")
}

//获取技能数据 通过技能ID和等级
func GetBuffData(typeid int32, level int32) *BuffData {
	//log.Info("find unitfile:%d", typeid)
	if level <= 0 {
		level = 1
	}
	key := string(typeid) + "_" + string(level)

	re := (BuffDatas[key])
	if re == nil {
		log.Info("not find buff unitfile:%d", typeid)
		return nil
	}
	return (re).(*BuffData)
}

//技能基本数据
type BuffBaseData struct {
	TypeID int32 //类型ID

	//HurtType              int32   //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害)
	MaxLevel               int32 //最高升级到的等级 5 表示能升级到5级
	OverlyingType          int32 //叠加类型 1:只更新最大时间 2:完美叠加(小鱼的偷属性)
	OverlyingAddTag        int32 //叠加时是否增加标记数字 1:表示增加 2:表示不增加
	ActiveUnitAcpabilities int32 //生效的单位共计类型(1:近程攻击 2:远程攻击 3:都生效)

	NoMove      int32 //禁止移动 1:是 2:非
	NoTurn      int32 //禁止转向 1:是 2:非
	NoAttack    int32 //禁止攻击 1:是 2:非
	NoSkill     int32 //禁止使用技能 1:是 2:非
	NoItem      int32 //禁止使用道具 1:是 2:非
	MagicImmune int32 //是否魔法免疫 1:是 2:非

	Invisible int32 //隐身 1:是 2:否  可以躲避攻击弹道 并且从显示屏上消失

	ActiveTime float32 //开始生效的时间 1.2表示 1.2秒后生效

	AttackedInvalid  int32 //攻击后失效 1:是 2:否
	DoSkilledInvalid int32 //使用技能后失效 1:是 2:否
}

//技能数据 (会根据等级变化的数据)
type BuffData struct {
	BuffBaseData
	//HurtValue  int32   //技能伤害
	BuffRange float32 //buff范围 小于等于0表示单体
	Time      float32 //持续时间

	//CV表示变化量  CR表示变化比率
	AttributeStrengthCV       float32 //力量变化值 20表示增加20点力量
	AttributeIntelligenceCV   float32 //智力变化值 20表示增加20点智力
	AttributeAgilityCV        float32 //敏捷变化值 20表示增加20点敏捷
	AttackSpeedCV             float32 //攻击速度变化值 -10表示降低10点攻击速度
	AttackCR                  float32 //攻击力变化比率 0.2就是增加20%
	AttackCV                  float32 //攻击力变化量 -20就是减少20点攻击力
	AttackRangeCV             float32 //攻击距离变化量 1.2就是增加1.2的攻击距离
	MoveSpeedCR               float32 //移动速度变化率 1.0就是增加100%的移动速度
	MoveSpeedCV               float32 //移动速度变化量 0.5就是增加0.5米每秒的移动速度
	MagicScaleCV              float32 //技能增强变化量 0.02表示技能增强增加2%
	MPRegainCR                float32 //魔法恢复变化率 0.5表示增加50%的魔法恢复
	MPRegainCV                float32 //魔法恢复变化量 1.5表示增加1.5mp每秒的魔法恢复速度
	PhysicalAmaorCR           float32 //护甲变化比率 0.2就是增加20%
	PhysicalAmaorCV           float32 //护甲变化量 20就是增加20点护甲
	MagicAmaorCV              float32 //魔法抗性变化量 0.1表示增加10%的魔抗
	StatusAmaorCV             float32 //状态抗性变化量 0.1表示增加10%的状态抗性
	DodgeCV                   float32 //闪避变化量 0.2就是增加20%的闪避
	HPRegainCR                float32 //生命恢复变化率 0.5表示增加50%的生命恢复
	HPRegainCV                float32 //生命恢复变化量 1.2表示增加1.2hp每秒的生命恢复
	NoCareDodgeCV             float32 //无视闪避变化量 0.2就是增加20%的无视闪避
	AddedMagicRangeCV         float32 //额外施法距离变化量 2.3表示增加施法距离2.3米
	ManaCostCV                float32 //魔法消耗变化量 -0.1表示降低10%的魔法消耗
	MagicCDCV                 float32 //技能CD变化量 -0.2表示降低20%的技能cd
	AttackTargetAttackSpeedCV float32 //攻击指定目标攻击速度变化值 -10表示降低10点攻击速度
}

//单位配置文件数据
type BuffFileData struct {
	//配置文件数据
	BuffBaseData
	//跟等级相关的数值 逗号分隔
	BuffRange string //伤害范围 小于等于0表示单体
	Time      string //持续时间

	//CV表示变化量  CR表示变化比率
	AttributeStrengthCV       string //力量变化值 20表示增加20点力量
	AttributeIntelligenceCV   string //智力变化值 20表示增加20点智力
	AttributeAgilityCV        string //敏捷变化值 20表示增加20点敏捷
	AttackSpeedCV             string //攻击速度变化值 -10表示降低10点攻击速度
	AttackCR                  string //攻击力变化比率 0.2就是增加20%
	AttackCV                  string //攻击力变化量 -20就是减少20点攻击力
	AttackRangeCV             string //攻击距离变化量 1.2就是增加1.2的攻击距离
	MoveSpeedCR               string //移动速度变化率 1.0就是增加100%的移动速度
	MoveSpeedCV               string //移动速度变化量 0.5就是增加0.5米每秒的移动速度
	MagicScaleCV              string //技能增强变化量 0.02表示技能增强增加2%
	MPRegainCR                string //魔法恢复变化率 0.5表示增加50%的魔法恢复
	MPRegainCV                string //魔法恢复变化量 1.5表示增加1.5mp每秒的魔法恢复速度
	PhysicalAmaorCR           string //护甲变化比率 0.2就是增加20%
	PhysicalAmaorCV           string //护甲变化量 20就是增加20点护甲
	MagicAmaorCV              string //魔法抗性变化量 0.1表示增加10%的魔抗
	StatusAmaorCV             string //状态抗性变化量 0.1表示增加10%的状态抗性
	DodgeCV                   string //闪避变化量 0.2就是增加20%的闪避
	HPRegainCR                string //生命恢复变化率 0.5表示增加50%的生命恢复
	HPRegainCV                string //生命恢复变化量 1.2表示增加1.2hp每秒的生命恢复
	NoCareDodgeCV             string //无视闪避变化量 0.2就是增加20%的无视闪避
	AddedMagicRangeCV         string //额外施法距离变化量 2.3表示增加施法距离2.3米
	ManaCostCV                string //魔法消耗变化量 -0.1表示降低10%的魔法消耗
	MagicCDCV                 string //技能CD变化量 -0.2表示降低20%的技能cd
	AttackTargetAttackSpeedCV string //攻击指定目标攻击速度变化值 -10表示降低10点攻击速度
}

//把等级相关的字符串 转成具体类型数据
func (this *BuffFileData) Trans2BuffData(re *[]BuffData) {
	if this.MaxLevel <= 0 {
		this.MaxLevel = 1
	}

	BuffRange := utils.GetFloat32FromString2(this.BuffRange)
	Time := utils.GetFloat32FromString2(this.Time)
	AttributeStrengthCV := utils.GetFloat32FromString2(this.AttributeStrengthCV)
	AttributeIntelligenceCV := utils.GetFloat32FromString2(this.AttributeIntelligenceCV)
	AttributeAgilityCV := utils.GetFloat32FromString2(this.AttributeAgilityCV)
	AttackSpeedCV := utils.GetFloat32FromString2(this.AttackSpeedCV)
	AttackCR := utils.GetFloat32FromString2(this.AttackCR)
	AttackCV := utils.GetFloat32FromString2(this.AttackCV)
	AttackRangeCV := utils.GetFloat32FromString2(this.AttackRangeCV)
	MoveSpeedCR := utils.GetFloat32FromString2(this.MoveSpeedCR)
	MoveSpeedCV := utils.GetFloat32FromString2(this.MoveSpeedCV)
	MagicScaleCV := utils.GetFloat32FromString2(this.MagicScaleCV)
	MPRegainCR := utils.GetFloat32FromString2(this.MPRegainCR)
	MPRegainCV := utils.GetFloat32FromString2(this.MPRegainCV)
	PhysicalAmaorCR := utils.GetFloat32FromString2(this.PhysicalAmaorCR)
	PhysicalAmaorCV := utils.GetFloat32FromString2(this.PhysicalAmaorCV)
	MagicAmaorCV := utils.GetFloat32FromString2(this.MagicAmaorCV)
	StatusAmaorCV := utils.GetFloat32FromString2(this.StatusAmaorCV)
	DodgeCV := utils.GetFloat32FromString2(this.DodgeCV)
	HPRegainCR := utils.GetFloat32FromString2(this.HPRegainCR)
	HPRegainCV := utils.GetFloat32FromString2(this.HPRegainCV)
	NoCareDodgeCV := utils.GetFloat32FromString2(this.NoCareDodgeCV)
	AddedMagicRangeCV := utils.GetFloat32FromString2(this.AddedMagicRangeCV)
	ManaCostCV := utils.GetFloat32FromString2(this.ManaCostCV)
	MagicCDCV := utils.GetFloat32FromString2(this.MagicCDCV)
	AttackTargetAttackSpeedCV := utils.GetFloat32FromString2(this.AttackTargetAttackSpeedCV)

	for i := int32(0); i < this.MaxLevel; i++ {
		ssd := BuffData{}
		ssd.BuffBaseData = this.BuffBaseData

		if int32(len(BuffRange)) <= i {
			ssd.BuffRange = BuffRange[len(BuffRange)-1]
		} else {
			ssd.BuffRange = BuffRange[i]
		}
		if int32(len(Time)) <= i {
			ssd.Time = Time[len(Time)-1]
		} else {
			ssd.Time = Time[i]
		}
		if int32(len(AttributeStrengthCV)) <= i {
			ssd.AttributeStrengthCV = AttributeStrengthCV[len(AttributeStrengthCV)-1]
		} else {
			ssd.AttributeStrengthCV = AttributeStrengthCV[i]
		}
		if int32(len(AttributeIntelligenceCV)) <= i {
			ssd.AttributeIntelligenceCV = AttributeIntelligenceCV[len(AttributeIntelligenceCV)-1]
		} else {
			ssd.AttributeIntelligenceCV = AttributeIntelligenceCV[i]
		}
		if int32(len(AttributeAgilityCV)) <= i {
			ssd.AttributeAgilityCV = AttributeAgilityCV[len(AttributeAgilityCV)-1]
		} else {
			ssd.AttributeAgilityCV = AttributeAgilityCV[i]
		}
		if int32(len(AttackSpeedCV)) <= i {
			ssd.AttackSpeedCV = AttackSpeedCV[len(AttackSpeedCV)-1]
		} else {
			ssd.AttackSpeedCV = AttackSpeedCV[i]
		}
		if int32(len(AttackCR)) <= i {
			ssd.AttackCR = AttackCR[len(AttackCR)-1]
		} else {
			ssd.AttackCR = AttackCR[i]
		}
		if int32(len(AttackCV)) <= i {
			ssd.AttackCV = AttackCV[len(AttackCV)-1]
		} else {
			ssd.AttackCV = AttackCV[i]
		}
		if int32(len(AttackRangeCV)) <= i {
			ssd.AttackRangeCV = AttackRangeCV[len(AttackRangeCV)-1]
		} else {
			ssd.AttackRangeCV = AttackRangeCV[i]
		}
		if int32(len(MoveSpeedCR)) <= i {
			ssd.MoveSpeedCR = MoveSpeedCR[len(MoveSpeedCR)-1]
		} else {
			ssd.MoveSpeedCR = MoveSpeedCR[i]
		}
		if int32(len(MoveSpeedCV)) <= i {
			ssd.MoveSpeedCV = MoveSpeedCV[len(MoveSpeedCV)-1]
		} else {
			ssd.MoveSpeedCV = MoveSpeedCV[i]
		}
		if int32(len(MagicScaleCV)) <= i {
			ssd.MagicScaleCV = MagicScaleCV[len(MagicScaleCV)-1]
		} else {
			ssd.MagicScaleCV = MagicScaleCV[i]
		}
		if int32(len(MPRegainCR)) <= i {
			ssd.MPRegainCR = MPRegainCR[len(MPRegainCR)-1]
		} else {
			ssd.MPRegainCR = MPRegainCR[i]
		}
		if int32(len(MPRegainCV)) <= i {
			ssd.MPRegainCV = MPRegainCV[len(MPRegainCV)-1]
		} else {
			ssd.MPRegainCV = MPRegainCV[i]
		}
		if int32(len(PhysicalAmaorCR)) <= i {
			ssd.PhysicalAmaorCR = PhysicalAmaorCR[len(PhysicalAmaorCR)-1]
		} else {
			ssd.PhysicalAmaorCR = PhysicalAmaorCR[i]
		}
		if int32(len(PhysicalAmaorCV)) <= i {
			ssd.PhysicalAmaorCV = PhysicalAmaorCV[len(PhysicalAmaorCV)-1]
		} else {
			ssd.PhysicalAmaorCV = PhysicalAmaorCV[i]
		}
		if int32(len(MagicAmaorCV)) <= i {
			ssd.MagicAmaorCV = MagicAmaorCV[len(MagicAmaorCV)-1]
		} else {
			ssd.MagicAmaorCV = MagicAmaorCV[i]
		}
		if int32(len(StatusAmaorCV)) <= i {
			ssd.StatusAmaorCV = StatusAmaorCV[len(StatusAmaorCV)-1]
		} else {
			ssd.StatusAmaorCV = StatusAmaorCV[i]
		}
		if int32(len(DodgeCV)) <= i {
			ssd.DodgeCV = DodgeCV[len(DodgeCV)-1]
		} else {
			ssd.DodgeCV = DodgeCV[i]
		}
		if int32(len(HPRegainCR)) <= i {
			ssd.HPRegainCR = HPRegainCR[len(HPRegainCR)-1]
		} else {
			ssd.HPRegainCR = HPRegainCR[i]
		}
		if int32(len(HPRegainCV)) <= i {
			ssd.HPRegainCV = HPRegainCV[len(HPRegainCV)-1]
		} else {
			ssd.HPRegainCV = HPRegainCV[i]
		}
		if int32(len(NoCareDodgeCV)) <= i {
			ssd.NoCareDodgeCV = NoCareDodgeCV[len(NoCareDodgeCV)-1]
		} else {
			ssd.NoCareDodgeCV = NoCareDodgeCV[i]
		}
		if int32(len(AddedMagicRangeCV)) <= i {
			ssd.AddedMagicRangeCV = AddedMagicRangeCV[len(AddedMagicRangeCV)-1]
		} else {
			ssd.AddedMagicRangeCV = AddedMagicRangeCV[i]
		}
		if int32(len(ManaCostCV)) <= i {
			ssd.ManaCostCV = ManaCostCV[len(ManaCostCV)-1]
		} else {
			ssd.ManaCostCV = ManaCostCV[i]
		}
		if int32(len(MagicCDCV)) <= i {
			ssd.MagicCDCV = MagicCDCV[len(MagicCDCV)-1]
		} else {
			ssd.MagicCDCV = MagicCDCV[i]
		}
		if int32(len(AttackTargetAttackSpeedCV)) <= i {
			ssd.AttackTargetAttackSpeedCV = AttackTargetAttackSpeedCV[len(AttackTargetAttackSpeedCV)-1]
		} else {
			ssd.AttackTargetAttackSpeedCV = AttackTargetAttackSpeedCV[i]
		}

		//log.Info("111-:%v--%d", ssd)
		*re = append(*re, ssd)

	}
}
