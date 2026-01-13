package handlers

import (
	"fmt"
	"net/http"
	"scandata/database"
	"scandata/models"
	"scandata/services"
	"time"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

type DailyReport struct {
	Date     string `json:"date"`
	Total    int64  `json:"total"`
	Match    int64  `json:"match"`
	NotMatch int64  `json:"not_match"`
}

type UserPerformance struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
	Total    int64  `json:"total"`
	Match    int64  `json:"match"`
	NotMatch int64  `json:"not_match"`
}

func (h *ReportHandler) Summary(c *gin.Context) {
	role := c.MustGet("role").(models.Role)
	userID := c.MustGet("user_id").(uint)

	today := time.Now().Truncate(24 * time.Hour)
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

	var totalToday, matchToday, notMatchToday int64
	var totalWeek, matchWeek, notMatchWeek int64
	var totalMonth, matchMonth, notMatchMonth int64

	// Today
	todayQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ?", today)
	if role != models.RoleAdmin {
		todayQuery = todayQuery.Where("user_id = ?", userID)
	}
	todayQuery.Count(&totalToday)

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

	// This week
	weekQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ?", weekStart)
	if role != models.RoleAdmin {
		weekQuery = weekQuery.Where("user_id = ?", userID)
	}
	weekQuery.Count(&totalWeek)

	weekMatchQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND is_match = ?", weekStart, true)
	if role != models.RoleAdmin {
		weekMatchQuery = weekMatchQuery.Where("user_id = ?", userID)
	}
	weekMatchQuery.Count(&matchWeek)

	weekNotMatchQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND is_match = ?", weekStart, false)
	if role != models.RoleAdmin {
		weekNotMatchQuery = weekNotMatchQuery.Where("user_id = ?", userID)
	}
	weekNotMatchQuery.Count(&notMatchWeek)

	// This month
	monthQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ?", monthStart)
	if role != models.RoleAdmin {
		monthQuery = monthQuery.Where("user_id = ?", userID)
	}
	monthQuery.Count(&totalMonth)

	monthMatchQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND is_match = ?", monthStart, true)
	if role != models.RoleAdmin {
		monthMatchQuery = monthMatchQuery.Where("user_id = ?", userID)
	}
	monthMatchQuery.Count(&matchMonth)

	monthNotMatchQuery := database.DB.Model(&models.ScanLog{}).Where("scanned_at >= ? AND is_match = ?", monthStart, false)
	if role != models.RoleAdmin {
		monthNotMatchQuery = monthNotMatchQuery.Where("user_id = ?", userID)
	}
	monthNotMatchQuery.Count(&notMatchMonth)

	c.JSON(http.StatusOK, gin.H{
		"today": gin.H{
			"total":     totalToday,
			"match":     matchToday,
			"not_match": notMatchToday,
		},
		"week": gin.H{
			"total":     totalWeek,
			"match":     matchWeek,
			"not_match": notMatchWeek,
		},
		"month": gin.H{
			"total":     totalMonth,
			"match":     matchMonth,
			"not_match": notMatchMonth,
		},
	})
}

func (h *ReportHandler) Daily(c *gin.Context) {
	days := 7
	if d := c.Query("days"); d != "" {
		fmt.Sscanf(d, "%d", &days)
	}
	if days > 30 {
		days = 30
	}

	reports := make([]DailyReport, days)
	today := time.Now().Truncate(24 * time.Hour)

	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, -i)
		nextDate := date.Add(24 * time.Hour)

		var total, match, notMatch int64
		database.DB.Model(&models.ScanLog{}).
			Where("scanned_at >= ? AND scanned_at < ?", date, nextDate).Count(&total)
		database.DB.Model(&models.ScanLog{}).
			Where("scanned_at >= ? AND scanned_at < ? AND is_match = ?", date, nextDate, true).Count(&match)
		database.DB.Model(&models.ScanLog{}).
			Where("scanned_at >= ? AND scanned_at < ? AND is_match = ?", date, nextDate, false).Count(&notMatch)

		reports[i] = DailyReport{
			Date:     date.Format("2006-01-02"),
			Total:    total,
			Match:    match,
			NotMatch: notMatch,
		}
	}

	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) UserPerformance(c *gin.Context) {
	today := time.Now().Truncate(24 * time.Hour)
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))

	var users []models.User
	database.DB.Where("role = ?", models.RoleUser).Find(&users)

	performances := make([]UserPerformance, len(users))
	for i, user := range users {
		var total, match, notMatch int64
		database.DB.Model(&models.ScanLog{}).
			Where("user_id = ? AND scanned_at >= ?", user.ID, weekStart).Count(&total)
		database.DB.Model(&models.ScanLog{}).
			Where("user_id = ? AND scanned_at >= ? AND is_match = ?", user.ID, weekStart, true).Count(&match)
		database.DB.Model(&models.ScanLog{}).
			Where("user_id = ? AND scanned_at >= ? AND is_match = ?", user.ID, weekStart, false).Count(&notMatch)

		performances[i] = UserPerformance{
			UserID:   user.ID,
			UserName: user.Name,
			Total:    total,
			Match:    match,
			NotMatch: notMatch,
		}
	}

	c.JSON(http.StatusOK, performances)
}

func (h *ReportHandler) Export(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := database.DB.Model(&models.ScanLog{}).Preload("Unit").Preload("User")

	if startDate != "" {
		if start, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("scanned_at >= ?", start)
		}
	}
	if endDate != "" {
		if end, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("scanned_at < ?", end.Add(24*time.Hour))
		}
	}

	var scans []models.ScanLog
	query.Order("scanned_at DESC").Find(&scans)

	excelFile, err := services.GenerateExcel(scans)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel"})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=scan_report_%s.xlsx", time.Now().Format("20060102_150405")))

	excelFile.Write(c.Writer)
}
