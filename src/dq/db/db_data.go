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
	SceneName   string  `json:"scenename"`
	X           float32 `json:"x"`
	Y           float32 `json:"y"`
	Skill       string  `json:"skill"`
}
