package handlers

import (
	"net/http"
	"strconv"

	"ms-operations/internal/database"
	"ms-operations/internal/models"

	"github.com/gin-gonic/gin"
)

// GET /stats/dashboard
func GetDashboardStats(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	// 1. Unread Notifications (Direct User ID)
	var unreadCount int64
	database.DB.Model(&models.Notificacion{}).Where("id_usuario = ? AND leido = ?", userID, false).Count(&unreadCount)

	// Get Funcionario ID
	var funcionario models.Funcionario
	if err := database.DB.Where("id_usuario = ?", userID).First(&funcionario).Error; err != nil {
		// If no funcionario, return 0 for everything
		c.JSON(http.StatusOK, gin.H{
			"pending_requests":     0,
			"last_delivery_status": "N/A",
			"unread_notifications": unreadCount,
		})
		return
	}

	// 2. Pending Requests (Funcionario ID)
	// Assuming State 1 = Pendiente
	var pendingCount int64
	database.DB.Model(&models.PeticionUniforme{}).
		Where("id_funcionario = ? AND id_estado IN (?)", funcionario.IDFuncionario, []int{1, 2}).
		Count(&pendingCount)

	// 3. Last Delivery Status
	var lastPeticion models.PeticionUniforme
	// Find latest peticion with a despacho
	var lastDeliveryStatusString = "Sin envíos"

	err = database.DB.Where("id_funcionario = ? AND id_despacho IS NOT NULL", funcionario.IDFuncionario).
		Order("fecha_registro desc").
		First(&lastPeticion).Error

	if err == nil && lastPeticion.IdDespacho != nil {
		var despacho models.Despacho
		if err := database.DB.Preload("EstadoDespacho").First(&despacho, *lastPeticion.IdDespacho).Error; err == nil {
			lastDeliveryStatusString = despacho.EstadoDespacho.NombreEstado
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_requests":     pendingCount,
		"last_delivery_status": lastDeliveryStatusString,
		"unread_notifications": unreadCount,
	})
}
