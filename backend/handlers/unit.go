package handlers

import (
	"net/http"
	"scandata/database"
	"scandata/models"

	"github.com/gin-gonic/gin"
)

type UnitHandler struct{}

func NewUnitHandler() *UnitHandler {
	return &UnitHandler{}
}

type CreateUnitRequest struct {
	QRCode        string `json:"qr_code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	ExpectedGrade string `json:"expected_grade"`
	Location      string `json:"location"`
}

type UpdateUnitRequest struct {
	QRCode        string `json:"qr_code"`
	Name          string `json:"name"`
	ExpectedGrade string `json:"expected_grade"`
	Location      string `json:"location"`
	IsActive      *bool  `json:"is_active"`
}

func (h *UnitHandler) List(c *gin.Context) {
	var units []models.Unit
	query := database.DB

	// Filter by active status
	if active := c.Query("active"); active != "" {
		query = query.Where("is_active = ?", active == "true")
	}

	// Search by name or qr_code
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR qr_code LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Order("created_at DESC").Find(&units)
	c.JSON(http.StatusOK, units)
}

func (h *UnitHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var unit models.Unit
	if err := database.DB.First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}
	c.JSON(http.StatusOK, unit)
}

func (h *UnitHandler) GetByQRCode(c *gin.Context) {
	qrCode := c.Param("qr_code")
	var unit models.Unit
	if err := database.DB.Where("qr_code = ? AND is_active = ?", qrCode, true).First(&unit).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}
	c.JSON(http.StatusOK, unit)
}

func (h *UnitHandler) Create(c *gin.Context) {
	var req CreateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit := &models.Unit{
		QRCode:        req.QRCode,
		Name:          req.Name,
		ExpectedGrade: req.ExpectedGrade,
		Location:      req.Location,
		IsActive:      true,
	}

	if err := database.DB.Create(unit).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR Code already exists"})
		return
	}

	c.JSON(http.StatusCreated, unit)
}

func (h *UnitHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var unit models.Unit
	if err := database.DB.First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	var req UpdateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.QRCode != "" {
		unit.QRCode = req.QRCode
	}
	if req.Name != "" {
		unit.Name = req.Name
	}
	if req.ExpectedGrade != "" {
		unit.ExpectedGrade = req.ExpectedGrade
	}
	if req.Location != "" {
		unit.Location = req.Location
	}
	if req.IsActive != nil {
		unit.IsActive = *req.IsActive
	}

	if err := database.DB.Save(&unit).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR Code already exists"})
		return
	}

	c.JSON(http.StatusOK, unit)
}

func (h *UnitHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var unit models.Unit
	if err := database.DB.First(&unit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
		return
	}

	database.DB.Delete(&unit)
	c.JSON(http.StatusOK, gin.H{"message": "Unit deleted"})
}
