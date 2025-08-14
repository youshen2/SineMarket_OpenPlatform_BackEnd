package models

type UserAction struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	ByUserID     int    `gorm:"column:by_userid" json:"by_userid"`
	Actions      string `gorm:"type:text;column:actions" json:"actions"`
	SurplusCount int    `gorm:"column:surplus_count" json:"surplus_count"`
}

func (UserAction) TableName() string {
	return "market_user_action_list"
}
