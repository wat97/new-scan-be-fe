package models

import (
	"time"
)

type ScanLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UnitID    uint      `gorm:"not null;index" json:"unit_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	IsMatch   bool      `gorm:"not null" json:"is_match"`
	Notes     string    `gorm:"size:500" json:"notes"`
	ScannedAt time.Time `gorm:"not null;index" json:"scanned_at"`
	Unit      Unit      `gorm:"foreignKey:UnitID" json:"unit,omitempty"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
