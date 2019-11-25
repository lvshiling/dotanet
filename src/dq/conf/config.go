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
	"bufio"
	"bytes"
	"dq/log"
	"encoding/json"
	"fmt"
	//"fmt"
	"dq/vec2d"
	"io/ioutil"
	"os"
	"strings"
)

var (
	Conf         = Config{}
	SceneRawData = SceneAllData{}
	SceneData    = make(map[string]*Scene)

	CreateUnitRawData = CreateUnitAllData{}
	CreateUnitData    = make(map[string]*CreateUnits)
)

//场景配置文件
func LoadScene(Path string) {
	// Read config.

	ApplicationDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	confPath := fmt.Sprintf("%s"+Path, ApplicationDir)

	f, err := os.Open(confPath)
	if err != nil {
		panic(err)
	}

	err, data := readBigFileInto(f.Name())
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &SceneRawData)
	if err != nil {
		panic(err)
	}

	//---------
	for _, v := range SceneRawData.Scenes {
		SceneData[v.Name] = &v
	}

}

func GetSceneData(name string) *Scene {
	if _, ok := SceneData[name]; !ok {

		return nil
	}
	return SceneData[name]
}

//场景文件
type SceneAllData struct {
	Scenes []Scene
}
type Scene struct {
	Name     string
	Collides []Collide
}
type Collide struct {
	IsRect  bool
	CenterX float64
	CenterY float64
	Width   float64
	Height  float64
	Points  []vec2d.Vec2
}

//场景配置文件
func LoadCreateUnit(Path string) {
	// Read config.

	ApplicationDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	confPath := fmt.Sprintf("%s"+Path, ApplicationDir)

	f, err := os.Open(confPath)
	if err != nil {
		panic(err)
	}

	err, data := readBigFileInto(f.Name())
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &CreateUnitRawData)
	if err != nil {
		panic(err)
	}

	//---------
	for k, v := range CreateUnitRawData.CreateUnits {
		CreateUnitData[v.Name] = &CreateUnitRawData.CreateUnits[k]
		//log.Info("createunit:%v", v)
	}

	//	for k, v := range CreateUnitData {
	//		log.Info("createunit111:%s %v", k, v)
	//	}

}

func GetCreateUnitData(name string) *CreateUnits {
	log.Info("GetCreateUnitData:%s", name)
	if _, ok := CreateUnitData[name]; !ok {

		return nil
	}
	return CreateUnitData[name]
}

//创建单位文件
type CreateUnitAllData struct {
	CreateUnits []CreateUnits
}
type CreateUnits struct {
	Name     string
	Units    []Unit
	DoorWays []DoorWay
}
type Unit struct {
	TypeID       int32
	X            float64
	Y            float64
	Z            float64
	ReCreateTime float64
	Rotation     float64
}
type DoorWay struct {
	NextSceneID int32
	X           float64
	Y           float64
	Z           float64
	R           float64
	NeedLevel   int32
	NeedPlayer  int32
	NextX       float32
	NextY       float32
	HaloTypeID  int32
}

//基础配置文件
func LoadConfig(Path string) {
	// Read config.
	err, data := readFileInto(Path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &Conf)
	if err != nil {
		panic(err)
	}
}

type Config struct {
	GateInfo     gateInfo
	LoginInfo    map[string]interface{}
	HallInfo     map[string]interface{}
	Game5GInfo   map[string]interface{}
	DataBaseInfo map[string]interface{}
}
type gateInfo struct {
	ClientListenPort    string
	ClientTcpListenPort string
	ServerListenPort    string
	MaxConnNum          int
	PendingWriteNum     int
	TimeOut             int
}

func readBigFileInto(path string) (error, []byte) {
	var data []byte
	buf := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		return err, data
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 1024*1024)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if len(line) > 0 {
				buf.Write(line)
			}
			break
		}
		//处理注释
		if !strings.HasPrefix(strings.TrimLeft(string(line), "\t "), "//") {
			buf.Write(line)
		}
	}
	data = buf.Bytes()
	//log.Info(string(data))
	return nil, data
}

func readFileInto(path string) (error, []byte) {
	var data []byte
	buf := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		return err, data
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if len(line) > 0 {
				buf.Write(line)
			}
			break
		}
		//处理注释
		if !strings.HasPrefix(strings.TrimLeft(string(line), "\t "), "//") {
			buf.Write(line)
		}
	}
	data = buf.Bytes()
	//log.Info(string(data))
	return nil, data
}

// If read the file has an error,it will throws a panic.
func fileToStruct(path string, ptr *[]byte) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	*ptr = data
}
