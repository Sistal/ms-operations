package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ms-operations/internal/database"
	"ms-operations/internal/models"
	"ms-operations/internal/utils"

	"github.com/gin-gonic/gin"
)

// DTOs para Despachos
type CreateDespachoDTO struct {
	GuiaDeDespacho       string  `json:"guia_de_despacho" binding:"required"`
	FechaDespacho        *string `json:"fecha_despacho"`
	NumeroVoucher        *string `json:"numero_voucher"`
	Sucursal             string  `json:"sucursal" binding:"required"`
	ResponsableRecepcion *string `json:"responsable_recepcion"`
	IdEstadoDespacho     *int    `json:"id_estado_despacho"`
	IdFactura            *int    `json:"id_factura"`
}

type UpdateDespachoDTO struct {
	NumeroVoucher        *string `json:"numero_voucher"`
	ResponsableRecepcion *string `json:"responsable_recepcion"`
	IdEstadoDespacho     *int    `json:"id_estado_despacho"`
	IdFactura            *int    `json:"id_factura"`
}

type CambiarEstadoDespachoDTO struct {
	IdEstadoDespacho     int     `json:"id_estado_despacho" binding:"required"`
	ResponsableRecepcion *string `json:"responsable_recepcion"`
}

// GetDespachos godoc
// @Summary Obtener despachos
// @Description Obtiene lista paginada de despachos con filtros opcionales
// @Tags Despachos
// @Accept json
// @Produce json
// @Param page query int false "Número de página"
// @Param limit query int false "Elementos por página"
// @Param id_estado_despacho query int false "Filtrar por ID de estado"
// @Param sucursal query string false "Filtrar por sucursal"
// @Param fecha_desde query string false "Fecha desde (YYYY-MM-DD)"
// @Param fecha_hasta query string false "Fecha hasta (YYYY-MM-DD)"
// @Param search query string false "Búsqueda en guía o voucher"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Lista de despachos con paginación"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /despachos [get]
func GetDespachos(c *gin.Context) {
	// Paginación
	page, limit := utils.ParsePaginationParams(c)
	offset := utils.CalculateOffset(page, limit)

	// Construir query base
	query := database.DB.Model(&models.Despacho{})

	// Filtros
	if idEstadoDespacho := c.Query("id_estado_despacho"); idEstadoDespacho != "" {
		query = query.Where("id_estado_despacho = ?", idEstadoDespacho)
	}
	if sucursal := c.Query("sucursal"); sucursal != "" {
		query = query.Where("sucursal ILIKE ?", "%"+sucursal+"%")
	}
	if fechaDesde := c.Query("fecha_desde"); fechaDesde != "" {
		query = query.Where("fecha_despacho >= ?", fechaDesde)
	}
	if fechaHasta := c.Query("fecha_hasta"); fechaHasta != "" {
		query = query.Where("fecha_despacho <= ?", fechaHasta)
	}
	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("guia_de_despacho ILIKE ? OR numero_voucher ILIKE ?", searchPattern, searchPattern)
	}

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener datos con relaciones
	var despachos []models.Despacho
	if err := query.
		Preload("EstadoDespacho").
		Preload("Factura").
		Offset(offset).
		Limit(limit).
		Order("fecha_despacho DESC").
		Find(&despachos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al obtener despachos"))
		return
	}

	// Construir respuesta
	var response []gin.H
	for _, d := range despachos {
		// Contar peticiones asociadas
		var peticionesCount int64
		database.DB.Model(&models.PeticionUniforme{}).Where("id_despacho = ?", d.IDDespacho).Count(&peticionesCount)

		item := gin.H{
			"id_despacho":           d.IDDespacho,
			"guia_de_despacho":      d.GuiaDeDespacho,
			"fecha_despacho":        nil,
			"numero_voucher":        d.NumeroVoucher,
			"sucursal":              d.Sucursal,
			"responsable_recepcion": d.ResponsableRecepcion,
			"estado_despacho": gin.H{
				"id_estado":     d.EstadoDespacho.IDEstado,
				"nombre_estado": d.EstadoDespacho.NombreEstado,
			},
			"peticiones_asociadas": peticionesCount,
		}

		if d.FechaDespacho != nil {
			item["fecha_despacho"] = d.FechaDespacho.Format("2006-01-02")
		}

		if d.Factura != nil {
			item["factura"] = gin.H{
				"id_factura":     d.Factura.IdFactura,
				"numero_factura": d.Factura.NumeroFactura,
			}
		}

		response = append(response, item)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(response, utils.PaginationMeta(page, limit, int(total))))
}

