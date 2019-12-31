package db

//玩家角色信息
type DB_CharacterInfo struct {
	Characterid           int32   `json:"characterid"`
	Uid                   int32   `json:"uid"`
	Name                  string  `json:"name"`
	Typeid                int32   `json:"typeid"`
	Level                 int32   `json:"level"`
	Experience            int32   `json:"experience"`
	Gold                  int32   `json:"gold"`
	Diamond               int32   `json:"diamond"`
	HP                    float32 `json:"hp"`
	MP                    float32 `json:"mp"`
	SceneID               int32   `json:"sceneid"`
	SceneName             string  `json:"scenename"`
	X                     float32 `json:"x"`
	Y                     float32 `json:"y"`
	Skill                 string  `json:"skill"`
	Item1                 string  `json:"item1"`
	Item2                 string  `json:"item2"`
	Item3                 string  `json:"item3"`
	Item4                 string  `json:"item4"`
	Item5                 string  `json:"item5"`
	Item6                 string  `json:"item6"`
	BagInfo               string  `json:"baginfo"`
	ItemSkillCDInfo       string  `json:"itemskillcd"`
	GetExperienceDay      string  `json:"getexperienceday"`
	RemainExperience      int32   `json:"remainexperience"`
	RemainReviveTime      float32 `json:"remainerevivetime"`
	KillCount             int32   `json:"killcount"`
	ContinuityKillCount   int32   `json:"continuitykillcount"`
	DieCount              int32   `json:"diecount"`
	KillGetGold           int32   `json:"killgetgold"`
	Friends               string  `json:"friends"`
	FriendsRequest        string  `json:"friendsrequest"`
	WatchVedioCountOneDay int32   `json:"watchvediocountoneday"`
	Mails                 string  `json:"mails"`
}

//角色邮件信息
type DB_MailInfo struct {
	Id             int32  `json:"id"`
	Sendname       string `json:"sendname"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	RecUid         int32  `json:"recUid"`
	RecCharacterid int32  `json:"recCharacterid"`
	Date           string `json:"date"`
	Rewardstr      string `json:"rewardstr"`
	Getstate       int32  `json:"getstate"`
}
