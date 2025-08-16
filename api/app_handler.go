package api

import (
	"encoding/json"
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
)

var allowedApkExtensions = []string{"apk"}

func GetAppTags(c *gin.Context) {
	var tags []models.AppTag
	db.DB.Find(&tags)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": tags})
}

func GetAppTypes(c *gin.Context) {
	var types []models.AppType
	db.DB.Find(&types)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": types})
}

func GetAppVersionTypes(c *gin.Context) {
	var versionTypes []models.AppVersionType
	db.DB.Find(&versionTypes)
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": versionTypes})
}

func ListApps(c *gin.Context) {
	currentUser := c.MustGet("user").(models.User)
	query := db.DB.Model(&models.App{}).Preload("Uploader")

	scope := c.Query("scope")

	if scope == "all" && currentUser.UserPermission >= 1 {
		if statusStr := c.Query("audit_status"); statusStr != "" {
			status, err := strconv.Atoi(statusStr)
			if err == nil {
				query = query.Where("audit_status = ?", status)
			}
		}
	} else {
		query = query.Where("by_userid = ?", currentUser.ID)
	}

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("app_name LIKE ? OR package_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	sortField := c.DefaultQuery("sortField", "update_time")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	allowedSortFields := map[string]string{
		"update_time": "update_time",
	}
	dbSortField, ok := allowedSortFields[sortField]
	if !ok {
		dbSortField = "update_time"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	var apps []models.App
	query.Order(fmt.Sprintf("%s %s", dbSortField, sortOrder)).Offset(offset).Limit(pageSize).Find(&apps)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  apps,
			"total": total,
		},
	})
}

func GetApp(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var app models.App
	if err := db.DB.Preload("Uploader").First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "应用不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": app})
}

func PreUploadApp(c *gin.Context) {
	currentUser := c.MustGet("user").(models.User)
	baseURL := viper.GetString("server.base_url")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "解析表单失败: " + err.Error()})
		return
	}

	iconFile, ok := form.File["icon"]
	if !ok || len(iconFile) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "必须上传应用图标"})
		return
	}
	if !utils.ValidateFileExtension(iconFile[0].Filename, allowedImageExtensions) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "不支持的图标格式，请上传 jpg, jpeg, png, webp 格式的图片"})
		return
	}
	relativeIconPath, err := utils.SaveUploadedFile(iconFile[0], viper.GetString("storage.icon_path"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存图标失败: " + err.Error()})
		return
	}
	fullIconURL := fmt.Sprintf("%s/%s", baseURL, relativeIconPath)

	var screenshotURLs []string
	screenshotFiles := form.File["screenshots"]
	for _, file := range screenshotFiles {
		if !utils.ValidateFileExtension(file.Filename, allowedImageExtensions) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "包含不支持的截图格式，请上传 jpg, jpeg, png, webp 格式的图片"})
			return
		}
		relativePath, err := utils.SaveUploadedFile(file, viper.GetString("storage.preview_path"))
		if err != nil {
			fmt.Printf("Warning: could not save screenshot %s: %v\n", file.Filename, err)
			continue
		}
		fullScreenshotURL := fmt.Sprintf("%s/%s", baseURL, relativePath)
		screenshotURLs = append(screenshotURLs, fullScreenshotURL)
	}

	if screenshotURLs == nil {
		screenshotURLs = make([]string, 0)
	}
	screenshotsJSON, _ := json.Marshal(screenshotURLs)

	versionCode, _ := strconv.Atoi(c.PostForm("version_code"))
	appTypeID, _ := strconv.Atoi(c.PostForm("app_type_id"))
	appVersionTypeID, _ := strconv.Atoi(c.PostForm("app_version_type_id"))
	appABI, _ := strconv.Atoi(c.PostForm("app_abi"))
	appSdkMin, _ := strconv.Atoi(c.PostForm("app_sdk_min"))
	appSdkTarget, _ := strconv.Atoi(c.PostForm("app_sdk_target"))
	downloadSize, _ := strconv.ParseInt(c.PostForm("download_size"), 10, 64)
	appIsWearOS, _ := strconv.Atoi(c.PostForm("app_is_wearos"))

	app := models.App{
		PackageName:      c.PostForm("package_name"),
		AppName:          c.PostForm("app_name"),
		Keyword:          c.PostForm("keyword"),
		VersionCode:      versionCode,
		VersionName:      c.PostForm("version_name"),
		AppIcon:          fullIconURL,
		ByUserID:         currentUser.ID,
		AppTypeID:        appTypeID,
		AppVersionTypeID: appVersionTypeID,
		AppABI:           appABI,
		AppTags:          c.PostForm("app_tags"),
		AppPages:         ",",
		AppPreviews:      string(screenshotsJSON),
		AppDescribe:      c.PostForm("app_describe"),
		AppUpdateLog:     c.PostForm("app_update_log"),
		AppDeveloper:     c.PostForm("app_developer"),
		AppSource:        c.PostForm("app_source"),
		UploadMessage:    c.PostForm("upload_message"),
		AuditStatus:      0,
		AuditReason:      "应用还在审核中",
		AppSdkMin:        appSdkMin,
		AppSdkTarget:     appSdkTarget,
		AppIsWearOS:      appIsWearOS,
		DownloadSize:     utils.FormatSizeUnits(downloadSize),
		UploadTime:       time.Now().UnixMilli(),
		UpdateTime:       time.Now().UnixMilli(),
		LocalIconPath:    relativeIconPath,
	}

	tx := db.DB.Begin()

	if err := tx.Create(&app).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建应用记录失败: " + err.Error()})
		return
	}

	remoteApkPath := fmt.Sprintf("apks/%d.apk", app.ID)

	defaultDownload := models.AppDownload{
		AppID:       app.ID,
		Name:        "新版路线",
		URL:         remoteApkPath,
		IsExtra:     1,
		AuditStatus: 1,
	}
	if err := tx.Create(&defaultDownload).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建默认下载路线失败: " + err.Error()})
		return
	}

	uploadToken, err := utils.GetUploadToken(remoteApkPath)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取文件上传凭证失败: " + err.Error()})
		return
	}

	if err := tx.Model(&app).Update("local_apk_path", remoteApkPath).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新应用APK路径失败: " + err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "元数据上传成功，请继续上传APK文件",
		"data": gin.H{
			"upload_token": uploadToken,
			"app_id":       app.ID,
		},
	})
}

