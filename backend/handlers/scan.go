package handlers

import (
	"net/http"
	"scandata/database"
	"scandata/models"
	"time"

	"github.com/gin-gonic/gin"
)

type ScanHandler struct{}

func NewScanHandler() *ScanHandler {
	return &ScanHandler{}
}

type SubmitScanRequest struct {
	QRCode  string `json:"qr_code" binding:"required"`
	IsMatch bool   `json:"is_match"`
	Notes   string `json:"notes"`
}

func (h *ScanHandler) Submit(c *gin.Context) {
	var req SubmitScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find unit by QR code
	var unit models.Unit
	if err := database.DB.Where("qr_code = ?", req.QRCode).First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	if !unit.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unit is not active"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	scanLog := &models.ScanLog{
		UnitID:    unit.ID,
		UserID:    userID,
		IsMatch:   req.IsMatch,
		Notes:     req.Notes,
		ScannedAt: time.Now(),
	}

	database.DB.Create(scanLog)

	// Load unit info for response
	scanLog.Unit = unit

	c.JSON(http.StatusCreated, scanLog)
}

func (h *ScanHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(models.Role)

	var scans []models.ScanLog
	query := database.DB.Preload("Unit").Preload("User")

	// Regular users can only see their own scans
	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	// Filter by date
	if date := c.Query("date"); date != "" {
		startDate, err := time.Parse("2006-01-02", date)
		if err == nil {
			endDate := startDate.Add(24 * time.Hour)
			query = query.Where("scanned_at >= ? AND scanned_at < ?", startDate, endDate)
		}
	}

	// Filter by match status
	if match := c.Query("is_match"); match != "" {
		query = query.Where("is_match = ?", match == "true")
	}

	query.Order("scanned_at DESC").Limit(100).Find(&scans)
	c.JSON(http.StatusOK, scans)
}

func (h *ScanHandler) GetStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(models.Role)

	today := time.Now().Truncate(24 * time.Hour)

	query := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ?", today)
	if role != models.RoleAdmin {
		query = query.Where("user_id = ?", userID)
	}

	var totalToday int64
	var matchToday int64
	var notMatchToday int64

	query.Count(&totalToday)
	query.Where("is_match = ?", true).Count(&matchToday)
	query.Where("is_match = ?", false).Count(&notMatchToday)

	c.JSON(http.StatusOK, gin.H{
		"today": gin.H{
			"total":     totalToday,
			"match":     matchToday,
			"not_match": notMatchToday,
		},
	})
}
