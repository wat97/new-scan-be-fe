package database

import (
	"log"
	"scandata/config"
	"scandata/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate tables
	err = DB.AutoMigrate(&models.User{}, &models.Unit{}, &models.ScanLog{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Create default admin if not exists
	createDefaultAdmin()

	log.Println("Database connected and migrated successfully")
}

func createDefaultAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&count)
	if count == 0 {
		admin := &models.User{
			Username: "admin",
			Name:     "Administrator",
			Role:     models.RoleAdmin,
		}
		admin.SetPassword("admin123")
		DB.Create(admin)
		log.Println("Default admin created (username: admin, password: admin123)")
	}
}
