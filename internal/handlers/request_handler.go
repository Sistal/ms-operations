package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"ms-operations/internal/database"
	"ms-operations/internal/models"
	"ms-operations/internal/utils"

	"github.com/gin-gonic/gin"
)

// DTOs para Peticiones
type TallajeItemDTO struct {
	IdPrenda   int    `json:"id_prenda" binding:"required"`
	ValorTalla string `json:"valor_talla" binding:"required"`
	Cantidad   int    `json:"cantidad" binding:"required,min=1"`
}

type CreatePeticionDTO struct {
	IdFuncionario  int              `json:"id_funcionario" binding:"required"`
	IdUniforme     int              `json:"id_uniforme" binding:"required"`
	IdTipoPeticion int              `json:"id_tipo_peticion" binding:"required"`
	IdTemporada    *int             `json:"id_temporada"`
	IdEstado       *int             `json:"id_estado"`
	Maternal       bool             `json:"maternal"`
	Observacion    *string          `json:"observacion"`
	Tallaje        []TallajeItemDTO `json:"tallaje" binding:"required,min=1"`
}

type UpdatePeticionDTO struct {
	IdEstado    *int    `json:"id_estado"`
	IdDespacho  *int    `json:"id_despacho"`
	Observacion *string `json:"observacion"`
	Maternal    *bool   `json:"maternal"`
}

type CambiarEstadoPeticionDTO struct {
	IdEstado    int     `json:"id_estado" binding:"required"`
	Observacion *string `json:"observacion"`
}

type CreatePeticionUniformeRequest struct {
	IdUniforme    int    `json:"id_uniforme" binding:"required"`
	Cantidad      int    `json:"cantidad" binding:"required,min=1"`
	Talla         string `json:"talla" binding:"required"`
	Motivo        string `json:"motivo" binding:"required"`
	IdFuncionario int    `json:"id_funcionario" binding:"required"`
}

