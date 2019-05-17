package utils

import (
	"bytes"
	"dq/log"
	"encoding/gob"
	"errors"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

//func FormatFloat(f float64, m int) string {
//	n := strconv.FormatFloat(f, 'f', -1, 32)
//	if n == "" {
//		return ""
//	}
//	if m >= len(n) {
//		return n
//	}
//	newn := strings.Split(n, ".")
//	if len(newn) < 2 || m >= len(newn[1]) {
//		return n
//	}
//	return newn[0] + "." + newn[1][:m]
//}

//检查随机概率是否命中
func CheckRandom(radio float32) bool {
	if rand.Intn(10000) < int(10000.0*radio) {
		return true
	}
	return false
}

//从字符串中获取数据 逗号分隔
func GetFloat32FromString(str string, params ...(*float32)) {
	str1 := strings.Split(str, ",")
	count := 0
	//log.Info("str1len:%d", len(str1))
	for _, v := range params {

		if len(str1) <= count {
			return
		}
		//log.Info("str1:%s", str1[count])

		value, err1 := strconv.ParseFloat(str1[count], 32)
		if err1 == nil {
			*v = float32(value)
			//log.Info("v:%f", v)
		}
		count++
	}
}

//从字符串中获取数据 逗号分隔
func GetFloat32FromString2(str string) []float32 {
	re := make([]float32, 0)
	str1 := strings.Split(str, ",")
	//log.Info("str1len:%d", len(str1))
	for _, v := range str1 {

		value, err1 := strconv.ParseFloat(v, 32)
		if err1 == nil {
			re = append(re, float32(value))
		}
	}
	return re
}

//从字符串中获取数据 逗号分隔
func GetInt32FromString2(str string) []int32 {
	re := make([]int32, 0)
	str1 := strings.Split(str, ",")
	//log.Info("str1len:%d", len(str1))
	for _, v := range str1 {

		value, err1 := strconv.Atoi(v)
		if err1 == nil {
			re = append(re, int32(value))
		}
	}
	return re
}
func GetInt32FromString(str string, params ...(*int32)) {
	str1 := strings.Split(str, ",")
	count := 0
	//log.Info("str1len:%d", len(str1))
	for _, v := range params {

		if len(str1) <= count {
			return
		}
		//log.Info("str1:%s", str1[count])

		value, err1 := strconv.Atoi(str1[count])
		if err1 == nil {
			*v = int32(value)
			//log.Info("v:%f", v)
		}
		count++
	}
}

func GetCurTimeOfSecond() float64 {
	return float64(time.Now().UnixNano()) / 1000000000.0
}

func Struct2Bytes(data interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}
func Bytes2Struct(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

func OpenXlsl(path string) *excelize.File {
	xlsx, err := excelize.OpenFile(path)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return xlsx
}

func ReadXlsxData(path string, data interface{}) (error, map[interface{}]interface{}) {
	re := make(map[interface{}]interface{})
	xlsx := OpenXlsl(path)
	if xlsx == nil {
		return errors.New("open fail " + path), nil
	}

	//
	//	fieldnames := make([]string, 0)
	//	for i := 0; i < reflect.ValueOf(datatype).NumField(); i++ {
	//		obj := reflect.TypeOf(datatype).Field(i)
	//		fieldnames := append(fieldnames, obj.Name)
	//	}

	datatype := reflect.TypeOf(data).Elem()

	nameandindex := make(map[int]string)
	firstrow := xlsx.GetRows("Sheet1")[0]
	for k, v := range firstrow {
		nameandindex[k] = v
	}

	for i := 1; i < len(xlsx.GetRows("Sheet1")); i++ {
		onedata := xlsx.GetRows("Sheet1")[i]

		person := reflect.New(datatype).Interface()
		//person := datatype
		pp := reflect.ValueOf(person) // 取得struct变量的指针
		key := 0
		for k, v := range nameandindex {
			//log.Info("k_val %d---%s", k, v)
			field := pp.Elem().FieldByName(v) //获取指定Field

			if field.Kind() == reflect.Int32 || field.Kind() == reflect.Int8 || field.Kind() == reflect.Int {
				val, err := strconv.ParseInt(onedata[k], 10, 64)
				if err == nil {
					field.SetInt(val)
				} else {
					field.SetInt(0)
				}

				if k == 0 {
					key = (int)(field.Int())
				}

			} else if field.Kind() == reflect.Float32 || field.Kind() == reflect.Float64 {
				val, err := strconv.ParseFloat(onedata[k], 64)
				if err == nil {
					field.SetFloat(val)
				} else {
					field.SetFloat(0)
				}

			} else if field.Kind() == reflect.String {
				field.SetString(onedata[k])
			}

		}

		re[key] = person

	}

	return nil, re
}
