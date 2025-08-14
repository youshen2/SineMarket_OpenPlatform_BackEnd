package api

import (
	"fmt"
	"market-api/db"
	"market-api/models"
	"market-api/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 403, "msg": "用户名或密码错误"})
		return
	}

	if utils.MD5(req.Password) != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 403, "msg": "用户名或密码错误"})
		return
	}

	if user.VerifyEmail != 1 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "账号未验证，请先前往客户端验证账号后再试。"})
		return
	}

	if user.BanTime > time.Now().UnixMilli() {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "您已被封禁，请等待解禁后再登录后台。"})
		return
	}

	loginToken := models.UserToken{
		LoginDevice:  "网页端后台",
		LoginVersion: 0,
		LoginIP:      c.ClientIP(),
		CreateTime:   time.Now().UnixMilli(),
		ExpireTime:   time.Now().Add(time.Duration(viper.GetInt("jwt.expire_hours")) * time.Hour).UnixMilli(),
		ByUserID:     user.ID,
		Status:       1,
	}
	if err := db.DB.Create(&loginToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建登录记录失败"})
		return
	}

	tokenString, err := utils.GenerateToken(&user, loginToken.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成Token失败"})
		return
	}

	db.DB.Model(&loginToken).Update("token", tokenString)

	db.DB.Model(&user).Updates(models.User{
		LastLoginIP:     c.ClientIP(),
		LastLoginDevice: "网页端后台",
		LastOnlineTime:  time.Now().UnixMilli(),
	})

	// Send login reminder email
	if user.BindEmail != "" && user.VerifyEmail == 1 {
		emailData := struct {
			DisplayName string
			Time        string
			DeviceName  string
			IP          string
		}{
			DisplayName: user.DisplayName,
			Time:        time.Now().Format("2006-01-02 15:04"),
			DeviceName:  "网页端后台",
			IP:          c.ClientIP(),
		}
		body, err := utils.ParseTemplate("login_reminder.html", emailData)
		if err != nil {
			fmt.Printf("Error parsing email template: %v\n", err)
		} else {
			err = utils.SendEmail(user.BindEmail, "【弦-应用商店】登录提醒", body)
			if err != nil {
				fmt.Printf("Error sending login email: %v\n", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": tokenString,
			"user":  user,
		},
	})
}

func UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的用户ID"})
		return
	}

	var reqUser models.User
	if err := c.ShouldBindJSON(&reqUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败: " + err.Error()})
		return
	}

	var targetUser models.User
	if err := db.DB.First(&targetUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	currentUser := c.MustGet("user").(models.User)

	if currentUser.ID == targetUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请通过 '个人中心' 修改自己的信息"})
		return
	}

	if currentUser.UserPermission <= targetUser.UserPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权修改该用户的信息"})
		return
	}

	updates := map[string]interface{}{
		"display_name":    reqUser.DisplayName,
		"user_describe":   reqUser.UserDescribe,
		"user_official":   reqUser.UserOfficial,
		"user_badge":      reqUser.UserBadge,
		"user_permission": reqUser.UserPermission,
		"verify_email":    reqUser.VerifyEmail,
	}

	if err := db.DB.Model(&targetUser).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
		return
	}

	var updatedUser models.User
	db.DB.First(&updatedUser, id)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
		"data": gin.H{
			"user": updatedUser,
		},
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	currentUser := c.MustGet("user").(models.User)

	if utils.MD5(req.OldPassword) != currentUser.Password {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "旧密码不正确"})
		return
	}

	if err := db.DB.Model(&currentUser).Update("password", utils.MD5(req.NewPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "修改密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "密码修改成功"})
}

func GetSelfInfo(c *gin.Context) {
	_user, _ := c.Get("user")
	user := _user.(models.User)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user})
}

func Logout(c *gin.Context) {
	tokenID, exists := c.Get("token_id")
	if exists {
		db.DB.Model(&models.UserToken{}).Where("id = ?", tokenID).Update("status", 0)
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "登出成功"})
}

func ListUsers(c *gin.Context) {
	var users []models.User
	query := db.DB.Model(&models.User{})

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("id LIKE ? OR username LIKE ? OR display_name LIKE ? OR bind_qq LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	query.Order("join_time desc").Offset(offset).Limit(pageSize).Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  users,
			"total": total,
		},
	})
}

func CreateUser(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败: " + err.Error()})
		return
	}

	var count int64
	db.DB.Model(&models.User{}).Where("username = ?", newUser.Username).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "用户名已存在"})
		return
	}

	newUser.Password = utils.MD5("123456")
	newUser.JoinTime = time.Now().UnixMilli()
	newUser.UserAvatar = viper.GetString("user.default_avatar_url")
	newUser.VerifyEmail = 1
	newUser.RegisterIP = c.ClientIP()

	if err := db.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功，默认密码为 123456", "data": newUser})
}

func DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	currentUser := c.MustGet("user").(models.User)

	if id == currentUser.ID {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无法删除当前登录的用户"})
		return
	}

	var targetUser models.User
	if err := db.DB.First(&targetUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	if currentUser.UserPermission <= targetUser.UserPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权删除该用户"})
		return
	}

	if err := db.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

