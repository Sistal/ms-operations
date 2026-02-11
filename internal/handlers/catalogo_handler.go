package handlers

import (
	"net/http"

	"ms-operations/internal/database"
	"ms-operations/internal/models"
	"ms-operations/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetTiposPeticion godoc
// @Summary Obtener tipos de petición
// @Description Obtiene catálogo de tipos de petición disponibles
// @Tags Catálogos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Lista de tipos de petición"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /tipos-peticion [get]
func GetTiposPeticion(c *gin.Context) {
	var tipos []models.TipoPeticion
	if err := database.DB.Find(&tipos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al obtener tipos de petición"))
		return
	}

	var response []gin.H
	for _, t := range tipos {
		response = append(response, gin.H{
			"id_tipo_peticion":     t.IdTipoPeticion,
			"nombre_tipo_peticion": t.NombreTipoPeticion,
		})
	}

	meta := gin.H{
		"total": len(tipos),
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(response, meta))
}