// GetPeticiones godoc
// @Summary Obtener peticiones
// @Description Obtiene lista paginada de peticiones de uniforme con filtros opcionales
// @Tags Peticiones
// @Accept json
// @Produce json
// @Param page query int false "Número de página"
// @Param limit query int false "Elementos por página"
// @Param id_funcionario query int false "Filtrar por ID de funcionario"
// @Param id_estado query int false "Filtrar por ID de estado"
// @Param id_tipo_peticion query int false "Filtrar por ID de tipo petición"
// @Param id_temporada query int false "Filtrar por ID de temporada"
// @Param maternal query string false "Filtrar por maternal (true/false)"
// @Param con_despacho query string false "Filtrar con/sin despacho (true/false)"
// @Param fecha_desde query string false "Fecha desde (YYYY-MM-DD)"
// @Param fecha_hasta query string false "Fecha hasta (YYYY-MM-DD)"
// @Param search query string false "Búsqueda en observaciones"
// @Param sort_by query string false "Campo para ordenar (default fecha_registro)"
// @Param order query string false "Orden (asc/desc, default desc)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Lista de peticiones con paginación"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /peticiones [get]
func GetPeticiones(c *gin.Context) {
	// Paginación
	page, limit := utils.ParsePaginationParams(c)
	offset := utils.CalculateOffset(page, limit)

	// Construir query base
	query := database.DB.Model(&models.PeticionUniforme{})

	// Filtros
	if idFuncionario := c.Query("id_funcionario"); idFuncionario != "" {
		query = query.Where("id_funcionario = ?", idFuncionario)
	}
	if idEstado := c.Query("id_estado"); idEstado != "" {
		query = query.Where("id_estado = ?", idEstado)
	}
	if idTipoPeticion := c.Query("id_tipo_peticion"); idTipoPeticion != "" {
		query = query.Where("id_tipo_peticion = ?", idTipoPeticion)
	}
	if idTemporada := c.Query("id_temporada"); idTemporada != "" {
		query = query.Where("id_temporada = ?", idTemporada)
	}
	if maternal := c.Query("maternal"); maternal != "" {
		maternalBool := maternal == "true"
		query = query.Where("maternal = ?", maternalBool)
	}
	if conDespacho := c.Query("con_despacho"); conDespacho != "" {
		if conDespacho == "true" {
			query = query.Where("id_despacho IS NOT NULL")
		} else if conDespacho == "false" {
			query = query.Where("id_despacho IS NULL")
		}
	}
	if fechaDesde := c.Query("fecha_desde"); fechaDesde != "" {
		query = query.Where("fecha_registro >= ?", fechaDesde)
	}
	if fechaHasta := c.Query("fecha_hasta"); fechaHasta != "" {
		query = query.Where("fecha_registro <= ?", fechaHasta)
	}
	if search := c.Query("search"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("observación ILIKE ?", searchPattern)
	}

	// Ordenamiento
	sortBy := c.DefaultQuery("sort_by", "fecha_registro")
	order := c.DefaultQuery("order", "desc")
	query = query.Order(sortBy + " " + strings.ToUpper(order))

	// Contar total
	var total int64
	query.Count(&total)

	// Obtener datos con relaciones
	var peticiones []models.PeticionUniforme
	if err := query.
		Preload("Funcionario.Sucursal").
		Preload("Funcionario.Cargo").
		Preload("Uniforme").
		Preload("TipoPeticion").
		Preload("Temporada").
		Preload("Estado").
		Preload("Despacho").
		Offset(offset).
		Limit(limit).
		Find(&peticiones).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al obtener peticiones"))
		return
	}

	// Transformar a respuesta con nombre completo
	var response []gin.H
	for _, p := range peticiones {
		item := gin.H{
			"id_peticion": p.IDPeticion,
			"funcionario": gin.H{
				"id_funcionario":  p.Funcionario.IDFuncionario,
				"nombre_completo": p.Funcionario.NombreCompleto(),
				"rut_funcionario": p.Funcionario.RutFuncionario,
				"email":           p.Funcionario.Email,
				"sucursal": gin.H{
					"id_sucursal":     p.Funcionario.Sucursal.IdSucursal,
					"nombre_sucursal": p.Funcionario.Sucursal.NombreSucursal,
				},
			},
			"uniforme": gin.H{
				"id_uniforme":     p.Uniforme.IdUniforme,
				"nombre_uniforme": p.Uniforme.NombreUniforme,
			},
			"tipo_peticion": gin.H{
				"id_tipo_peticion":     p.TipoPeticion.IdTipoPeticion,
				"nombre_tipo_peticion": p.TipoPeticion.NombreTipoPeticion,
			},
			"estado": gin.H{
				"id_estado":     p.Estado.IDEstado,
				"nombre_estado": p.Estado.NombreEstado,
			},
			"maternal":           p.Maternal,
			"observacion":        p.Observacion,
			"fecha_registro":     p.FechaRegistro.Format("2006-01-02"),
			"fecha_modificacion": nil,
		}

		if p.FechaModificacion != nil {
			item["fecha_modificacion"] = p.FechaModificacion.Format("2006-01-02")
		}

		if p.Temporada != nil {
			item["temporada"] = gin.H{
				"id_temporada":     p.Temporada.IdTemporada,
				"nombre_temporada": p.Temporada.NombreTemporada,
			}
		}

		if p.Despacho != nil {
			item["despacho"] = gin.H{
				"id_despacho":      p.Despacho.IDDespacho,
				"guia_de_despacho": p.Despacho.GuiaDeDespacho,
			}
		}

		response = append(response, item)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(response, utils.PaginationMeta(page, limit, int(total))))
}