func UpdateApp(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	currentUser := c.MustGet("user").(models.User)
	baseURL := viper.GetString("server.base_url")

	var app models.App
	if err := db.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "应用不存在"})
		return
	}

	if app.ByUserID != currentUser.ID && currentUser.UserPermission < 3 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权修改此应用"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的请求数据: " + err.Error()})
		return
	}

	updates := make(map[string]interface{})
	for key, values := range form.Value {
		if len(values) > 0 {
			updates[key] = values[0]
		}
	}

	iconFileHeader, ok := form.File["icon"]
	if ok && len(iconFileHeader) > 0 {
		if !utils.ValidateFileExtension(iconFileHeader[0].Filename, allowedImageExtensions) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "不支持的图标格式，请上传 jpg, jpeg, png, webp 格式的图片"})
			return
		}
		if app.LocalIconPath != "" {
			if err := utils.DeleteFile(app.LocalIconPath); err != nil {
				fmt.Printf("Warning: failed to delete old icon %s: %v\n", app.LocalIconPath, err)
			}
		}
		relativeNewIconPath, err := utils.SaveUploadedFile(iconFileHeader[0], viper.GetString("storage.icon_path"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存新图标失败: " + err.Error()})
			return
		}
		updates["app_icon"] = fmt.Sprintf("%s/%s", baseURL, relativeNewIconPath)
		updates["local_icon_path"] = relativeNewIconPath
	}

	var finalScreenshotURLs []string
	existingScreenshotsJSON := c.PostForm("existing_screenshots")
	var keptScreenshotURLs []string
	if err := json.Unmarshal([]byte(existingScreenshotsJSON), &keptScreenshotURLs); err == nil {
		finalScreenshotURLs = append(finalScreenshotURLs, keptScreenshotURLs...)
	}

	var currentScreenshotURLs []string
	if app.AppPreviews != "" {
		json.Unmarshal([]byte(app.AppPreviews), &currentScreenshotURLs)
	}

	keptScreenshotsMap := make(map[string]bool)
	for _, url := range keptScreenshotURLs {
		keptScreenshotsMap[url] = true
	}

	for _, url := range currentScreenshotURLs {
		if !keptScreenshotsMap[url] {
			relativePath := strings.TrimPrefix(url, baseURL+"/")
			if err := utils.DeleteFile(relativePath); err != nil {
				fmt.Printf("Warning: failed to delete screenshot %s: %v\n", relativePath, err)
			}
		}
	}

	newScreenshotFiles := form.File["screenshots"]
	for _, file := range newScreenshotFiles {
		if !utils.ValidateFileExtension(file.Filename, allowedImageExtensions) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "包含不支持的截图格式，请上传 jpg, jpeg, png, webp 格式的图片"})
			return
		}
		relativePath, err := utils.SaveUploadedFile(file, viper.GetString("storage.preview_path"))
		if err != nil {
			fmt.Printf("Warning: could not save new screenshot %s: %v\n", file.Filename, err)
			continue
		}
		fullNewURL := fmt.Sprintf("%s/%s", baseURL, relativePath)
		finalScreenshotURLs = append(finalScreenshotURLs, fullNewURL)
	}

	if finalScreenshotURLs == nil {
		finalScreenshotURLs = make([]string, 0)
	}
	newScreenshotsJSON, _ := json.Marshal(finalScreenshotURLs)
	updates["app_previews"] = string(newScreenshotsJSON)

	if val, ok := updates["version_code"]; ok {
		updates["version_code"], _ = strconv.Atoi(val.(string))
	}
	if val, ok := updates["app_type_id"]; ok {
		updates["app_type"], _ = strconv.Atoi(val.(string))
		delete(updates, "app_type_id")
	}
	if val, ok := updates["app_version_type_id"]; ok {
		updates["app_version_type"], _ = strconv.Atoi(val.(string))
		delete(updates, "app_version_type_id")
	}
	if val, ok := updates["app_tags"]; ok {
		updates["app_tags"] = fmt.Sprintf(",%s,", val.(string))
	}
	if val, ok := updates["app_sdk_min"]; ok {
		updates["app_sdk_min"], _ = strconv.Atoi(val.(string))
	}
	if val, ok := updates["app_sdk_target"]; ok {
		updates["app_sdk_target"], _ = strconv.Atoi(val.(string))
	}
	if val, ok := updates["app_abi"]; ok {
		updates["app_abi"], _ = strconv.Atoi(val.(string))
	}
	if val, ok := updates["app_is_wearos"]; ok {
		updates["app_is_wearos"], _ = strconv.Atoi(val.(string))
	}

	updates["audit_status"] = 0
	updates["audit_reason"] = "资料已更新，等待重新审核"
	updates["update_time"] = time.Now().UnixMilli()

	delete(updates, "package_name")
	delete(updates, "existing_screenshots")
	delete(updates, "uploader")

	if err := db.DB.Model(&app).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新应用失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "应用更新成功，已提交审核"})
}

