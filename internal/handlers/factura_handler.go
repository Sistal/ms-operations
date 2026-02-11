package handlers

import (
	"net/http"
	"time"

	"ms-operations/internal/database"
	"ms-operations/internal/models"
	"ms-operations/internal/utils"

	"github.com/gin-gonic/gin"
)

// DTOs para Facturas
type CreateFacturaDTO struct {
	NumeroFactura string  `json:"numero_factura" binding:"required"`
	FechaEmision  *string `json:"fecha_emision"`
	MontoTotal    float64 `json:"monto_total" binding:"required,min=0"`
	IdEstado      *int    `json:"id_estado"`
	IdEmpresa     int     `json:"id_empresa" binding:"required"`
}

// GetFacturas godoc
// @Summary Obtener facturas
// @Description Obtiene lista paginada de facturas con filtros opcionales
// @Tags Facturas
// @Accept json
// @Produce json
// @Param page query int false "Número de página"
// @Param limit query int false "Elementos por página"
// @Param id_estado query int false "Filtrar por ID de estado"
// @Param id_empresa query int false "Filtrar por ID de empresa"
// @Param fecha_desde query string false "Fecha desde (YYYY-MM-DD)"
// @Param fecha_hasta query string false "Fecha hasta (YYYY-MM-DD)"
// @Param search query string false "Búsqueda en número de factura"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Lista de facturas con paginación"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /facturas [get]
func GetFacturas(c *gin.Context) {
	// Paginación
	page, limit := utils.ParsePaginationParams(c)
	offset := utils.CalculateOffset(page, limit)

	// Construir query base
	query := database.DB.Model(&models.Factura{})

	// Filtros
	if idEstado := c.Query("id_estado"); idEstado != "" {
		query = query.Where("id_estado = ?", idEstado)
	}
	if idEmpresa := c.Query("id_empresa"); idEmpresa != "" {
		query = query.Where("id_empresa = ?", idEmpresa)
	}
	if fechaDesde := c.Query("fecha_desde"); fechaDesde != "" {
		query = query.Where("fecha_emision >= ?", fechaDesde)
	}
	if fechaHasta := c.Query("fecha_hasta"); fechaHasta != "" {
		query = query.Where("fecha_emision <= ?", fechaHasta)
	}
	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("numero_factura ILIKE ?", searchPattern)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener datos con relaciones
	var facturas []models.Factura
	if err := query.
		Preload("Estado").
		Preload("Empresa").
		Offset(offset).
		Limit(limit).
		Order("fecha_emision DESC").
		Find(&facturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al obtener facturas"))
		return
	}

	// Construir respuesta
	var response []gin.H
	for _, f := range facturas {
		// Contar despachos asociados
		var despachosCount int64
		database.DB.Model(&models.Despacho{}).Where("id_factura = ?", f.IdFactura).Count(&despachosCount)

		response = append(response, gin.H{
			"id_factura":     f.IdFactura,
			"numero_factura": f.NumeroFactura,
			"fecha_emision":  f.FechaEmision.Format("2006-01-02"),
			"monto_total":    f.MontoTotal,
			"estado": gin.H{
				"id_estado":     f.Estado.IDEstado,
				"nombre_estado": f.Estado.NombreEstado,
			},
			"empresa": gin.H{
				"id_empresa":     f.Empresa.IdEmpresa,
				"nombre_empresa": f.Empresa.NombreEmpresa,
				"rut_empresa":    f.Empresa.RutEmpresa,
			},
			"despachos_asociados": despachosCount,
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(response, utils.PaginationMeta(page, limit, int(total))))
}

// CreateFactura godoc
// @Summary Crear factura
// @Description Crea una nueva factura en el sistema
// @Tags Facturas
// @Accept json
// @Produce json
// @Param input body CreateFacturaDTO true "Datos de la factura"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Factura creada exitosamente"
// @Failure 400 {object} map[string]interface{} "Datos de entrada inválidos"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /facturas [post]
func CreateFactura(c *gin.Context) {
	var dto CreateFacturaDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos: "+err.Error()))
		return
	}

	// Validar número de factura único
	var count int64
	database.DB.Model(&models.Factura{}).Where("numero_factura = ?", dto.NumeroFactura).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponseWithCode("DUPLICATE_FACTURA", "El número de factura ya existe"))
		return
	}

	// Validar que empresa existe
	var empresa models.Empresa
	if err := database.DB.First(&empresa, dto.IdEmpresa).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ValidationErrorResponse("Error en la validación", []gin.H{
			{"field": "id_empresa", "message": "La empresa no existe"},
		}))
		return
	}

	// Estado default 60 (Pendiente)
	estadoID := 60
	if dto.IdEstado != nil {
		estadoID = *dto.IdEstado
		if err := utils.ValidateEstadoFactura(estadoID); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
			return
		}
	}

	// Fecha de emisión (default hoy)
	fechaEmision := time.Now()
	if dto.FechaEmision != nil {
		parsedDate, err := time.Parse("2006-01-02", *dto.FechaEmision)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse("Formato de fecha inválido. Usar: YYYY-MM-DD"))
			return
		}
		fechaEmision = parsedDate
	}

	factura := models.Factura{
		NumeroFactura: dto.NumeroFactura,
		FechaEmision:  fechaEmision,
		MontoTotal:    dto.MontoTotal,
		IdEstado:      estadoID,
		IdEmpresa:     dto.IdEmpresa,
	}

	if err := database.DB.Create(&factura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al crear factura"))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessMessageResponse("Factura creada exitosamente", gin.H{
		"id_factura":     factura.IdFactura,
		"numero_factura": factura.NumeroFactura,
		"fecha_emision":  factura.FechaEmision.Format("2006-01-02"),
		"monto_total":    factura.MontoTotal,
		"id_estado":      factura.IdEstado,
		"id_empresa":     factura.IdEmpresa,
	}))
}