type BanUserRequest struct {
	Reason string `json:"reason" binding:"required"`
	Hours  int    `json:"hours" binding:"required"`
}

func BanUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	var targetUser models.User
	if err := db.DB.First(&targetUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	currentUser := c.MustGet("user").(models.User)
	if currentUser.UserPermission <= targetUser.UserPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权封禁该用户"})
		return
	}

	banTime := time.Now().Add(time.Duration(req.Hours) * time.Hour).UnixMilli()
	updates := map[string]interface{}{
		"ban_time":           banTime,
		"user_status_reason": req.Reason,
	}

	if err := db.DB.Model(&targetUser).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "封禁失败: " + err.Error()})
		return
	}

	notice := models.Notice{
		ByUserID:     targetUser.ID,
		SenderUserID: currentUser.ID,
		Title:        "账号被封禁",
		Content:      fmt.Sprintf("您的账号因为〖%s〗被运营【%s】封禁。", req.Reason, currentUser.DisplayName),
		Desc:         "异议请联系对应运营",
		Actions:      "[]",
		Time:         time.Now().UnixMilli(),
		ReadStatus:   0,
	}
	db.DB.Create(&notice)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "封禁成功"})
}

func UnbanUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var targetUser models.User
	if err := db.DB.First(&targetUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	currentUser := c.MustGet("user").(models.User)
	if currentUser.UserPermission <= targetUser.UserPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权解封该用户"})
		return
	}

	if err := db.DB.Model(&targetUser).Update("ban_time", time.Now().UnixMilli()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "解封失败: " + err.Error()})
		return
	}

	notice := models.Notice{
		ByUserID:     targetUser.ID,
		SenderUserID: currentUser.ID,
		Title:        "封禁已解除",
		Content:      fmt.Sprintf("您的账号由运营【%s】解除封禁。", currentUser.DisplayName),
		Desc:         "欢迎回来",
		Actions:      "[]",
		Time:         time.Now().UnixMilli(),
		ReadStatus:   0,
	}
	db.DB.Create(&notice)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "解封成功"})
}

func ResetPassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var targetUser models.User
	if err := db.DB.First(&targetUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	currentUser := c.MustGet("user").(models.User)
	if currentUser.UserPermission <= targetUser.UserPermission {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权重置该用户密码"})
		return
	}

	newPassword := "123456"
	if err := db.DB.Model(&targetUser).Update("password", utils.MD5(newPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "重置密码失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "密码已重置为 123456"})
}

func ListUserTokens(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var tokens []models.UserToken
	db.DB.Where("by_userid = ?", id).Order("create_time desc").Find(&tokens)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tokens})
}

func KickUserToken(c *gin.Context) {
	tokenId, _ := strconv.Atoi(c.Param("id"))
	var token models.UserToken
	if err := db.DB.First(&token, tokenId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "登录记录不存在"})
		return
	}

	currentUser := c.MustGet("user").(models.User)
	if currentUser.UserPermission < 2 && currentUser.ID != token.ByUserID {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作"})
		return
	}

	if err := db.DB.Model(&token).Update("status", 0).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "操作失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功"})
}

func GetUserByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": user})
}

type UpdateProfileRequest struct {
	DisplayName  string `json:"display_name"`
	UserDescribe string `json:"user_describe"`
}

func UpdateSelfInfo(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数绑定失败: " + err.Error()})
		return
	}

	currentUser := c.MustGet("user").(models.User)

	updates := map[string]interface{}{
		"display_name":  req.DisplayName,
		"user_describe": req.UserDescribe,
	}

	if err := db.DB.Model(&currentUser).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
		return
	}

	var updatedUser models.User
	db.DB.First(&updatedUser, currentUser.ID)

	tokenID, _ := c.Get("token_id")
	newToken, err := utils.GenerateToken(&updatedUser, tokenID.(int))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200, "msg": "更新成功，但Token刷新失败",
			"data": gin.H{"user": updatedUser},
		})
		return
	}

	db.DB.Model(&models.UserToken{}).Where("id = ?", tokenID).Update("token", newToken)

	c.JSON(http.StatusOK, gin.H{
		"code": 200, "msg": "更新成功",
		"data": gin.H{"user": updatedUser, "token": newToken},
	})
}

func ResetAvatar(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var targetUser models.User
	if err := db.DB.First(&targetUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "用户不存在"})
		return
	}

	currentUser := c.MustGet("user").(models.User)
	if currentUser.UserPermission < 3 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权操作"})
		return
	}

	defaultAvatar := viper.GetString("user.default_avatar_url")
	if err := db.DB.Model(&targetUser).Update("user_avatar", defaultAvatar).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "重置头像失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "头像已重置为默认"})
}

func ListMyReports(c *gin.Context) {
	currentUser := c.MustGet("user").(models.User)
	query := db.DB.Model(&models.Report{}).Where("by_userid = ?", currentUser.ID)

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	var reports []models.Report
	query.Order("report_time desc").Offset(offset).Limit(pageSize).Find(&reports)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  reports,
			"total": total,
		},
	})
}

func ListMyComments(c *gin.Context) {
	currentUser := c.MustGet("user").(models.User)
	query := db.DB.Model(&models.AppReply{}).Preload("App").Where("by_userid = ?", currentUser.ID)

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
