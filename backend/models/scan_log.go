package models

import (
	"time"
)

type ScanLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Barcode   string    `gorm:"size:100;not null;index" json:"barcode"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	IsMatch   bool      `gorm:"not null" json:"is_match"`
	Notes     string    `gorm:"size:500" json:"notes"`
	ScannedAt time.Time `gorm:"not null;index" json:"scanned_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
