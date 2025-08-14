package models

type Banner struct {
	ID         int    `gorm:"primaryKey;column:id" json:"id"`
	BannerURL  string `gorm:"column:banner" json:"banner_url"`
	Actions    string `gorm:"column:actions" json:"actions"`
	Visibility int    `gorm:"column:visibility" json:"visibility"`
}

func (Banner) TableName() string {
	return "market_banana_list"
}

type BannedIP struct {
	ID int    `gorm:"primaryKey;column:id" json:"id"`
	IP string `gorm:"column:ip" json:"ip"`
}

func (BannedIP) TableName() string {
	return "market_ip_ban_list"
}

type ProhibitedWord struct {
	ID      int    `gorm:"primaryKey;column:id" json:"id"`
	Content string `gorm:"column:content" json:"content"`
}

func (ProhibitedWord) TableName() string {
	return "market_prohibited_word_list"
}

type UsernameBlacklist struct {
	ID       int    `gorm:"primaryKey;column:id" json:"id"`
	Username string `gorm:"column:username" json:"username"`
}

func (UsernameBlacklist) TableName() string {
	return "market_username_blacklist"
}
