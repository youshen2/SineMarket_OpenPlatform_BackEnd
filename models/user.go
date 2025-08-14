package models

type User struct {
	ID               int    `gorm:"primaryKey;column:id" json:"id"`
	Username         string `gorm:"column:username" json:"username"`
	Password         string `gorm:"column:password" json:"-"`
	DisplayName      string `gorm:"column:display_name" json:"display_name"`
	UserDescribe     string `gorm:"column:user_describe" json:"user_describe"`
	UserOfficial     string `gorm:"column:user_official" json:"user_official"`
	UserAvatar       string `gorm:"column:user_avatar" json:"user_avatar"`
	UserBadge        string `gorm:"column:user_badge" json:"user_badge"`
	UserStatus       int    `gorm:"column:user_status" json:"user_status"`
	UserStatusReason string `gorm:"column:user_status_reason" json:"user_status_reason"`
	BanTime          int64  `gorm:"column:ban_time" json:"ban_time"`
	JoinTime         int64  `gorm:"column:join_time" json:"join_time"`
	UserPermission   int    `gorm:"column:user_permission" json:"user_permission"`
	BindQQ           int64  `gorm:"column:bind_qq" json:"bind_qq"`
	BindEmail        string `gorm:"column:bind_email" json:"bind_email"`
	BindBilibili     int64  `gorm:"column:bind_bilibili" json:"bind_bilibili"`
	VerifyEmail      int    `gorm:"column:verify_email" json:"verify_email"`
	RegisterIP       string `gorm:"column:register_ip" json:"register_ip"`
	LastLoginIP      string `gorm:"column:last_login_ip" json:"last_login_ip"`
	LastLoginDevice  string `gorm:"column:last_login_device" json:"last_login_device"`
	LastLoginVersion int    `gorm:"column:last_login_version" json:"last_login_version"`
	LastOnlineTime   int64  `gorm:"column:last_online_time" json:"last_online_time"`
	PubFavourite     int    `gorm:"column:pub_favourite" json:"pub_favourite"`
}

func (User) TableName() string {
	return "market_user_list"
}

type UserToken struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	Token        string `gorm:"type:text;column:token" json:"token"` // Changed to text to store JWT
	LoginDevice  string `gorm:"column:login_device" json:"login_device"`
	LoginVersion int    `gorm:"column:login_version" json:"login_version"`
	LoginIP      string `gorm:"column:login_ip" json:"login_ip"`
	CreateTime   int64  `gorm:"column:create_time" json:"create_time"`
	LoginTime    int64  `gorm:"column:login_time" json:"login_time"`
	ExpireTime   int64  `gorm:"column:expire_time" json:"expire_time"`
	ByUserID     int    `gorm:"column:by_userid" json:"by_userid"`
	Status       int    `gorm:"column:status" json:"status"`
}

func (UserToken) TableName() string {
	return "market_user_token_list"
}

type Notice struct {
	ID           int    `gorm:"primaryKey;column:id" json:"id"`
	ByUserID     int    `gorm:"column:by_userid" json:"by_userid"`
	SenderUserID int    `gorm:"column:sender_userid" json:"sender_userid"`
	Title        string `gorm:"column:title" json:"title"`
	Content      string `gorm:"column:content" json:"content"`
	Desc         string `gorm:"column:desc" json:"desc"`
	Actions      string `gorm:"column:actions" json:"actions"`
	Time         int64  `gorm:"column:time" json:"time"`
	ReadStatus   int    `gorm:"column:read_status" json:"read_status"`
}

func (Notice) TableName() string {
	return "market_notice_list"
}
