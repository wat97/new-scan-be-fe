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
	Barcode string `json:"barcode" binding:"required"`
	IsMatch bool   `json:"is_match"`
	Notes   string `json:"notes"`
}

// Submit - Langsung simpan hasil scan ke database
func (h *ScanHandler) Submit(c *gin.Context) {
	var req SubmitScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uint)

	scanLog := &models.ScanLog{
		Barcode:   req.Barcode,
		UserID:    userID,
		IsMatch:   req.IsMatch,
		Notes:     req.Notes,
		ScannedAt: time.Now(),
	}

	if err := database.DB.Create(scanLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save scan"})
		return
	}

	// Load user info for response
	database.DB.Preload("User").First(scanLog, scanLog.ID)

	c.JSON(http.StatusCreated, scanLog)
}

// List - Tampilkan history scan
func (h *ScanHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(models.Role)

	var scans []models.ScanLog
	query := database.DB.Preload("User")

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

	// Filter by barcode
	if barcode := c.Query("barcode"); barcode != "" {
		query = query.Where("barcode LIKE ?", "%"+barcode+"%")
	}

	// Filter by match status
	if match := c.Query("is_match"); match != "" {
		query = query.Where("is_match = ?", match == "true")
	}

	query.Order("scanned_at DESC").Limit(100).Find(&scans)
	c.JSON(http.StatusOK, scans)
}

// GetStats - Statistik scan hari ini
func (h *ScanHandler) GetStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(models.Role)

	today := time.Now().Truncate(24 * time.Hour)

	baseQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ?", today)
	if role != models.RoleAdmin {
		baseQuery = baseQuery.Where("user_id = ?", userID)
	}

	var totalToday int64
	var matchToday int64
	var notMatchToday int64

	baseQuery.Count(&totalToday)
	
	matchQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND is_match = ?", today, true)
	if role != models.RoleAdmin {
		matchQuery = matchQuery.Where("user_id = ?", userID)
	}
	matchQuery.Count(&matchToday)

	notMatchQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND is_match = ?", today, false)
	if role != models.RoleAdmin {
		notMatchQuery = notMatchQuery.Where("user_id = ?", userID)
	}
	notMatchQuery.Count(&notMatchToday)

	c.JSON(http.StatusOK, gin.H{
		"today": gin.H{
			"total":     totalToday,
			"match":     matchToday,
			"not_match": notMatchToday,
		},
	})
}
