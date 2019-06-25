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
		t := GetSkillData(6, int32(i))
		if t != nil {
			log.Info("skill %d:%v", i, *t)
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

type CallUnitInfo struct {
	//召唤相关
	CallUnitCount     int32   //召唤数量 0表示没有召唤
	CallUnitTypeID    int32   //召唤出来的单位 类型ID 0表示当前召唤者 -1表示目标对象 其他类型id对应其他单位
	CallUnitBuff      string  //召唤出来的单位携带额外buff
	CallUnitHalo      string  //召唤出来的单位携带额外halo
	CallUnitOffsetPos float32 //召唤出来的单位在目标位置的随机偏移位置
	//CallUnitAliveTime float32 //召唤单位的生存时间
}

//技能基本数据
type SkillBaseData struct {
	TypeID                 int32   //类型ID
	CastType               int32   // 施法类型:  1:主动技能  2:被动技能
	CastTargetType         int32   //施法目标类型 1:自身为目标 2:以单位为目标 3:以地面1点为目标 4:攻击时自动释放(攻击特效) 5:以地面一点为方向
	CastTargetRange        float32 //施法目标范围 小于等于0表示单体 以施法目标点为中心的范围内的多个目标为 最终弹道目标
	UnitTargetTeam         int32   //目标单位关系 1:友方  2:敌方 3:友方敌方都行包括自己  4:友方敌方都行不包括自己
	UnitTargetCamp         int32   //目标单位阵营 (1:英雄 2:普通单位 3:远古 4:boss) 5:都行
	NoCareMagicImmune      int32   //无视技能免疫 (1:无视技能免疫 2:非)
	BulletModeType         string  //子弹模型
	HurtType               int32   //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害)
	TriggerAttackEffect    int32   //能否触发普通攻击特效 (1:触发 2:不触发)
	CastPoint              float32 //施法前摇(以施法时间为比列 0.5表示 施法的中间时间点触发)
	CastTime               float32 //施法时间(以秒为单位的时间 比如1秒)
	AnimotorState          int32   //动画
	RequiredLevel          int32   //初始等级需求 1级 需要玩家多少级才能学习
	LevelsBetweenUpgrades  int32   //等级需求步长 2
	MaxLevel               int32   //最高升级到的等级 5 表示能升级到5级
	InitLevel              int32   //技能初始等级
	Index                  int32   //技能索引 按升序排列  在屏幕右下角的显示位置
	Visible                int32   //技能是否显示 1:是 2:否
	VisibleTime            float32 //技能显示时间 -1表示永久
	UseToHide              int32   //1:释放技能的时候隐藏 2:释放技能的时候不隐藏
	VisibleRelationSkillID int32   //技能显示关联id 当使用本技能的时候 显示关联id的技能隐身本技能 0表示没有关联
	TargetBuff             string  //释放时 对目标造成的buff 比如 1,2 表示对目标造成typeid为 1和2的buff
	BlinkToTarget          int32   //是否瞬间移动到目的地 1:是 2:否
	MyBuff                 string  //释放时 对自己造成的buff 比如 1,2 表示对目标造成typeid为 1和2的buff
	InitBuff               string  //拥有技能技能时的buff (技能携带的buff)
	TargetHalo             string  //释放时 对目标造成的halo
	MyHalo                 string  //释放时 对自己造成的halo 比如 1,2 表示对目标造成typeid为 1和2的halo
	InitHalo               string  //拥有技能技能时的halo (技能携带的halo)
	PathHalo               string  //路径光环 在路径上创建光环
	PathHaloMinTime        float32 //路径光环的最短时间 1表示 相差1秒才创建光环
	MyClearLevel           int32   //释放时 对自己的驱散等级  能驱散 驱散等级 小于等于该值的buff
	TargetClearLevel       int32   //释放时 对目标的驱散等级  能驱散 驱散等级 小于等于该值的buff
	AwaysHurt              int32   //总是造成伤害 1:是 2:否
	CallUnitInfo                   //召唤信息

	//被动技能相关参数
	TriggerTime      int32 //触发时间 0:表示不触发 1:攻击时 2:被攻击时
	TriggerOtherRule int32 //触发需满足的额外条件 0:表示没有额外条件 1:表示范围内地方英雄不超过几个

	ForceMoveType int32  //强制移动类型 0:表示不强制移动 1:表示用子弹向后推开目标(小黑) 2:强制移动自己到指定位置
	ForceMoveBuff string //强制移动时的buff 随着移动结束消失

	//加血相关
	AddHPType   int32 //加血类型 0:不加 1:以AddHPValue为固定值 2:以AddHPValue为时间 加单位在此时间内受到的伤害值
	AddHPTarget int32 //加血的目标 1:表示自己 2:表示目标

	//互换位置
	SwitchedPlaces     int32 //互换位置 1:是 2:否 只对目标为单位的情况生效
	DestForceAttackSrc int32 //目标强制攻击施法者 1:是 2:否

	//特殊情况处理 //1:混沌间隙的目标和自己的瞬移
	Exception int32 //0表示没有特殊情况

	//

}

//技能数据 (会根据等级变化的数据)
type SkillData struct {
	SkillBaseData
	CastRange         float32 //施法距离
	Cooldown          float32 //技能冷却时间
	HurtValue         int32   //技能伤害
	HurtRange         float32 //伤害范围 小于等于0表示单体
	NormalHurt        float32 //附带普通攻击百分比 (0.5 为 50%的普通攻击伤害) 一般为0
	ManaCost          int32   //技能魔法消耗
	OtherManaCostVal  float32 //额外魔法消耗值
	OtherManaCostType int32   //额外魔法消耗类型
	BulletCount       int32   //子弹数量 仅对 对自己施法有效 在自己周围创造多个弹道
	SkillCount        int32   //技能点数
	EjectionCount     int32   //弹射次数
	EjectionRange     float32 //弹射范围
	EjectionDecay     float32 //弹射衰减

	//被动技能相关参数
	TriggerProbability float32 //触发几率 0.5表示50%
	TriggerCrit        float32 //触发的暴击 倍数 2.5表示2.5倍攻击 1表示正常攻击
	NoCareDodge        float32 //无视闪避几率
	PhysicalAmaorCV    int32   //物理护甲削弱 -7表示本次计算伤害减7点护甲  -10000表示本次计算伤害减光目标的基础护甲

	BulletSpeed float32 //子弹速度
	//强制移动相关
	ForceMoveTime      float32 //强制移动时间
	ForceMoveSpeedSize float32 //强制移动速度大小
	ForceMoveLevel     int32   //强制移动等级

	AddHPValue        float32 //加血值
	PhysicalHurtAddHP float32 //物理伤害吸血 0.1表示 增加攻击造成伤害的10%的HP
	MagicHurtAddHP    float32 //魔法伤害吸血 0.1表示 增加攻击造成伤害的10%的HP

	EveryDoHurtChangeHurtCR float32 //每对一个目标造成伤害后 伤害变化率 1表示没有变化 0.8表示递减20%

	ExceptionParam string //特殊情况处理参数

	TriggerOtherRuleParam string //触发需满足的额外条件参数

}

//单位配置文件数据
type SkillFileData struct {
	//配置文件数据
	SkillBaseData
	//跟等级相关的数值 逗号分隔
	CastRange         string //施法距离
	Cooldown          string //技能冷却时间
	HurtValue         string //技能伤害
	HurtRange         string //伤害范围 小于等于0表示单体
	NormalHurt        string //附带普通攻击百分比 (0.5 为 50%的普通攻击伤害) 一般为0
	ManaCost          string //技能魔法消耗
	OtherManaCostVal  string //额外魔法消耗值
	OtherManaCostType string //额外魔法消耗类型
	BulletCount       string //子弹数量 仅对 对自己施法有效 在自己周围创造多个弹道
	SkillCount        string //技能点数
	EjectionCount     string //弹射次数
	EjectionRange     string //弹射范围
	EjectionDecay     string //弹射衰减

	//被动技能相关参数
	TriggerProbability string //触发几率 0.5表示50%
	TriggerCrit        string //触发的暴击 倍数 2.5表示2.5倍攻击
	NoCareDodge        string //无视闪避几率
	PhysicalAmaorCV    string //物理护甲削弱 -7表示本次计算伤害减7点护甲  -10000表示本次计算伤害减光基础护甲

	BulletSpeed string //子弹速度
	//强制移动相关
	ForceMoveTime      string //强制移动时间
	ForceMoveSpeedSize string //强制移动速度大小
	ForceMoveLevel     string //强制移动等级

	AddHPValue        string //加血值
	PhysicalHurtAddHP string //物理伤害吸血 0.1表示 增加攻击造成伤害的10%的HP
	MagicHurtAddHP    string //魔法伤害吸血 0.1表示 增加攻击造成伤害的10%的HP

	EveryDoHurtChangeHurtCR string //每对一个目标造成伤害后 伤害变化率 1表示没有变化 0.8表示递减20%

	ExceptionParam        string //特殊情况处理参数
	TriggerOtherRuleParam string //触发需满足的额外条件参数
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
	OtherManaCostVal := utils.GetFloat32FromString2(this.OtherManaCostVal)
	OtherManaCostType := utils.GetInt32FromString2(this.OtherManaCostType)

	BulletCount := utils.GetInt32FromString2(this.BulletCount)
	SkillCount := utils.GetInt32FromString2(this.SkillCount)
	EjectionCount := utils.GetInt32FromString2(this.EjectionCount)
	EjectionRange := utils.GetFloat32FromString2(this.EjectionRange)
	EjectionDecay := utils.GetFloat32FromString2(this.EjectionDecay)

	//被动技能相关参数
	TriggerProbability := utils.GetFloat32FromString2(this.TriggerProbability)
	TriggerCrit := utils.GetFloat32FromString2(this.TriggerCrit)
	NoCareDodge := utils.GetFloat32FromString2(this.NoCareDodge)
	PhysicalAmaorCV := utils.GetInt32FromString2(this.PhysicalAmaorCV)

	BulletSpeed := utils.GetFloat32FromString2(this.BulletSpeed)
	//强制移动相关
	ForceMoveTime := utils.GetFloat32FromString2(this.ForceMoveTime)
	ForceMoveSpeedSize := utils.GetFloat32FromString2(this.ForceMoveSpeedSize)
	ForceMoveLevel := utils.GetInt32FromString2(this.ForceMoveLevel)

	AddHPValue := utils.GetFloat32FromString2(this.AddHPValue)
	PhysicalHurtAddHP := utils.GetFloat32FromString2(this.PhysicalHurtAddHP)
	MagicHurtAddHP := utils.GetFloat32FromString2(this.MagicHurtAddHP)

	EveryDoHurtChangeHurtCR := utils.GetFloat32FromString2(this.EveryDoHurtChangeHurtCR)

	ExceptionParam := utils.GetStringFromString2(this.ExceptionParam)
	TriggerOtherRuleParam := utils.GetStringFromString2(this.TriggerOtherRuleParam)

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
		if int32(len(OtherManaCostVal)) <= i {
			ssd.OtherManaCostVal = OtherManaCostVal[len(OtherManaCostVal)-1]
		} else {
			ssd.OtherManaCostVal = OtherManaCostVal[i]
		}
		if int32(len(OtherManaCostType)) <= i {
			ssd.OtherManaCostType = OtherManaCostType[len(OtherManaCostType)-1]
		} else {
			ssd.OtherManaCostType = OtherManaCostType[i]
		}
		if int32(len(BulletCount)) <= i {
			ssd.BulletCount = BulletCount[len(BulletCount)-1]
		} else {
			ssd.BulletCount = BulletCount[i]
		}
		if int32(len(SkillCount)) <= i {
			ssd.SkillCount = SkillCount[len(SkillCount)-1]
		} else {
			ssd.SkillCount = SkillCount[i]
		}
		if int32(len(EjectionCount)) <= i {
			ssd.EjectionCount = EjectionCount[len(EjectionCount)-1]
		} else {
			ssd.EjectionCount = EjectionCount[i]
		}
		if int32(len(EjectionRange)) <= i {
			ssd.EjectionRange = EjectionRange[len(EjectionRange)-1]
		} else {
			ssd.EjectionRange = EjectionRange[i]
		}
		if int32(len(EjectionDecay)) <= i {
			ssd.EjectionDecay = EjectionDecay[len(EjectionDecay)-1]
		} else {
			ssd.EjectionDecay = EjectionDecay[i]
		}

		if int32(len(TriggerProbability)) <= i {
			ssd.TriggerProbability = TriggerProbability[len(TriggerProbability)-1]
		} else {
			ssd.TriggerProbability = TriggerProbability[i]
		}
		if int32(len(TriggerCrit)) <= i {
			ssd.TriggerCrit = TriggerCrit[len(TriggerCrit)-1]
		} else {
			ssd.TriggerCrit = TriggerCrit[i]
		}
		if int32(len(NoCareDodge)) <= i {
			ssd.NoCareDodge = NoCareDodge[len(NoCareDodge)-1]
		} else {
			ssd.NoCareDodge = NoCareDodge[i]
		}
		if int32(len(PhysicalAmaorCV)) <= i {
			ssd.PhysicalAmaorCV = PhysicalAmaorCV[len(PhysicalAmaorCV)-1]
		} else {
			ssd.PhysicalAmaorCV = PhysicalAmaorCV[i]
		}

		if int32(len(BulletSpeed)) <= i {
			ssd.BulletSpeed = BulletSpeed[len(BulletSpeed)-1]
		} else {
			ssd.BulletSpeed = BulletSpeed[i]
		}
		if int32(len(ForceMoveTime)) <= i {
			ssd.ForceMoveTime = ForceMoveTime[len(ForceMoveTime)-1]
		} else {
			ssd.ForceMoveTime = ForceMoveTime[i]
		}
		if int32(len(ForceMoveSpeedSize)) <= i {
			ssd.ForceMoveSpeedSize = ForceMoveSpeedSize[len(ForceMoveSpeedSize)-1]
		} else {
			ssd.ForceMoveSpeedSize = ForceMoveSpeedSize[i]
		}
		if int32(len(ForceMoveLevel)) <= i {
			ssd.ForceMoveLevel = ForceMoveLevel[len(ForceMoveLevel)-1]
		} else {
			ssd.ForceMoveLevel = ForceMoveLevel[i]
		}

		if int32(len(AddHPValue)) <= i {
			ssd.AddHPValue = AddHPValue[len(AddHPValue)-1]
		} else {
			ssd.AddHPValue = AddHPValue[i]
		}
		if int32(len(PhysicalHurtAddHP)) <= i {
			ssd.PhysicalHurtAddHP = PhysicalHurtAddHP[len(PhysicalHurtAddHP)-1]
		} else {
			ssd.PhysicalHurtAddHP = PhysicalHurtAddHP[i]
		}
		if int32(len(MagicHurtAddHP)) <= i {
			ssd.MagicHurtAddHP = MagicHurtAddHP[len(MagicHurtAddHP)-1]
		} else {
			ssd.MagicHurtAddHP = MagicHurtAddHP[i]
		}
		if int32(len(EveryDoHurtChangeHurtCR)) <= i {
			ssd.EveryDoHurtChangeHurtCR = EveryDoHurtChangeHurtCR[len(EveryDoHurtChangeHurtCR)-1]
		} else {
			ssd.EveryDoHurtChangeHurtCR = EveryDoHurtChangeHurtCR[i]
		}

		if int32(len(ExceptionParam)) <= i {
			ssd.ExceptionParam = ExceptionParam[len(ExceptionParam)-1]
		} else {
			ssd.ExceptionParam = ExceptionParam[i]
		}

		if int32(len(TriggerOtherRuleParam)) <= i {
			ssd.TriggerOtherRuleParam = TriggerOtherRuleParam[len(TriggerOtherRuleParam)-1]
		} else {
			ssd.TriggerOtherRuleParam = TriggerOtherRuleParam[i]
		}

		//log.Info("111-:%v--%d", ssd)
		*re = append(*re, ssd)

	}
}
