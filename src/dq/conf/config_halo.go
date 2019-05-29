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
	HaloFileDatas = make(map[interface{}]interface{})

	//key 为 id_level
	HaloDatas = make(map[interface{}]interface{})
)

//场景配置文件
func LoadHaloFileData() {
	_, HaloFileDatas = utils.ReadXlsxData("bin/conf/halo.xlsx", (*HaloFileData)(nil))

	InitHaloDatas()

}

//初始化具体技能数据
func InitHaloDatas() {

	for _, v := range HaloFileDatas {
		ssd := make([]HaloData, 0)
		v.(*HaloFileData).Trans2HaloData(&ssd)

		for k, v1 := range ssd {
			test := v1
			HaloDatas[string(v1.TypeID)+"_"+string(k+1)] = &test
		}
	}

	//log.Info("----------1---------")

	//log.Info("-:%v", HaloDatas)
	for i := 1; i < 5; i++ {
		t := GetHaloData(1000, int32(i))
		if t != nil {
			log.Info("halo %d:%v", i, *t)
		}

	}

	//log.Info("----------2---------")
}

//获取技能数据 通过技能ID和等级
func GetHaloData(typeid int32, level int32) *HaloData {
	//log.Info("find unitfile:%d", typeid)
	if level <= 0 {
		level = 1
	}
	key := string(typeid) + "_" + string(level)

	re := (HaloDatas[key])
	if re == nil {
		log.Info("not find skill unitfile:%d", typeid)
		return nil
	}
	return (re).(*HaloData)
}

//技能基本数据
type HaloBaseData struct {
	TypeID int32 //类型ID

	UnitTargetTeam      int32   //目标单位关系 1:友方  2:敌方 3:友方敌方都行
	NoCareMagicImmune   int32   //无视技能免疫 (1:无视技能免疫 2:非)
	BulletModeType      string  //子弹模型
	BulletSpeed         float32 //子弹速度
	HurtType            int32   //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害 4:不造成伤害)
	TriggerAttackEffect int32   //能否触发普通攻击特效 (1:触发 2:不触发)
	MaxLevel            int32   //最高升级到的等级 5 表示能升级到5级
	TargetBuff          string  //对施法目标造成的buff 比如 1,2 表示对目标造成typeid为 1和2的buff
	InitBuff            string  //对范围内目标施加的buf 持续0.1s
	FollowParent        int32   //跟随主角  1:是 2:否
	HaloModeType        string  //光环模型
}

//技能数据 (会根据等级变化的数据)
type HaloData struct {
	HaloBaseData
	UnitTargetMaxCount int32   //最大选择目标数量
	Time               float32 //持续时间
	Cooldown           float32 //技能冷却时间 为施法间隔
	HurtValue          int32   //技能伤害
	HaloRange          float32 //光环范围 小于等于0表示单体
	NormalHurt         float32 //附带普通攻击百分比 (0.5 为 50%的普通攻击伤害) 一般为0

}

//单位配置文件数据
type HaloFileData struct {
	//配置文件数据
	HaloBaseData
	//跟等级相关的数值 逗号分隔
	UnitTargetMaxCount string //最大选择目标数量
	Time               string //持续时间
	Cooldown           string //技能冷却时间 为施法间隔
	HurtValue          string //技能伤害
	HaloRange          string //光环范围 小于等于0表示单体
	NormalHurt         string //附带普通攻击百分比 (0.5 为 50%的普通攻击伤害) 一般为0

}

//把等级相关的字符串 转成具体类型数据
func (this *HaloFileData) Trans2HaloData(re *[]HaloData) {
	if this.MaxLevel <= 0 {
		this.MaxLevel = 1
	}
	UnitTargetMaxCount := utils.GetInt32FromString2(this.UnitTargetMaxCount)
	Time := utils.GetFloat32FromString2(this.Time)
	Cooldown := utils.GetFloat32FromString2(this.Cooldown)
	HurtValue := utils.GetInt32FromString2(this.HurtValue)
	HaloRange := utils.GetFloat32FromString2(this.HaloRange)
	NormalHurt := utils.GetFloat32FromString2(this.NormalHurt)

	for i := int32(0); i < this.MaxLevel; i++ {
		ssd := HaloData{}
		ssd.HaloBaseData = this.HaloBaseData
		if int32(len(UnitTargetMaxCount)) <= i {
			ssd.UnitTargetMaxCount = UnitTargetMaxCount[len(UnitTargetMaxCount)-1]
		} else {
			ssd.UnitTargetMaxCount = UnitTargetMaxCount[i]
		}
		if int32(len(Time)) <= i {
			ssd.Time = Time[len(Time)-1]
		} else {
			ssd.Time = Time[i]
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
		if int32(len(HaloRange)) <= i {
			ssd.HaloRange = HaloRange[len(HaloRange)-1]
		} else {
			ssd.HaloRange = HaloRange[i]
		}
		if int32(len(NormalHurt)) <= i {
			ssd.NormalHurt = NormalHurt[len(NormalHurt)-1]
		} else {
			ssd.NormalHurt = NormalHurt[i]
		}

		//log.Info("111-:%v--%d", ssd)
		*re = append(*re, ssd)

	}
}
