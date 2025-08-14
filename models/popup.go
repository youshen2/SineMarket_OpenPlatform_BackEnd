package models

type Popup struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	ImgURL       string `gorm:"type:text;column:img_url" json:"img_url"`
	Actions      string `gorm:"type:text;column:actions" json:"actions"`
	ByUserID     int    `gorm:"column:by_userid" json:"by_userid"`
	SurplusCount int    `gorm:"column:surplus_count" json:"surplus_count"`
	SenderUserID int    `gorm:"column:sender_userid" json:"sender_userid"`
}

func (Popup) TableName() string {
	return "market_popup_list"
}
