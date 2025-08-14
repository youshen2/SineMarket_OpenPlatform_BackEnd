package models

type App struct {
	ID               int    `gorm:"primaryKey;column:id" json:"id"`
	PackageName      string `gorm:"column:package_name" json:"package_name"`
	AppName          string `gorm:"column:app_name" json:"app_name"`
	Keyword          string `gorm:"column:keyword" json:"keyword"`
	VersionCode      int    `gorm:"column:version_code" json:"version_code"`
	VersionName      string `gorm:"column:version_name" json:"version_name"`
	AppIcon          string `gorm:"column:app_icon" json:"app_icon"`
	ByUserID         int    `gorm:"column:by_userid" json:"by_userid"`
	AppTypeID        int    `gorm:"column:app_type" json:"app_type_id"`
	AppVersionTypeID int    `gorm:"column:app_version_type" json:"app_version_type_id"`
	AppABI           int    `gorm:"column:app_abi" json:"app_abi"`
	AppTags          string `gorm:"column:app_tags" json:"app_tags"`
	AppPages         string `gorm:"column:app_pages" json:"app_pages"`
	AppPreviews      string `gorm:"column:app_previews" json:"app_previews"`
	AppDescribe      string `gorm:"column:app_describe" json:"app_describe"`
	AppUpdateLog     string `gorm:"column:app_update_log" json:"app_update_log"`
	AppDeveloper     string `gorm:"column:app_developer" json:"app_developer"`
	AppSource        string `gorm:"column:app_source" json:"app_source"`
	UploadMessage    string `gorm:"column:upload_message" json:"upload_message"`
	AuditStatus      int    `gorm:"column:audit_status" json:"audit_status"`
	AuditReason      string `gorm:"column:audit_reason" json:"audit_reason"`
	AuditUser        int    `gorm:"column:audit_user" json:"audit_user"`
	AppSdkMin        int    `gorm:"column:app_sdk_min" json:"app_sdk_min"`
	AppSdkTarget     int    `gorm:"column:app_sdk_target" json:"app_sdk_target"`
	AppIsWearOS      int    `gorm:"column:app_is_wearos" json:"app_is_wearos"`
	DownloadSize     string `gorm:"column:download_size" json:"download_size"`
	UploadTime       int64  `gorm:"column:upload_time" json:"upload_time"`
	UpdateTime       int64  `gorm:"column:update_time" json:"update_time"`
	LocalApkPath     string `gorm:"column:local_apk_path" json:"local_apk_path"`
	LocalIconPath    string `gorm:"column:local_icon_path" json:"local_icon_path"`
	AppWeight        int    `gorm:"column:app_weight" json:"app_weight"`
	HasAppUpdateNotice int  `gorm:"column:has_app_update_notice" json:"has_app_update_notice"`
	Uploader         User   `gorm:"foreignKey:ByUserID" json:"uploader"`
}

func (App) TableName() string {
	return "market_app_list"
}

type AppTag struct {
	ID               int    `gorm:"primaryKey;column:id" json:"id"`
	Name             string `gorm:"column:name" json:"name"`
	Icon             string `gorm:"column:icon" json:"icon"`
	UploadPermission int    `gorm:"column:upload_permission" json:"upload_permission"`
}

func (AppTag) TableName() string {
	return "market_app_tags_list"
}

type AppType struct {
	ID   int    `gorm:"primaryKey;column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
}

func (AppType) TableName() string {
	return "market_app_type_list"
}

type AppVersionType struct {
	ID   int    `gorm:"primaryKey;column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
}

func (AppVersionType) TableName() string {
	return "market_app_version_type_list"
}

type AppDownload struct {
	ID          int    `gorm:"primaryKey;column:id" json:"id"`
	AppID       int    `gorm:"column:app_id" json:"app_id"`
	Name        string `gorm:"column:name" json:"name"`
	URL         string `gorm:"column:url" json:"url"`
	IsExtra     int    `gorm:"column:is_extra" json:"is_extra"`
	AuditStatus int    `gorm:"column:audit_status" json:"audit_status"`
	App         App    `gorm:"foreignKey:AppID" json:"app"`
}

func (AppDownload) TableName() string {
	return "market_app_download_list"
}

type Splash struct {
	ID   int   `gorm:"primaryKey;column:id"`
	Time int64 `gorm:"column:time"`
}

func (Splash) TableName() string {
	return "market_splash_list"
}

type UserDownloadCount struct {
	ID    int   `gorm:"primaryKey;column:id"`
	Count int   `gorm:"column:count"`
	Time  int64 `gorm:"column:time"`
}

func (UserDownloadCount) TableName() string {
	return "market_user_download_count_list"
}

type AppReply struct {
	ID       int `gorm:"primaryKey;column:id"`
	ByUserID int `gorm:"column:by_userid"`
}

func (AppReply) TableName() string {
	return "market_app_reply_list"
}
