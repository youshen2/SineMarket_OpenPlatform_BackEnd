package middleware

import (
	"market-api/db"
	"market-api/models"
	"market-api/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请求未携带token，无权限访问"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "请求头中auth格式有误"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "无效的Token"})
			c.Abort()
			return
		}

		var dbToken models.UserToken
		if err := db.DB.Where("token = ? AND status = ? AND expire_time > ?", tokenString, 1, time.Now().UnixMilli()).First(&dbToken).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token已失效或已在别处登录"})
			c.Abort()
			return
		}

		if dbToken.ByUserID != claims.UserID {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token信息不匹配"})
			c.Abort()
			return
		}

		var user models.User
		if err := db.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户不存在"})
			c.Abort()
			return
		}

		if user.BanTime > time.Now().UnixMilli() {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "用户已被封禁"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("token_id", dbToken.ID)
		c.Next()
	}
}

func PermissionMiddleware(requiredPermission int) gin.HandlerFunc {
	return func(c *gin.Context) {
		_user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "无法获取用户信息"})
			c.Abort()
			return
		}

		user := _user.(models.User)
		if user.UserPermission < requiredPermission {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}