func DeleteApp(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	currentUser := c.MustGet("user").(models.User)

	var app models.App
	if err := db.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "应用不存在"})
		return
	}

	if app.ByUserID != currentUser.ID && currentUser.UserPermission < 3 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权删除此应用"})
		return
	}

	if app.LocalIconPath != "" {
		if err := utils.DeleteFile(app.LocalIconPath); err != nil {
			fmt.Printf("Warning: failed to delete icon file %s for app ID %d: %v\n", app.LocalIconPath, app.ID, err)
		}
	}

	if app.AppPreviews != "" {
		var screenshotURLs []string
		if err := json.Unmarshal([]byte(app.AppPreviews), &screenshotURLs); err == nil {
			baseURL := viper.GetString("server.base_url")
			for _, url := range screenshotURLs {
				if strings.HasPrefix(url, baseURL) {
					relativePath := strings.TrimPrefix(url, baseURL+"/")
					if err := utils.DeleteFile(relativePath); err != nil {
						fmt.Printf("Warning: failed to delete screenshot file %s for app ID %d: %v\n", relativePath, app.ID, err)
					}
				}
			}
		}
	}

	tx := db.DB.Begin()
	if err := tx.Where("app_id = ?", id).Delete(&models.AppDownload{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除应用下载链接失败"})
		return
	}
	if err := tx.Delete(&app).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除应用失败"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "应用删除成功"})
}

func ListAppDownloads(c *gin.Context) {
	appID, _ := strconv.Atoi(c.Param("id"))
	var downloads []models.AppDownload
	if err := db.DB.Where("app_id = ?", appID).Order("id asc").Find(&downloads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询下载路线失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": downloads})
}

type AddDownloadRequest struct {
	Name string `json:"name" binding:"required"`
	URL  string `json:"url" binding:"required"`
}

func AddAppDownload(c *gin.Context) {
	appID, _ := strconv.Atoi(c.Param("id"))
	currentUser := c.MustGet("user").(models.User)
	var req AddDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	var app models.App
	if err := db.DB.First(&app, appID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "应用不存在"})
		return
	}

	if app.ByUserID != currentUser.ID && currentUser.UserPermission < 1 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权为该应用添加路线"})
		return
	}

	auditStatus := 0
	if currentUser.UserPermission >= 1 {
		auditStatus = 1
	}

	download := models.AppDownload{
		AppID:       appID,
		Name:        req.Name,
		URL:         req.URL,
		IsExtra:     -1,
		AuditStatus: auditStatus,
	}

	if err := db.DB.Create(&download).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加下载路线失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功", "data": download})
}