// GetDespachoByID godoc
// @Summary Obtener despacho por ID
// @Description Obtiene detalles completos de un despacho específico
// @Tags Despachos
// @Accept json
// @Produce json
// @Param id_despacho path int true "ID del despacho"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Detalles del despacho"
// @Failure 400 {object} map[string]interface{} "ID inválido"
// @Failure 404 {object} map[string]interface{} "Despacho no encontrado"
// @Router /despachos/{id_despacho} [get]
func GetDespachoByID(c *gin.Context) {
	idParam := c.Param("id_despacho")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("ID de despacho inválido"))
		return
	}

	var despacho models.Despacho
	if err := database.DB.
		Preload("EstadoDespacho").
		Preload("Factura.Estado").
		Preload("Factura.Empresa").
		First(&despacho, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Despacho no encontrado"))
		return
	}

	// Obtener peticiones asociadas
	var peticiones []models.PeticionUniforme
	database.DB.
		Preload("Funcionario").
		Preload("Uniforme").
		Where("id_despacho = ?", id).
		Find(&peticiones)

	// Construir respuesta detallada
	response := gin.H{
		"id_despacho":           despacho.IDDespacho,
		"guia_de_despacho":      despacho.GuiaDeDespacho,
		"fecha_despacho":        nil,
		"numero_voucher":        despacho.NumeroVoucher,
		"sucursal":              despacho.Sucursal,
		"responsable_recepcion": despacho.ResponsableRecepcion,
		"estado_despacho": gin.H{
			"id_estado":     despacho.EstadoDespacho.IDEstado,
			"nombre_estado": despacho.EstadoDespacho.NombreEstado,
			"tabla_estado":  "Despacho",
		},
	}

	if despacho.FechaDespacho != nil {
		response["fecha_despacho"] = despacho.FechaDespacho.Format("2006-01-02")
	}

	if despacho.Factura != nil {
		response["factura"] = gin.H{
			"id_factura":     despacho.Factura.IdFactura,
			"numero_factura": despacho.Factura.NumeroFactura,
			"fecha_emision":  despacho.Factura.FechaEmision.Format("2006-01-02"),
			"monto_total":    despacho.Factura.MontoTotal,
			"empresa": gin.H{
				"id_empresa":     despacho.Factura.Empresa.IdEmpresa,
				"nombre_empresa": despacho.Factura.Empresa.NombreEmpresa,
			},
		}
	}

	// Peticiones asociadas
	var peticionesResponse []gin.H
	for _, p := range peticiones {
		peticionesResponse = append(peticionesResponse, gin.H{
			"id_peticion": p.IDPeticion,
			"funcionario": gin.H{
				"nombre_completo": p.Funcionario.NombreCompleto(),
				"rut_funcionario": p.Funcionario.RutFuncionario,
			},
			"uniforme": gin.H{
				"nombre_uniforme": p.Uniforme.NombreUniforme,
			},
		})
	}
	response["peticiones"] = peticionesResponse

	c.JSON(http.StatusOK, utils.SuccessDataResponse(response))
}

// CreateDespacho godoc
// @Summary Crear despacho
// @Description Crea un nuevo despacho en el sistema
// @Tags Despachos
// @Accept json
// @Produce json
// @Param input body CreateDespachoDTO true "Datos del despacho"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Despacho creado exitosamente"
// @Failure 400 {object} map[string]interface{} "Datos de entrada inválidos"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /despachos [post]
func CreateDespacho(c *gin.Context) {
	var dto CreateDespachoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos: "+err.Error()))
		return
	}

	// Validar guía única
	var count int64
	database.DB.Model(&models.Despacho{}).Where("guia_de_despacho = ?", dto.GuiaDeDespacho).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponseWithCode("DUPLICATE_GUIA", "La guía de despacho ya existe"))
		return
	}

	// Estado default 30 (Pendiente)
	estadoID := 30
	if dto.IdEstadoDespacho != nil {
		estadoID = *dto.IdEstadoDespacho
		if err := utils.ValidateEstadoDespacho(estadoID); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
			return
		}
	}

	// Fecha de despacho (default hoy)
	var fechaDespacho *time.Time
	if dto.FechaDespacho != nil {
		parsedDate, err := time.Parse("2006-01-02", *dto.FechaDespacho)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse("Formato de fecha inválido. Usar: YYYY-MM-DD"))
			return
		}
		fechaDespacho = &parsedDate
	} else {
		now := time.Now()
		fechaDespacho = &now
	}

	despacho := models.Despacho{
		GuiaDeDespacho:       dto.GuiaDeDespacho,
		FechaDespacho:        fechaDespacho,
		NumeroVoucher:        dto.NumeroVoucher,
		Sucursal:             dto.Sucursal,
		ResponsableRecepcion: dto.ResponsableRecepcion,
		IdEstadoDespacho:     estadoID,
		IdFactura:            dto.IdFactura,
	}

	if err := database.DB.Create(&despacho).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al crear despacho"))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessMessageResponse("Despacho creado exitosamente", gin.H{
		"id_despacho":        despacho.IDDespacho,
		"guia_de_despacho":   despacho.GuiaDeDespacho,
		"fecha_despacho":     fechaDespacho.Format("2006-01-02"),
		"sucursal":           despacho.Sucursal,
		"id_estado_despacho": despacho.IdEstadoDespacho,
	}))
}

