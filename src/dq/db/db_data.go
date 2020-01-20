package db

//"encoding/json"

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
	GuildId               int32   `json:"guildid"`
	GuildPinLevel         int32   `json:"guildpinlevel"`
	GuildPinExperience    int32   `json:"guildpinexperience"`
	GuildPost             int32   `json:"guildpost"`
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

//玩家之间的物品交易
type DB_PlayerItemTransactionInfo struct {
	Id                int32 `json:"id"`
	ItemID            int32 `json:"itemid"`
	Level             int32 `json:"level"`
	PriceType         int32 `json:"pricetype"`         //价格类型 1金币 2砖石
	Price             int32 `json:"price"`             //价格
	SellerUid         int32 `json:"sellerUid"`         //卖家UID(账号ID)
	SellerCharacterid int32 `json:"sellerCharacterid"` //卖家角色ID
	ShelfTime         int32 `json:"shelftime"`         //上架时间(秒)
}

//公会数据
type DB_GuildInfo struct {
	Id                   int32  `json:"id"`
	PresidentCharacterid int32  `json:"presidentCharacterid"`
	Level                int32  `json:"level"`
	Experience           int32  `json:"experience"`
	Createday            string `json:"createday"`
	Name                 string `json:"name"`
	Notice               string `json:"notice"`
	Joinaudit            int32  `json:"joinaudit"`
	Joinlevellimit       int32  `json:"joinlevellimit"`
	Characters           string `json:"characters"`
}