func DeleteAppDownload(c *gin.Context) {
	downloadID, _ := strconv.Atoi(c.Param("download_id"))
	currentUser := c.MustGet("user").(models.User)

	var download models.AppDownload
	if err := db.DB.Preload("App").First(&download, downloadID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "下载路线不存在"})
		return
	}

	if download.App.ByUserID != currentUser.ID && currentUser.UserPermission < 3 {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权删除该路线"})
		return
	}

	if err := db.DB.Delete(&models.AppDownload{}, downloadID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "删除下载路线失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

type AuditRequest struct {
	Success bool   `json:"success"`
	Reason  string `json:"reason"`
}

func AuditApp(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req AuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	var app models.App
	if err := db.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "应用不存在"})
		return
	}

	currentUser := c.MustGet("user").(models.User)
	newStatus := 2
	if req.Success {
		newStatus = 1
	}

	if !req.Success && req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "驳回应用必须填写原因"})
		return
	}

	updates := map[string]interface{}{
		"audit_status": newStatus,
		"audit_reason": req.Reason,
		"audit_user":   currentUser.ID,
	}

	if err := db.DB.Model(&app).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "审核操作失败: " + err.Error()})
		return
	}

	var title, content string
	if newStatus == 1 {
		title = "应用审核通过"
		content = fmt.Sprintf("您上传的「%s」已由审核员【%s】审核通过。", app.AppName, currentUser.DisplayName)
	} else {
		title = "应用审核不通过"
		content = fmt.Sprintf("您上传的「%s」被审核员【%s】驳回，原因：%s", app.AppName, currentUser.DisplayName, req.Reason)
	}

	notice := models.Notice{
		ByUserID:     app.ByUserID,
		SenderUserID: -1,
		Title:        title,
		Content:      content,
		Time:         time.Now().UnixMilli(),
		Actions:      "[]",
	}
	db.DB.Create(&notice)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "审核操作成功"})
}

type DownloadLinkResponse struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func GetAppDownloadTestURL(c *gin.Context) {
	appID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的应用ID"})
		return
	}

	var downloads []models.AppDownload
	if err := db.DB.Where("app_id = ? AND audit_status = 1", appID).Order("id asc").Find(&downloads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查询下载链接失败: " + err.Error()})
		return
	}

	fileServerApiURL := viper.GetString("file_server.api_url")
	if fileServerApiURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "文件服务器地址未配置"})
		return
	}

	var processedDownloads []DownloadLinkResponse

	for _, download := range downloads {
		var finalURL string

		if download.IsExtra == 1 {
			apkPath := fmt.Sprintf("apks/%d.apk", download.AppID)

			token, err := utils.GetDownloadToken(apkPath)
			if err != nil {
				fmt.Printf("Error getting download token for path %s: %v\n", apkPath, err)
				continue
			}

			finalURL = fmt.Sprintf("%s/download?token=%s", fileServerApiURL, token)
		} else {
			finalURL = download.URL
		}

		processedDownloads = append(processedDownloads, DownloadLinkResponse{
			Name: download.Name,
			URL:  finalURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": processedDownloads})
}

func ListDownloadsToAudit(c *gin.Context) {
	var downloads []models.AppDownload
	query := db.DB.Model(&models.AppDownload{}).Preload("App").Where("audit_status = ?", 0)

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	offset := (page - 1) * pageSize

	query.Order("id desc").Offset(offset).Limit(pageSize).Find(&downloads)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  downloads,
			"total": total,
		},
	})
}

func AuditAppDownload(c *gin.Context) {
	downloadID, _ := strconv.Atoi(c.Param("download_id"))
	var req AuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误: " + err.Error()})
		return
	}

	newStatus := 2
	if req.Success {
		newStatus = 1
	}

	if err := db.DB.Model(&models.AppDownload{}).Where("id = ?", downloadID).Update("audit_status", newStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "审核操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "审核成功"})
}

func ListAllSimpleApps(c *gin.Context) {
	query := db.DB.Model(&models.App{})

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("app_name LIKE ? OR id LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	var apps []struct {
		ID       int    `json:"id"`
		AppName  string `json:"app_name"`
		AppIcon  string `json:"app_icon"`
		AppPages string `json:"app_pages"`
	}

	if pageSize > 500 {
		query.Select("id, app_name, app_icon, app_pages").Order("id desc").Find(&apps)
	} else {
		offset := (page - 1) * pageSize
		query.Select("id, app_name, app_icon, app_pages").Order("id desc").Offset(offset).Limit(pageSize).Find(&apps)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"list":  apps,
			"total": total,
		},
	})
}
