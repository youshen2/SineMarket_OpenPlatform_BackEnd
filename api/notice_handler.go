package api

import (
	"market-api/db"
	"market-api/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetUnreadNotices(c *gin.Context) {
	currentUser := c.MustGet("user").(models.User)
	var notices []models.Notice
	db.DB.Where("by_userid = ? AND read_status = 0", currentUser.ID).
		Order("time desc").
		Limit(10).
		Find(&notices)

	var count int64
	db.DB.Model(&models.Notice{}).Where("by_userid = ? AND read_status = 0", currentUser.ID).Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  notices,
			"count": count,
		},
	})
}

func MarkAllAsRead(c *gin.Context) {
	currentUser := c.MustGet("user").(models.User)
	err := db.DB.Model(&models.Notice{}).
		Where("by_userid = ? AND read_status = 0", currentUser.ID).
		Update("read_status", 1).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "操作失败: " + err.Error()})
		return
	}

	db.DB.Model(&models.Notice{}).
		Where("by_userid = ? AND read_status = 1", currentUser.ID).
		Update("updated_at", time.Now())

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功"})
}
