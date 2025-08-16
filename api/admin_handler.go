package api

import (
	"fmt"
	"market-api/db"
	"market-api/models"
	"market-api/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var allowedImageExtensions = []string{"jpg", "jpeg", "png", "webp"}

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

	if !utils.ValidateFileExtension(file.Filename, allowedImageExtensions) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "不支持的图片格式，请上传 jpg, jpeg, png, webp 格式的图片"})
		return
	}

	relativeBannerPath, err := utils.SaveUploadedFile(file, viper.GetString("storage.banner_path"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存图片失败: " + err.Error()})
		return
	}

	baseURL := viper.GetString("server.base_url")
	fullBannerURL := fmt.Sprintf("%s/%s", baseURL, relativeBannerPath)

	banner := models.Banner{
		BannerURL:  fullBannerURL,
		Actions:    actions,
		Visibility: visibility,
	}

	if err := db.DB.Create(&banner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建头图失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": banner})
}

func UpdateBanner(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Actions    string `json:"actions"`
		Visibility int    `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	updates := map[string]interface{}{
		"actions":    req.Actions,
		"visibility": req.Visibility,
	}

	if err := db.DB.Model(&models.Banner{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

func DeleteBanner(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var banner models.Banner
	if err := db.DB.First(&banner, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "头图不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询头图失败: " + err.Error()})
		return
	}

	baseURL := viper.GetString("server.base_url")
	if banner.BannerURL != "" && strings.HasPrefix(banner.BannerURL, baseURL) {
		relativePath := strings.TrimPrefix(banner.BannerURL, baseURL+"/")
		if err := utils.DeleteFile(relativePath); err != nil {
			fmt.Printf("Warning: failed to delete banner image file %s: %v\n", relativePath, err)
		}
	}

	if err := db.DB.Delete(&models.Banner{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除数据库记录失败: " + err.Error()})
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

func ListReports(c *gin.Context) {
	query := db.DB.Model(&models.Report{}).Preload("Reporter")

	if statusStr := c.Query("status"); statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		query = query.Where("report_status = ?", status)
	}

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("report_title LIKE ? OR report_ip LIKE ? OR by_userid LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	sortField := c.DefaultQuery("sortField", "report_time")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	allowedSortFields := map[string]string{
		"report_time": "report_time",
	}
	dbSortField, ok := allowedSortFields[sortField]
	if !ok {
		dbSortField = "report_time"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	var reports []models.Report
	query.Order(fmt.Sprintf("%s %s", dbSortField, sortOrder)).Offset(offset).Limit(pageSize).Find(&reports)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  reports,
			"total": total,
		},
	})
}

func GetReportDetails(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var report models.Report
	if err := db.DB.Preload("Reporter").First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "举报不存在"})
		return
	}

	var reportedItem interface{}
	var parentComment interface{}

	if report.ReportType == 1 {
		var app models.App
		db.DB.Preload("Uploader").First(&app, report.ReportID)
		reportedItem = app
	} else if report.ReportType == 2 {
		var comment models.AppReply
		db.DB.Preload("User").Preload("App").First(&comment, report.ReportID)
		reportedItem = comment
		if comment.FatherReplyID != 0 {
			var pComment models.AppReply
			if err := db.DB.Preload("User").First(&pComment, comment.FatherReplyID).Error; err == nil {
				parentComment = pComment
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"report":        report,
			"reportedItem":  reportedItem,
			"parentComment": parentComment,
		},
	})
}

type AuditReportRequest struct {
	Reply      string `json:"reply" binding:"required"`
	TakeAction bool   `json:"take_action"`
}

func AuditReport(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req AuditReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	var report models.Report
	if err := db.DB.First(&report, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "举报不存在"})
		return
	}

	if report.ReportStatus == 1 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "该举报已被处理"})
		return
	}

	currentUser := c.MustGet("user").(models.User)
	tx := db.DB.Begin()

	updateData := map[string]interface{}{
		"report_status": 1,
		"report_reply":  req.Reply,
		"reply_time":    time.Now().UnixMilli(),
	}
	if err := tx.Model(&models.Report{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新举报状态失败"})
		return
	}

	noticeToReporter := models.Notice{
		ByUserID:     report.ByUserID,
		SenderUserID: -1,
		Title:        "举报已处理",
		Content:      fmt.Sprintf("您关于「%s」的举报已由运营【%s】处理。", report.ReportTitle, currentUser.DisplayName),
		Desc:         "处理回复：" + req.Reply,
		Time:         time.Now().UnixMilli(),
		Actions:      "[]",
	}
	if err := tx.Create(&noticeToReporter).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发送通知给举报者失败"})
		return
	}

	if req.TakeAction {
		if report.ReportType == 1 {
			var app models.App
			if err := tx.First(&app, report.ReportID).Error; err == nil {
				appUpdates := map[string]interface{}{
					"audit_status": 2,
					"audit_reason": "被用户举报，验证后处理下架",
					"audit_user":   currentUser.ID,
				}
				if err := tx.Model(&app).Updates(appUpdates).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "下架应用失败"})
					return
				}
				noticeToUploader := models.Notice{
					ByUserID:     app.ByUserID,
					SenderUserID: -1,
					Title:        "应用审核不通过",
					Content:      fmt.Sprintf("您上传的应用「%s」经用户举报，由【%s】审核后被处理下架。", app.AppName, currentUser.DisplayName),
					Desc:         "异议请联系对应运营",
					Time:         time.Now().UnixMilli(),
					Actions:      "[]",
				}
				if err := tx.Create(&noticeToUploader).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "发送通知给上传者失败"})
					return
				}
			}
		} else if report.ReportType == 2 {
			if err := tx.Model(&models.AppReply{}).Where("id = ?", report.ReportID).Update("visibility", 0).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "隐藏评论失败"})
				return
			}
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "举报处理成功"})
}

func GetSetting(c *gin.Context) {
	key := c.Param("key")
	var setting models.Setting
	if err := db.DB.Where("setting_key = ?", key).First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{"code": 200, "data": models.Setting{Key: key, Value: ""}})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "获取设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": setting})
}

func UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	setting := models.Setting{Key: key, Value: req.Value}
	if err := db.DB.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新设置失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "设置更新成功"})
}

func ListAppPages(c *gin.Context) {
	query := db.DB.Model(&models.AppPage{})

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	var pages []models.AppPage
	query.Order("id desc").Offset(offset).Limit(pageSize).Find(&pages)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  pages,
			"total": total,
		},
	})
}

func GetAppPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var page models.AppPage
	if err := db.DB.First(&page, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "专题不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": page})
}

func CreateAppPage(c *gin.Context) {
	var page models.AppPage
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}
	page.Time = time.Now().UnixMilli()
	if err := db.DB.Create(&page).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建专题失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": page})
}

func UpdateAppPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var page models.AppPage
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}
	updates := map[string]interface{}{
		"title":        page.Title,
		"content":      page.Content,
		"img_list":     page.ImgList,
		"has_app_list": page.HasAppList,
		"show_in_list": page.ShowInList,
	}
	if err := db.DB.Model(&models.AppPage{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新专题失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

func DeleteAppPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	tx := db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "开启事务失败: " + tx.Error.Error()})
		return
	}

	pageIDStr := strconv.Itoa(id)
	searchStr := fmt.Sprintf("%s,", pageIDStr)

	if err := tx.Model(&models.App{}).
		Where("app_pages LIKE ?", "%,"+searchStr+"%").
		Update("app_pages", gorm.Expr("REPLACE(app_pages, ?, ?)", searchStr, "")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "从应用中移除专题关联失败: " + err.Error()})
		return
	}

	if err := tx.Delete(&models.AppPage{}, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除专题失败: " + err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

func SyncAppsToPage(c *gin.Context) {
	pageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的专题ID"})
		return
	}

	var req struct {
		AppIDs []int `json:"app_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "开启事务失败: " + tx.Error.Error()})
		return
	}

	pageIDStr := strconv.Itoa(pageID)
	searchStr := fmt.Sprintf("%s,", pageIDStr)

	if err := tx.Model(&models.App{}).
		Where("app_pages LIKE ?", "%,"+searchStr+"%").
		Update("app_pages", gorm.Expr("REPLACE(app_pages, ?, ?)", searchStr, "")).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "移除旧的应用关联失败: " + err.Error()})
		return
	}

	if len(req.AppIDs) > 0 {
		appendStr := fmt.Sprintf("%s,", pageIDStr)
		if err := tx.Model(&models.App{}).
			Where("id IN ?", req.AppIDs).
			Update("app_pages", gorm.Expr("CONCAT(app_pages, ?)", appendStr)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加新的应用关联失败: " + err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "应用关联同步成功"})
}
