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

type Report struct {
	ID            int    `gorm:"primaryKey;column:id" json:"id"`
	ByUserID      int    `gorm:"column:by_userid" json:"by_userid"`
	ReportType    int    `gorm:"column:report_type" json:"report_type"`
	ReportID      int    `gorm:"column:report_id" json:"report_id"`
	ReportReason  string `gorm:"column:report_reason" json:"report_reason"`
	ReportTitle   string `gorm:"column:report_title" json:"report_title"`
	ReportStatus  int    `gorm:"column:report_status" json:"report_status"`
	ReportReply   string `gorm:"column:report_reply" json:"report_reply"`
	ReportTime    int64  `gorm:"column:report_time" json:"report_time"`
	ReplyTime     int64  `gorm:"column:reply_time" json:"reply_time"`
	ReportDevice  string `gorm:"column:report_device" json:"report_device"`
	ReportIP      string `gorm:"column:report_ip" json:"report_ip"`
	ReportVersion int    `gorm:"column:report_version" json:"report_version"`
	Reporter      User   `gorm:"foreignKey:ByUserID" json:"reporter"`
}

func (Report) TableName() string {
	return "market_report_list"
}