// GetPeticionByID godoc
// @Summary Obtener petición por ID
// @Description Obtiene detalles completos de una petición específica incluyendo tallajes
// @Tags Peticiones
// @Accept json
// @Produce json
// @Param id_peticion path int true "ID de la petición"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Detalles de la petición"
// @Failure 400 {object} map[string]interface{} "ID inválido"
// @Failure 404 {object} map[string]interface{} "Petición no encontrada"
// @Router /peticiones/{id_peticion} [get]
func GetPeticionByID(c *gin.Context) {
	idParam := c.Param("id_peticion")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("ID de petición inválido"))
		return
	}

	var peticion models.PeticionUniforme
	if err := database.DB.
		Preload("Funcionario.Sucursal").
		Preload("Funcionario.Cargo").
		Preload("Uniforme").
		Preload("TipoPeticion").
		Preload("Temporada").
		Preload("Estado").
		Preload("Despacho.EstadoDespacho").
		Preload("Tallajes.Prenda.TipoPrenda").
		Preload("Tallajes.Prenda.Genero").
		First(&peticion, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Petición no encontrada"))
		return
	}

	// Construir respuesta detallada
	response := gin.H{
		"id_peticion": peticion.IDPeticion,
		"funcionario": gin.H{
			"id_funcionario":  peticion.Funcionario.IDFuncionario,
			"nombre_completo": peticion.Funcionario.NombreCompleto(),
			"rut_funcionario": peticion.Funcionario.RutFuncionario,
			"email":           peticion.Funcionario.Email,
			"celular":         peticion.Funcionario.Celular,
			"sucursal": gin.H{
				"id_sucursal":     peticion.Funcionario.Sucursal.IdSucursal,
				"nombre_sucursal": peticion.Funcionario.Sucursal.NombreSucursal,
				"direccion":       peticion.Funcionario.Sucursal.Direccion,
			},
			"cargo": gin.H{
				"id_cargo":     peticion.Funcionario.Cargo.IdCargo,
				"nombre_cargo": peticion.Funcionario.Cargo.NombreCargo,
			},
		},
		"uniforme": gin.H{
			"id_uniforme":     peticion.Uniforme.IdUniforme,
			"nombre_uniforme": peticion.Uniforme.NombreUniforme,
			"descripcion":     peticion.Uniforme.Descripcion,
		},
		"tipo_peticion": gin.H{
			"id_tipo_peticion":     peticion.TipoPeticion.IdTipoPeticion,
			"nombre_tipo_peticion": peticion.TipoPeticion.NombreTipoPeticion,
		},
		"estado": gin.H{
			"id_estado":     peticion.Estado.IDEstado,
			"nombre_estado": peticion.Estado.NombreEstado,
			"tabla_estado":  peticion.Estado.TablaEstado,
		},
		"maternal":           peticion.Maternal,
		"observacion":        peticion.Observacion,
		"fecha_registro":     peticion.FechaRegistro.Format("2006-01-02"),
		"fecha_modificacion": nil,
	}

	if peticion.FechaModificacion != nil {
		response["fecha_modificacion"] = peticion.FechaModificacion.Format("2006-01-02")
	}

	if peticion.Temporada != nil {
		response["temporada"] = gin.H{
			"id_temporada":     peticion.Temporada.IdTemporada,
			"nombre_temporada": peticion.Temporada.NombreTemporada,
			"fecha_inicio":     peticion.Temporada.FechaInicio,
			"fecha_fin":        peticion.Temporada.FechaFin,
		}
	}

	if peticion.Despacho != nil {
		response["despacho"] = gin.H{
			"id_despacho":      peticion.Despacho.IDDespacho,
			"guia_de_despacho": peticion.Despacho.GuiaDeDespacho,
			"fecha_despacho":   peticion.Despacho.FechaDespacho,
			"sucursal":         peticion.Despacho.Sucursal,
			"estado_despacho": gin.H{
				"id_estado":     peticion.Despacho.EstadoDespacho.IDEstado,
				"nombre_estado": peticion.Despacho.EstadoDespacho.NombreEstado,
			},
		}
	}

	// Tallaje
	var tallajes []gin.H
	for _, t := range peticion.Tallajes {
		tallajes = append(tallajes, gin.H{
			"id_tallaje": t.IDTallaje,
			"prenda": gin.H{
				"id_prenda":     t.Prenda.IDPrenda,
				"nombre_prenda": t.Prenda.NombrePrenda,
				"tipo_prenda":   t.Prenda.TipoPrenda.NombreTipoPrenda,
			},
			"valor_talla": t.ValorTalla,
			"cantidad":    t.Cantidad,
		})
	}
	response["tallaje"] = tallajes

	c.JSON(http.StatusOK, utils.SuccessDataResponse(response))
}

