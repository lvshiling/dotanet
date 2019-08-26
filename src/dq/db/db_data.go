package db

type DB_CharacterInfo struct {
	Characterid int32   `json:"characterid"`
	Uid         int32   `json:"uid"`
	Name        string  `json:"name"`
	Typeid      int32   `json:"typeid"`
	Level       int32   `json:"level"`
	Experience  int32   `json:"experience"`
	Gold        int32   `json:"gold"`
	HP          float32 `json:"hp"`
	MP          float32 `json:"mp"`
	SceneID     int32   `json:"sceneid"`
	SceneName   string  `json:"scenename"`
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Skill       string  `json:"skill"`
	Item1       string  `json:"item1"`
	Item2       string  `json:"item2"`
	Item3       string  `json:"item3"`
	Item4       string  `json:"item4"`
	Item5       string  `json:"item5"`
	Item6       string  `json:"item6"`
	BagInfo     string  `json:"baginfo"`
}
