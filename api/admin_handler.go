package api

import (
	"market-api/db"
	"market-api/models"
	"market-api/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func ListBanners(c *gin.Context) {
	var banners []models.Banner
	db.DB.Order("id desc").Find(&banners)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": banners})
}

func CreateBanner(c *gin.Context) {
	actions := c.PostForm("actions")
	visibility, _ := strconv.Atoi(c.PostForm("visibility"))

	file, err := c.FormFile("banner")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "图片上传失败: " + err.Error()})
		return
	}

	bannerPath, err := utils.SaveUploadedFile(file, viper.GetString("storage.banner_path"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存图片失败: " + err.Error()})
		return
	}

	banner := models.Banner{
		BannerURL:  bannerPath,
		Actions:    actions,
		Visibility: visibility,
	}

	if err := db.DB.Create(&banner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建头图失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": banner})
}

func DeleteBanner(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := db.DB.Delete(&models.Banner{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

func ListBannedIPs(c *gin.Context) {
	var ips []models.BannedIP
	db.DB.Order("id desc").Find(&ips)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ips})
}

func CreateBannedIP(c *gin.Context) {
	var req struct {
		IP string `json:"ip" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}
	bannedIP := models.BannedIP{IP: req.IP}
	if err := db.DB.Create(&bannedIP).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功", "data": bannedIP})
}

func DeleteBannedIP(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := db.DB.Delete(&models.BannedIP{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

func ListProhibitedWords(c *gin.Context) {
	var words []models.ProhibitedWord
	db.DB.Order("id desc").Find(&words)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": words})
}

func CreateProhibitedWord(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}
	word := models.ProhibitedWord{Content: req.Content}
	if err := db.DB.Create(&word).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功", "data": word})
}

func DeleteProhibitedWord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := db.DB.Delete(&models.ProhibitedWord{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

func ListUsernameBlacklists(c *gin.Context) {
	var usernames []models.UsernameBlacklist
	db.DB.Order("id desc").Find(&usernames)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": usernames})
}

func CreateUsernameBlacklist(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}
	blacklist := models.UsernameBlacklist{Username: req.Username}
	if err := db.DB.Create(&blacklist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功", "data": blacklist})
}

func DeleteUsernameBlacklist(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := db.DB.Delete(&models.UsernameBlacklist{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