// CreatePeticion godoc
// @Summary Crear petición
// @Description Crea una nueva petición de uniforme con tallajes asociados
// @Tags Peticiones
// @Accept json
// @Produce json
// @Param input body CreatePeticionDTO true "Datos de la petición"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Petición creada exitosamente"
// @Failure 400 {object} map[string]interface{} "Datos de entrada inválidos"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /peticiones [post]
func CreatePeticion(c *gin.Context) {
	var dto CreatePeticionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos: "+err.Error()))
		return
	}

	// Validar estado (default 15 si no se envía)
	estadoID := 15
	if dto.IdEstado != nil {
		estadoID = *dto.IdEstado
		if err := utils.ValidateEstadoPeticion(estadoID); err != nil {
			c.JSON(http.StatusBadRequest, utils.ValidationErrorResponse("Error en la validación", []gin.H{
				{"field": "id_estado", "message": err.Error()},
			}))
			return
		}
	}

	// Validar que funcionario existe
	var funcionario models.Funcionario
	if err := database.DB.First(&funcionario, dto.IdFuncionario).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ValidationErrorResponse("Error en la validación", []gin.H{
			{"field": "id_funcionario", "message": "El funcionario no existe"},
		}))
		return
	}

	// Validar que uniforme existe
	var uniforme models.Uniforme
	if err := database.DB.First(&uniforme, dto.IdUniforme).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ValidationErrorResponse("Error en la validación", []gin.H{
			{"field": "id_uniforme", "message": "El uniforme no existe"},
		}))
		return
	}

	// Transacción para crear petición y tallajes
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	peticion := models.PeticionUniforme{
		IdFuncionario:  dto.IdFuncionario,
		IdUniforme:     dto.IdUniforme,
		IdTipoPeticion: dto.IdTipoPeticion,
		IdTemporada:    dto.IdTemporada,
		IdEstado:       estadoID,
		Maternal:       dto.Maternal,
		Observacion:    dto.Observacion,
		FechaRegistro:  now,
	}

	if err := tx.Create(&peticion).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al crear petición"))
		return
	}

	// Crear tallajes
	tallajesRegistrados := 0
	for _, item := range dto.Tallaje {
		tallaje := models.Tallaje{
			IDPeticion: peticion.IDPeticion,
			IDPrenda:   item.IdPrenda,
			ValorTalla: item.ValorTalla,
			Cantidad:   item.Cantidad,
		}
		if err := tx.Create(&tallaje).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al registrar tallaje"))
			return
		}
		tallajesRegistrados++
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al confirmar transacción"))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessMessageResponse("Petición creada exitosamente", gin.H{
		"id_peticion":        peticion.IDPeticion,
		"id_funcionario":     peticion.IdFuncionario,
		"id_uniforme":        peticion.IdUniforme,
		"id_tipo_peticion":   peticion.IdTipoPeticion,
		"id_estado":          peticion.IdEstado,
		"maternal":           peticion.Maternal,
		"fecha_registro":     peticion.FechaRegistro.Format("2006-01-02"),
		"tallaje_registrado": tallajesRegistrados,
	}))
}

// UpdatePeticion godoc
// @Summary Actualizar petición
// @Description Actualiza campos de una petición existente
// @Tags Peticiones
// @Accept json
// @Produce json
// @Param id_peticion path int true "ID de la petición"
// @Param input body UpdatePeticionDTO true "Datos a actualizar"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Petición actualizada"
// @Failure 400 {object} map[string]interface{} "Datos de entrada inválidos"
// @Failure 404 {object} map[string]interface{} "Petición no encontrada"
// @Router /peticiones/{id_peticion} [put]
func UpdatePeticion(c *gin.Context) {
	idParam := c.Param("id_peticion")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("ID de petición inválido"))
		return
	}

	var dto UpdatePeticionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos"))
		return
	}

	var peticion models.PeticionUniforme
	if err := database.DB.First(&peticion, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Petición no encontrada"))
		return
	}

	// Actualizar campos si se enviaron
	updates := make(map[string]interface{})
	if dto.IdEstado != nil {
		if err := utils.ValidateEstadoPeticion(*dto.IdEstado); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse(err.Error()))
			return
		}
		updates["id_estado"] = *dto.IdEstado
	}
	if dto.IdDespacho != nil {
		updates["id_despacho"] = *dto.IdDespacho
	}
	if dto.Observacion != nil {
		updates["observación"] = *dto.Observacion
	}
	if dto.Maternal != nil {
		updates["maternal"] = *dto.Maternal
	}

	now := time.Now()
	updates["fecha_modificacion"] = now

	if err := database.DB.Model(&peticion).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al actualizar petición"))
		return
	}

	// Recargar petición con estado actualizado
	database.DB.Preload("Estado").First(&peticion, id)

	c.JSON(http.StatusOK, utils.SuccessMessageResponse("Petición actualizada exitosamente", gin.H{
		"id_peticion": peticion.IDPeticion,
		"id_estado":   peticion.IdEstado,
		"estado": gin.H{
			"id_estado":     peticion.Estado.IDEstado,
			"nombre_estado": peticion.Estado.NombreEstado,
		},
		"fecha_modificacion": now.Format("2006-01-02"),
	}))
}

