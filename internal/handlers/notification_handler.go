package handlers

import (
	"net/http"
	"strconv"

	"ms-operations/internal/database"
	"ms-operations/internal/models"
	"ms-operations/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetNotifications godoc
// @Summary Obtener notificaciones (DEPRECADO)
// @Description [DEPRECADO] Obtiene notificaciones del usuario autenticado. Migrar a MS-Notifications
// @Tags Notificaciones
// @Accept json
// @Produce json
// @Param page query int false "Número de página"
// @Param limit query int false "Elementos por página"
// @Param leida query string false "Filtrar por estado leída (true/false)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Lista de notificaciones"
// @Failure 401 {object} map[string]interface{} "No autenticado"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /notificaciones [get]
// NOTA: Este endpoint no está en el contrato MS-OPERATIONS-CONTRACT.md
// Se recomienda migrar a un microservicio dedicado (MS-Notifications)
func GetNotifications(c *gin.Context) {
	// El user_id ahora viene del contexto de autenticación
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Usuario no autenticado"))
		return
	}

	// Permitir override vía query param (para administradores)
	userIDQuery := c.Query("user_id")
	if userIDQuery != "" {
		uid, err := strconv.Atoi(userIDQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse("user_id inválido"))
			return
		}
		userID = uid
	}

	// Paginación
	page, limit := utils.ParsePaginationParams(c)
	offset := utils.CalculateOffset(page, limit)

	// Query base
	query := database.DB.Model(&models.Notificacion{}).Where("id_usuario = ?", userID)

	// Filtro por estado leído/no leído
	if leido := c.Query("leida"); leido != "" {
		leidoBool := leido == "true"
		query = query.Where("leido = ?", leidoBool)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener notificaciones
	var notificaciones []models.Notificacion
	if err := query.Order("fecha DESC").Offset(offset).Limit(limit).Find(&notificaciones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al obtener notificaciones"))
		return
	}

	// Construir respuesta
	var response []gin.H
	for _, n := range notificaciones {
		response = append(response, gin.H{
			"id_notificacion": n.IDNotificacion,
			"titulo":          n.Titulo,
			"mensaje":         n.Cuerpo,
			"leida":           n.Leido,
			"fecha":           n.Fecha.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(response, utils.PaginationMeta(page, limit, int(total))))
}
