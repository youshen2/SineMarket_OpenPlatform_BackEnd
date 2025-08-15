package main

import (
	"fmt"
	"log"
	"market-api/api"
	"market-api/db"
	"market-api/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	db.Init()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour
	r.Use(cors.New(config))

	staticPath := viper.GetString("storage.base_path")
	if staticPath != "" {
		r.Static("/"+staticPath, "./"+staticPath)
	}

	setupRoutes(r)

	port := viper.GetInt("server.port")
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func setupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", api.Login)
			auth.POST("/logout", api.Logout)
		}

		authed := v1.Group("/")
		authed.Use(middleware.AuthMiddleware())
		{
			dashboardGroup := authed.Group("/dashboard")
			{
				dashboardGroup.GET("/stats", api.GetDashboardStats)
				dashboardGroup.GET("/history/splash", api.GetSplashHistory)
				dashboardGroup.GET("/history/downloads", api.GetDownloadHistory)
				dashboardGroup.GET("/history/registers", api.GetRegisterHistory)
			}

			meGroup := authed.Group("/me")
			{
				meGroup.GET("", api.GetSelfInfo)
				meGroup.PUT("", api.UpdateSelfInfo)
				meGroup.PUT("/password", api.ChangePassword)
				meGroup.GET("/reports", api.ListMyReports)
				meGroup.GET("/comments", api.ListMyComments)
			}

			userGroup := authed.Group("/users")
			{
				userGroup.GET("", middleware.PermissionMiddleware(2), api.ListUsers)
				userGroup.POST("", middleware.PermissionMiddleware(3), api.CreateUser)
				userGroup.GET("/:id", middleware.PermissionMiddleware(2), api.GetUserByID)
				userGroup.PUT("/:id", middleware.PermissionMiddleware(3), api.UpdateUser)
				userGroup.DELETE("/:id", middleware.PermissionMiddleware(3), api.DeleteUser)
				userGroup.GET("/:id/tokens", middleware.PermissionMiddleware(2), api.ListUserTokens)
				userGroup.POST("/:id/ban", middleware.PermissionMiddleware(1), api.BanUser)
				userGroup.POST("/:id/unban", middleware.PermissionMiddleware(1), api.UnbanUser)
				userGroup.POST("/:id/reset-password", middleware.PermissionMiddleware(3), api.ResetPassword)
				userGroup.POST("/:id/reset-avatar", middleware.PermissionMiddleware(3), api.ResetAvatar)
			}

			authed.DELETE("/tokens/:id", api.KickUserToken)

			appGroup := authed.Group("/apps")
			{
				appGroup.GET("", api.ListApps)
				appGroup.GET("/simple-list", api.ListAllSimpleApps)
				appGroup.GET("/:id", api.GetApp)
				appGroup.POST("/pre-upload", api.PreUploadApp)
				appGroup.PUT("/:id", api.UpdateApp)
				appGroup.DELETE("/:id", api.DeleteApp)

				appGroup.GET("/tags", api.GetAppTags)
				appGroup.GET("/types", api.GetAppTypes)
				appGroup.GET("/version-types", api.GetAppVersionTypes)

				appGroup.GET("/:id/downloads", api.ListAppDownloads)
				appGroup.POST("/:id/downloads", api.AddAppDownload)
				appGroup.DELETE("/downloads/:download_id", api.DeleteAppDownload)

				adminAppGroup := appGroup.Group("/")
				adminAppGroup.Use(middleware.PermissionMiddleware(1))
				{
					adminAppGroup.POST("/:id/audit", api.AuditApp)
					adminAppGroup.GET("/:id/download-test-url", api.GetAppDownloadTestURL)
					adminAppGroup.POST("/downloads/:download_id/audit", api.AuditAppDownload)
					adminAppGroup.GET("/downloads-to-audit", api.ListDownloadsToAudit)
				}
			}

			noticeGroup := authed.Group("/notices")
			{
				noticeGroup.GET("/unread", api.GetUnreadNotices)
				noticeGroup.POST("/readall", api.MarkAllAsRead)
			}

			operateGroup := authed.Group("/operate")
			operateGroup.Use(middleware.PermissionMiddleware(2))
			{
				operateGroup.POST("/notice", api.SendNotice)
				operateGroup.POST("/popup", api.SendPopup)
				operateGroup.POST("/actions", api.SendActions)
				operateGroup.POST("/email", api.SendEmailToUsers)
			}

			adminGroup := authed.Group("/admin")
			adminGroup.Use(middleware.PermissionMiddleware(2))
			{
				adminGroup.GET("/banners", api.ListBanners)
				adminGroup.POST("/banners", api.CreateBanner)
				adminGroup.PUT("/banners/:id", api.UpdateBanner)
				adminGroup.DELETE("/banners/:id", api.DeleteBanner)

				adminGroup.GET("/banned-ips", api.ListBannedIPs)
				adminGroup.POST("/banned-ips", api.CreateBannedIP)
				adminGroup.DELETE("/banned-ips/:id", api.DeleteBannedIP)

				adminGroup.GET("/prohibited-words", api.ListProhibitedWords)
				adminGroup.POST("/prohibited-words", api.CreateProhibitedWord)
				adminGroup.DELETE("/prohibited-words/:id", api.DeleteProhibitedWord)

				adminGroup.GET("/username-blacklists", api.ListUsernameBlacklists)
				adminGroup.POST("/username-blacklists", api.CreateUsernameBlacklist)
				adminGroup.DELETE("/username-blacklists/:id", api.DeleteUsernameBlacklist)

				adminGroup.GET("/comments", api.ListComments)
				adminGroup.PUT("/comments/:id", api.UpdateComment)
				adminGroup.DELETE("/comments/:id", api.DeleteComment)

				adminGroup.GET("/reports", api.ListReports)
				adminGroup.GET("/reports/:id", api.GetReportDetails)
				adminGroup.POST("/reports/:id/audit", api.AuditReport)

				adminGroup.GET("/settings/:key", api.GetSetting)
				adminGroup.PUT("/settings/:key", api.UpdateSetting)

				pageGroup := adminGroup.Group("/pages")
				{
					pageGroup.GET("", api.ListAppPages)
					pageGroup.POST("", api.CreateAppPage)
					pageGroup.GET("/:id", api.GetAppPage)
					pageGroup.PUT("/:id", api.UpdateAppPage)
					pageGroup.DELETE("/:id", api.DeleteAppPage)
					pageGroup.POST("/:id/sync-apps", api.SyncAppsToPage)
				}
			}
		}
	}
}
