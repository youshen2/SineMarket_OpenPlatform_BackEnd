package api

import (
	"encoding/json"
	"market-api/db"
	"market-api/models"
	"market-api/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type SendNoticeRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	UserIDs []int  `json:"user_ids" binding:"required"`
}

func SendNotice(c *gin.Context) {
	var req SendNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	sender := c.MustGet("user").(models.User)
	targetUserIDs := req.UserIDs
	desc := "运营发送的通知"

	if len(req.UserIDs) == 1 && req.UserIDs[0] == -1 {
		var allUsers []models.User
		db.DB.Select("id").Find(&allUsers)
		targetUserIDs = []int{}
		for _, u := range allUsers {
			targetUserIDs = append(targetUserIDs, u.ID)
		}
		desc = "全站通知"
	}

	if len(targetUserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "没有指定任何用户"})
		return
	}

	var notices []models.Notice
	for _, userID := range targetUserIDs {
		notice := models.Notice{
			ByUserID:     userID,
			SenderUserID: sender.ID,
			Title:        req.Title,
			Content:      req.Content,
			Desc:         desc,
			Time:         time.Now().UnixMilli(),
			ReadStatus:   0,
			Actions:      "[]",
		}
		notices = append(notices, notice)
	}

	if err := db.DB.Create(&notices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发送通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "通知发送成功"})
}

func SendPopup(c *gin.Context) {
	sender := c.MustGet("user").(models.User)

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图片上传失败: " + err.Error()})
		return
	}

	imagePath, err := utils.SaveUploadedFile(file, viper.GetString("storage.popup_path"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存图片失败: " + err.Error()})
		return
	}

	actions := c.PostForm("actions")
	surplusCount, _ := strconv.Atoi(c.PostForm("surplus_count"))
	userIDsStr := c.PostForm("user_ids")

	if userIDsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "用户ID列表不能为空"})
		return
	}

	var targetUserIDs []int
	if userIDsStr == "-1" {
		var allUsers []models.User
		db.DB.Select("id").Find(&allUsers)
		for _, u := range allUsers {
			targetUserIDs = append(targetUserIDs, u.ID)
		}
	} else {
		idStrs := strings.Split(userIDsStr, ",")
		for _, idStr := range idStrs {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err == nil {
				targetUserIDs = append(targetUserIDs, id)
			}
		}
	}

	if len(targetUserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "解析用户ID列表后为空"})
		return
	}

	var popups []models.Popup
	for _, userID := range targetUserIDs {
		popup := models.Popup{
			ByUserID:     userID,
			SenderUserID: sender.ID,
			ImgURL:       imagePath,
			Actions:      actions,
			SurplusCount: surplusCount,
		}
		popups = append(popups, popup)
	}

	if err := db.DB.Create(&popups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发送弹窗失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "弹窗发送成功"})
}

type SendActionsRequest struct {
	UserIDs      []int           `json:"user_ids" binding:"required"`
	Actions      json.RawMessage `json:"actions" binding:"required"`
	SurplusCount int             `json:"surplus_count" binding:"required"`
}

func SendActions(c *gin.Context) {
	var req SendActionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	targetUserIDs := req.UserIDs
	if len(req.UserIDs) == 1 && req.UserIDs[0] == -1 {
		var allUsers []models.User
		db.DB.Select("id").Find(&allUsers)
		targetUserIDs = []int{}
		for _, u := range allUsers {
			targetUserIDs = append(targetUserIDs, u.ID)
		}
	}

	if len(targetUserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "没有指定任何用户"})
		return
	}

	var userActions []models.UserAction
	for _, userID := range targetUserIDs {
		action := models.UserAction{
			ByUserID:     userID,
			Actions:      string(req.Actions),
			SurplusCount: req.SurplusCount,
		}
		userActions = append(userActions, action)
	}

	if err := db.DB.Create(&userActions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发送云控失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "云控发送成功"})
}