// CambiarEstadoPeticion godoc
// @Summary Cambiar estado de petición
// @Description Cambia el estado de una petición con validación de transiciones
// @Tags Peticiones
// @Accept json
// @Produce json
// @Param id_peticion path int true "ID de la petición"
// @Param input body CambiarEstadoPeticionDTO true "Nuevo estado"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Estado actualizado"
// @Failure 400 {object} map[string]interface{} "Transición de estado inválida"
// @Failure 404 {object} map[string]interface{} "Petición no encontrada"
// @Router /peticiones/{id_peticion}/estado [patch]
func CambiarEstadoPeticion(c *gin.Context) {
	idParam := c.Param("id_peticion")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("ID de petición inválido"))
		return
	}

	var dto CambiarEstadoPeticionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos"))
		return
	}

	var peticion models.PeticionUniforme
	if err := database.DB.Preload("Estado").First(&peticion, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Petición no encontrada"))
		return
	}

	// Validar transición de estado
	if err := utils.ValidateStateTransition(peticion.IdEstado, dto.IdEstado, "peticion"); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponseWithCode("INVALID_STATE_TRANSITION", err.Error()))
		return
	}

	estadoAnterior := peticion.Estado
	now := time.Now()

	// Actualizar estado
	updates := map[string]interface{}{
		"id_estado":          dto.IdEstado,
		"fecha_modificacion": now,
	}
	if dto.Observacion != nil {
		updates["observación"] = *dto.Observacion
	}

	if err := database.DB.Model(&peticion).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al cambiar estado"))
		return
	}

	// Obtener nuevo estado
	var estadoNuevo models.Estado
	database.DB.First(&estadoNuevo, dto.IdEstado)

	c.JSON(http.StatusOK, utils.SuccessMessageResponse("Estado de petición actualizado", gin.H{
		"id_peticion": peticion.IDPeticion,
		"estado_anterior": gin.H{
			"id_estado":     estadoAnterior.IDEstado,
			"nombre_estado": estadoAnterior.NombreEstado,
		},
		"estado_nuevo": gin.H{
			"id_estado":     estadoNuevo.IDEstado,
			"nombre_estado": estadoNuevo.NombreEstado,
		},
		"fecha_modificacion": now.Format("2006-01-02"),
	}))
}

// GetTallajePeticion godoc
// @Summary Obtener tallaje de petición
// @Description Obtiene el tallaje completo de una petición específica
// @Tags Peticiones
// @Accept json
// @Produce json
// @Param id_peticion path int true "ID de la petición"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Tallaje de la petición"
// @Failure 400 {object} map[string]interface{} "ID inválido"
// @Failure 404 {object} map[string]interface{} "Petición no encontrada"
// @Router /peticiones/{id_peticion}/tallaje [get]
func GetTallajePeticion(c *gin.Context) {
	idParam := c.Param("id_peticion")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("ID de petición inválido"))
		return
	}

	// Verificar que la petición existe
	var peticion models.PeticionUniforme
	if err := database.DB.First(&peticion, id).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Petición no encontrada"))
		return
	}

	// Obtener tallajes con relaciones
	var tallajes []models.Tallaje
	if err := database.DB.
		Preload("Prenda.TipoPrenda").
		Preload("Prenda.Genero").
		Where("id_peticion = ?", id).
		Find(&tallajes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al obtener tallaje"))
		return
	}

	// Construir respuesta
	var response []gin.H
	totalUnidades := 0
	for _, t := range tallajes {
		response = append(response, gin.H{
			"id_tallaje":  t.IDTallaje,
			"id_peticion": t.IDPeticion,
			"prenda": gin.H{
				"id_prenda":     t.Prenda.IDPrenda,
				"nombre_prenda": t.Prenda.NombrePrenda,
				"tipo_prenda": gin.H{
					"id_tipo_prenda":     t.Prenda.TipoPrenda.IdTipoPrenda,
					"nombre_tipo_prenda": t.Prenda.TipoPrenda.NombreTipoPrenda,
				},
				"genero": gin.H{
					"id_genero":     t.Prenda.Genero.IdGenero,
					"nombre_genero": t.Prenda.Genero.NombreGenero,
				},
			},
			"valor_talla": t.ValorTalla,
			"cantidad":    t.Cantidad,
		})
		totalUnidades += t.Cantidad
	}

	meta := gin.H{
		"total_prendas":  len(tallajes),
		"total_unidades": totalUnidades,
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(response, meta))
}

