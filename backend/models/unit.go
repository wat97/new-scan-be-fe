package models

import (
	"time"

	"gorm.io/gorm"
)

type Unit struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	QRCode        string         `gorm:"uniqueIndex;size:100;not null" json:"qr_code"`
	Name          string         `gorm:"size:200;not null" json:"name"`
	ExpectedGrade string         `gorm:"size:100" json:"expected_grade"`
	Location      string         `gorm:"size:200" json:"location"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	ScanLogs      []ScanLog      `gorm:"foreignKey:UnitID" json:"-"`
}
