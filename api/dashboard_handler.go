package api

import (
	"market-api/db"
	"market-api/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardStats struct {
	AppTotal        int64 `json:"app_total"`
	AppApproved     int64 `json:"app_approved"`
	AppRejected     int64 `json:"app_rejected"`
	AppPending      int64 `json:"app_pending"`
	SplashToday     int64 `json:"splash_today"`
	DownloadsToday  int64 `json:"downloads_today"`
	RegistersToday  int64 `json:"registers_today"`
	MyUploads       int64 `json:"my_uploads"`
	MyReplies       int64 `json:"my_replies"`
}

func GetDashboardStats(c *gin.Context) {
	currentUser := c.MustGet("user").(models.User)
	stats := DashboardStats{}

	db.DB.Model(&models.App{}).Count(&stats.AppTotal)
	db.DB.Model(&models.App{}).Where("audit_status = ?", 1).Count(&stats.AppApproved)
	db.DB.Model(&models.App{}).Where("audit_status = ?", 2).Count(&stats.AppRejected)
	db.DB.Model(&models.App{}).Where("audit_status = ?", 0).Count(&stats.AppPending)

	now := time.Now()
	year, month, day := now.Date()
	beginToday := time.Date(year, month, day, 0, 0, 0, 0, now.Location()).Unix()
    beginTodayMilli := beginToday * 1000

	db.DB.Model(&models.Splash{}).Where("time = ?", beginToday).Count(&stats.SplashToday)
	
	var totalDownloads int64
	db.DB.Model(&models.UserDownloadCount{}).Where("time >= ?", beginToday).Select("SUM(count)").Row().Scan(&totalDownloads)
	stats.DownloadsToday = totalDownloads

	db.DB.Model(&models.User{}).Where("join_time >= ?", beginTodayMilli).Count(&stats.RegistersToday)

	db.DB.Model(&models.App{}).Where("by_userid = ?", currentUser.ID).Count(&stats.MyUploads)
	db.DB.Model(&models.AppReply{}).Where("by_userid = ?", currentUser.ID).Count(&stats.MyReplies)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}