// CreatePeticionUniforme godoc
// @Summary Crear petición simplificada de uniforme
// @Description Crea una nueva petición de uniforme a partir de talla única y cantidad de uniformes
// @Tags Peticiones
// @Accept json
// @Produce json
// @Param input body CreatePeticionUniformeRequest true "Datos de la petición"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "Petición creada exitosamente"
// @Failure 400 {object} map[string]interface{} "Datos de entrada inválidos"
// @Failure 500 {object} map[string]interface{} "Error del servidor"
// @Router /peticiones/uniforme [post]
func CreatePeticionUniforme(c *gin.Context) {
	var req CreatePeticionUniformeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Datos de entrada inválidos: "+err.Error()))
		return
	}

	// 1. Validar Funcionario
	var funcionario models.Funcionario
	if err := database.DB.First(&funcionario, req.IdFuncionario).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("El funcionario no existe"))
		return
	}

	// 2. Validar Uniforme
	var uniforme models.Uniforme
	if err := database.DB.First(&uniforme, req.IdUniforme).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("El uniforme no existe"))
		return
	}

	// 3. Resolver TipoPeticion (Motivo)
	var tipoPeticion models.TipoPeticion
	if err := database.DB.Where("LOWER(nombre_tipo_peticion) = LOWER(?)", req.Motivo).First(&tipoPeticion).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("El motivo (tipo petición) no es válido"))
		return
	}

	// 4. Obtener composición del Uniforme (Prendas)
	var uniformePrendas []models.UniformePrenda
	if err := database.DB.Where("id_uniforme = ?", req.IdUniforme).Find(&uniformePrendas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al obtener composición del uniforme"))
		return
	}

	if len(uniformePrendas) == 0 {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("El uniforme seleccionado no tiene prendas asociadas o configuradas"))
		return
	}

	// 5. Transacción
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generar ID manual para Peticion (Fix: falta sequence en DB)
	var maxPeticionID int
	if err := tx.Model(&models.PeticionUniforme{}).Select("COALESCE(MAX(id_peticion), 0)").Scan(&maxPeticionID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al generar ID de petición"))
		return
	}
	newPeticionID := maxPeticionID + 1

	peticion := models.PeticionUniforme{
		IDPeticion:     newPeticionID,
		IdFuncionario:  req.IdFuncionario,
		IdUniforme:     req.IdUniforme,
		IdTipoPeticion: tipoPeticion.IdTipoPeticion,
		IdEstado:       15, // Estado Pendiente (según DB)
		FechaRegistro:  time.Now(),
	}

	if err := tx.Create(&peticion).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al crear la petición"))
		return
	}

	// Obtener base para IDs de Tallaje
	var maxTallajeID int
	if err := tx.Model(&models.Tallaje{}).Select("COALESCE(MAX(id_tallaje), 0)").Scan(&maxTallajeID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al generar ID de tallaje"))
		return
	}

	// 6. Crear detalles de Tallaje
	for i, up := range uniformePrendas {
		cantidadTotal := up.Cantidad * req.Cantidad
		tallaje := models.Tallaje{
			IDTallaje:  maxTallajeID + 1 + i,
			IDPeticion: peticion.IDPeticion,
			IDPrenda:   up.IdPrenda,
			ValorTalla: req.Talla,
			Cantidad:   cantidadTotal,
		}

		if err := tx.Create(&tallaje).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al registrar tallaje para prenda ID: "+strconv.Itoa(up.IdPrenda)))
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Error al confirmar la transacción"))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessMessageResponse("Petición creada exitosamente", gin.H{
		"id_peticion":    peticion.IDPeticion,
		"id_funcionario": peticion.IdFuncionario,
		"id_uniforme":    peticion.IdUniforme,
		"motivo":         req.Motivo,
		"fecha_registro": peticion.FechaRegistro.Format("2006-01-02"),
	}))
}