// UpdateDespacho godoc
// @Summary Actualizar despacho
// @Description Actualiza campos de un despacho existente
// @Tags Despachos
// @Accept json
// @Produce json
// @Param id_despacho path int true "ID del despacho"
// @Param input body UpdateDespachoDTO true "Datos a actualizar"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Despacho actualizado"
// @Failure 400 {object} map[string]interface{} "Datos de entrada inválidos"
// @Failure 404 {object} map[string]interface{} "Despacho no encontrado"
// @Router /despachos/{id_despacho} [put]
func UpdateDespacho(c *gin.Context) {
	idParam := c.Param("id_despacho")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("ID de despacho inválido"))
		return
	}

	var dto UpdateDespachoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos"))
		return
	}

	var despacho models.Despacho
	if err := database.DB.First(&despacho, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Despacho no encontrado"))
		return
	}

	// Actualizar campos si se enviaron
	updates := make(map[string]interface{})
	if dto.NumeroVoucher != nil {
		updates["numero_voucher"] = *dto.NumeroVoucher
	}
	if dto.ResponsableRecepcion != nil {
		updates["responsable_recepcion"] = *dto.ResponsableRecepcion
	}
	if dto.IdEstadoDespacho != nil {
		if err := utils.ValidateEstadoDespacho(*dto.IdEstadoDespacho); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
			return
		}
		updates["id_estado_despacho"] = *dto.IdEstadoDespacho
	}
	if dto.IdFactura != nil {
		updates["id_factura"] = *dto.IdFactura
	}

	if err := database.DB.Model(&despacho).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al actualizar despacho"))
		return
	}

	// Recargar despacho
	database.DB.First(&despacho, id)

	c.JSON(http.StatusOK, utils.SuccessMessageResponse("Despacho actualizado exitosamente", gin.H{
		"id_despacho":           despacho.IDDespacho,
		"guia_de_despacho":      despacho.GuiaDeDespacho,
		"numero_voucher":        despacho.NumeroVoucher,
		"responsable_recepcion": despacho.ResponsableRecepcion,
		"id_estado_despacho":    despacho.IdEstadoDespacho,
	}))
}

// CambiarEstadoDespacho godoc
// @Summary Cambiar estado de despacho
// @Description Cambia el estado de un despacho
// @Tags Despachos
// @Accept json
// @Produce json
// @Param id_despacho path int true "ID del despacho"
// @Param input body CambiarEstadoDespachoDTO true "Nuevo estado"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Estado actualizado"
// @Failure 400 {object} map[string]interface{} "Transición de estado inválida"
// @Failure 404 {object} map[string]interface{} "Despacho no encontrado"
// @Router /despachos/{id_despacho}/estado [patch]
func CambiarEstadoDespacho(c *gin.Context) {
	idParam := c.Param("id_despacho")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("ID de despacho inválido"))
		return
	}

	var dto CambiarEstadoDespachoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos"))
		return
	}

	var despacho models.Despacho
	if err := database.DB.Preload("EstadoDespacho").First(&despacho, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Despacho no encontrado"))
		return
	}

	// Validar transición de estado
	if err := utils.ValidateStateTransition(despacho.IdEstadoDespacho, dto.IdEstadoDespacho, "despacho"); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponseWithCode("INVALID_STATE_TRANSITION", err.Error()))
		return
	}

	estadoAnterior := despacho.EstadoDespacho

	// Actualizar estado
	updates := map[string]interface{}{
		"id_estado_despacho": dto.IdEstadoDespacho,
	}
	if dto.ResponsableRecepcion != nil {
		updates["responsable_recepcion"] = *dto.ResponsableRecepcion
	}

	if err := database.DB.Model(&despacho).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al cambiar estado"))
		return
	}

	// Obtener nuevo estado
	var estadoNuevo models.Estado
	database.DB.First(&estadoNuevo, dto.IdEstadoDespacho)

	response := gin.H{
		"id_despacho": despacho.IDDespacho,
		"estado_anterior": gin.H{
			"id_estado":     estadoAnterior.IDEstado,
			"nombre_estado": estadoAnterior.NombreEstado,
		},
		"estado_nuevo": gin.H{
			"id_estado":     estadoNuevo.IDEstado,
			"nombre_estado": estadoNuevo.NombreEstado,
		},
	}

	if dto.ResponsableRecepcion != nil {
		response["responsable_recepcion"] = *dto.ResponsableRecepcion
	}

	c.JSON(http.StatusOK, utils.SuccessMessageResponse("Estado de despacho actualizado", response))
}
