package api

import (
	"market-api/db"
	"market-api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListComments(c *gin.Context) {
	query := db.DB.Model(&models.AppReply{}).Preload("User").Preload("App")

	if appIDStr := c.Query("app_id"); appIDStr != "" {
		appID, _ := strconv.Atoi(appIDStr)
		if appID > 0 {
			query = query.Where("app_id = ?", appID)
		}
	}

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	var comments []models.AppReply
	query.Order("send_time desc").Offset(offset).Limit(pageSize).Find(&comments)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  comments,
			"total": total,
		},
	})
}

func UpdateComment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Visibility int `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	if err := db.DB.Model(&models.AppReply{}).Where("id = ?", id).Update("visibility", req.Visibility).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

func DeleteComment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := db.DB.Delete(&models.AppReply{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
