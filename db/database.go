package db

import (
	"fmt"
	"log"
	"market-api/models"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	var err error
	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	dbname := viper.GetString("database.dbname")
	charset := viper.GetString("database.charset")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		user, password, host, port, dbname, charset)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database object: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = DB.AutoMigrate(
		&models.User{},
		&models.UserToken{},
		&models.Notice{},
		&models.App{},
		&models.AppReply{},
		&models.Splash{},
		&models.UserDownloadCount{},
		&models.AppTag{},
		&models.AppType{},
		&models.AppVersionType{},
		&models.AppDownload{},
		&models.Popup{},
		&models.UserAction{},
		&models.Banner{},
		&models.BannedIP{},
		&models.ProhibitedWord{},
		&models.UsernameBlacklist{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	fmt.Println("Database connection successful.")
}
