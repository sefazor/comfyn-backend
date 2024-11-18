package database

import (
	"fmt"
	"log"
	"os"

	"github.com/sefazor/comfyn/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	// PostgreSQL sürücü ayarlarını yapılandır
	pgConfig := postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}

	config := &gorm.Config{
		PrepareStmt:                              false,
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	}

	var err error
	DB, err = gorm.Open(postgres.New(pgConfig), config)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance: ", err)
	}

	// Bağlantıyı test et
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	// Connection pool ayarları
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connection established successfully")

	// Migrationları çalıştır
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Post{},
		&models.Product{},
		&models.Category{},
		&models.Like{},
		&models.Comment{},
		&models.Hashtag{},
		&models.Notification{},
		&models.NotificationPreference{},
		&models.PostView{},
		&models.AffiliateLink{},
		&models.ClickLog{},
	); err != nil {
		log.Printf("Warning: Migration issues: %v", err)
	} else {
		log.Println("Migrations completed successfully")
	}
}
