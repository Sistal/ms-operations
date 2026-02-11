package handlers

import (
	"net/http"
	"time"

	"ms-operations/internal/database"
	"ms-operations/internal/models"
	"ms-operations/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetEstadisticasOperaciones godoc
// @Summary Obtener estadísticas de operaciones
// @Description Obtiene estadísticas de peticiones, despachos y facturas con filtros opcionales
// @Tags Operaciones
// @Accept json
// @Produce json
// @Param fecha_desde query string false "Fecha desde (YYYY-MM-DD)"
// @Param fecha_hasta query string false "Fecha hasta (YYYY-MM-DD)"
// @Param id_sucursal query int false "Filtrar por ID de sucursal"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Estadísticas completas"
// @Router /operaciones/estadisticas [get]
func GetEstadisticasOperaciones(c *gin.Context) {
	// Filtros opcionales
	fechaDesde := c.Query("fecha_desde")
	fechaHasta := c.Query("fecha_hasta")
	idSucursal := c.Query("id_sucursal")

	// Estadísticas de Peticiones
	queryPeticiones := database.DB.Model(&models.PeticionUniforme{})
	if fechaDesde != "" {
		queryPeticiones = queryPeticiones.Where("fecha_registro >= ?", fechaDesde)
	}
	if fechaHasta != "" {
		queryPeticiones = queryPeticiones.Where("fecha_registro <= ?", fechaHasta)
	}
	if idSucursal != "" {
		queryPeticiones = queryPeticiones.Joins("JOIN \"Funcionario\" ON \"Funcionario\".id_funcionario = \"Petición Uniforme\".id_funcionario").
			Where("\"Funcionario\".id_sucursal = ?", idSucursal)
	}

	var totalPeticiones int64
	queryPeticiones.Count(&totalPeticiones)

	// Por estado (20-25)
	var porEstado []struct {
		IdEstado int
		Count    int64
	}
	queryPeticiones.Select("id_estado, COUNT(*) as count").Group("id_estado").Scan(&porEstado)

	estadosPeticiones := gin.H{
		"pendientes":  int64(0),
		"aprobadas":   int64(0),
		"rechazadas":  int64(0),
		"en_proceso":  int64(0),
		"completadas": int64(0),
		"canceladas":  int64(0),
	}
	for _, e := range porEstado {
		switch e.IdEstado {
		case 20:
			estadosPeticiones["pendientes"] = e.Count
		case 21:
			estadosPeticiones["aprobadas"] = e.Count
		case 22:
			estadosPeticiones["rechazadas"] = e.Count
		case 23:
			estadosPeticiones["en_proceso"] = e.Count
		case 24:
			estadosPeticiones["completadas"] = e.Count
		case 25:
			estadosPeticiones["canceladas"] = e.Count
		}
	}

	// Por tipo de petición
	var porTipo []struct {
		IdTipoPeticion int
		Count          int64
	}
	queryPeticiones.Select("id_tipo_peticion, COUNT(*) as count").Group("id_tipo_peticion").Scan(&porTipo)

	tiposPeticiones := gin.H{
		"nuevo_ingreso":    int64(0),
		"reposicion":       int64(0),
		"cambio_talla":     int64(0),
		"prenda_malograda": int64(0),
		"maternal":         int64(0),
		"cambio_segmento":  int64(0),
	}
	for _, t := range porTipo {
		switch t.IdTipoPeticion {
		case 1:
			tiposPeticiones["nuevo_ingreso"] = t.Count
		case 2:
			tiposPeticiones["reposicion"] = t.Count
		case 3:
			tiposPeticiones["cambio_talla"] = t.Count
		case 4:
			tiposPeticiones["prenda_malograda"] = t.Count
		case 5:
			tiposPeticiones["maternal"] = t.Count
		case 6:
			tiposPeticiones["cambio_segmento"] = t.Count
		}
	}

	// Peticiones maternales
	var maternales int64
	queryPeticiones.Where("maternal = ?", true).Count(&maternales)

	// Estadísticas de Despachos
	queryDespachos := database.DB.Model(&models.Despacho{})
	if fechaDesde != "" {
		queryDespachos = queryDespachos.Where("fecha_despacho >= ?", fechaDesde)
	}
	if fechaHasta != "" {
		queryDespachos = queryDespachos.Where("fecha_despacho <= ?", fechaHasta)
	}
	if idSucursal != "" {
		// Filtrar por sucursal del despacho
		var sucursal models.Sucursal
		database.DB.First(&sucursal, idSucursal)
		queryDespachos = queryDespachos.Where("sucursal = ?", sucursal.NombreSucursal)
	}

	var totalDespachos int64
	queryDespachos.Count(&totalDespachos)

	// Por estado (30-34)
	var porEstadoDespacho []struct {
		IdEstadoDespacho int
		Count            int64
	}
	queryDespachos.Select("id_estado_despacho, COUNT(*) as count").Group("id_estado_despacho").Scan(&porEstadoDespacho)

	estadosDespachos := gin.H{
		"pendientes":  int64(0),
		"preparando":  int64(0),
		"en_transito": int64(0),
		"entregados":  int64(0),
		"devueltos":   int64(0),
	}
	for _, e := range porEstadoDespacho {
		switch e.IdEstadoDespacho {
		case 30:
			estadosDespachos["pendientes"] = e.Count
		case 31:
			estadosDespachos["preparando"] = e.Count
		case 32:
			estadosDespachos["en_transito"] = e.Count
		case 33:
			estadosDespachos["entregados"] = e.Count
		case 34:
			estadosDespachos["devueltos"] = e.Count
		}
	}

	// Por sucursal
	var porSucursal []struct {
		Sucursal string
		Count    int64
	}
	queryDespachos.Select("sucursal, COUNT(*) as count").Group("sucursal").Scan(&porSucursal)

	var sucursalesResponse []gin.H
	for _, s := range porSucursal {
		sucursalesResponse = append(sucursalesResponse, gin.H{
			"sucursal": s.Sucursal,
			"total":    s.Count,
		})
	}

	// Estadísticas de Facturas
	queryFacturas := database.DB.Model(&models.Factura{})
	if fechaDesde != "" {
		queryFacturas = queryFacturas.Where("fecha_emision >= ?", fechaDesde)
	}
	if fechaHasta != "" {
		queryFacturas = queryFacturas.Where("fecha_emision <= ?", fechaHasta)
	}

	var totalFacturas int64
	queryFacturas.Count(&totalFacturas)

	// Monto total
	var totalMonto struct {
		Total float64
	}
	queryFacturas.Select("SUM(monto_total) as total").Scan(&totalMonto)

	// Por estado (60-62)
	var porEstadoFactura []struct {
		IdEstado int
		Count    int64
		Monto    float64
	}
	queryFacturas.Select("id_estado, COUNT(*) as count, SUM(monto_total) as monto").Group("id_estado").Scan(&porEstadoFactura)

	estadosFacturas := gin.H{
		"pendientes": int64(0),
		"pagadas":    int64(0),
		"vencidas":   int64(0),
	}
	montoPendiente := float64(0)
	montoPagado := float64(0)

	for _, e := range porEstadoFactura {
		switch e.IdEstado {
		case 60:
			estadosFacturas["pendientes"] = e.Count
			montoPendiente = e.Monto
		case 61:
			estadosFacturas["pagadas"] = e.Count
			montoPagado = e.Monto
		case 62:
			estadosFacturas["vencidas"] = e.Count
		}
	}

	// Construir respuesta
	response := gin.H{
		"peticiones": gin.H{
			"total":      totalPeticiones,
			"por_estado": estadosPeticiones,
			"por_tipo":   tiposPeticiones,
			"maternales": maternales,
		},
		"despachos": gin.H{
			"total":        totalDespachos,
			"por_estado":   estadosDespachos,
			"por_sucursal": sucursalesResponse,
		},
		"facturas": gin.H{
			"total":           totalFacturas,
			"total_monto":     totalMonto.Total,
			"por_estado":      estadosFacturas,
			"monto_pendiente": montoPendiente,
			"monto_pagado":    montoPagado,
		},
		"periodo": gin.H{
			"fecha_desde": fechaDesde,
			"fecha_hasta": fechaHasta,
		},
	}

	c.JSON(http.StatusOK, utils.SuccessDataResponse(response))
}

// GetDashboardOperaciones godoc
// @Summary Dashboard de operaciones
// @Description Obtiene resumen de operaciones para dashboard (hoy, pendientes, alertas, actividad reciente)
// @Tags Operaciones
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Datos del dashboard"
// @Router /operaciones/dashboard [get]
func GetDashboardOperaciones(c *gin.Context) {
	hoy := time.Now().Format("2006-01-02")

	// Resumen de hoy
	var peticionesNuevas, peticionesAprobadas int64
	database.DB.Model(&models.PeticionUniforme{}).Where("DATE(fecha_registro) = ?", hoy).Count(&peticionesNuevas)
	database.DB.Model(&models.PeticionUniforme{}).Where("DATE(fecha_modificacion) = ? AND id_estado = ?", hoy, 21).Count(&peticionesAprobadas)

	var despachosEntregados int64
	database.DB.Model(&models.Despacho{}).Where("DATE(fecha_despacho) = ? AND id_estado_despacho = ?", hoy, 33).Count(&despachosEntregados)

	resumenHoy := gin.H{
		"peticiones_nuevas":    peticionesNuevas,
		"peticiones_aprobadas": peticionesAprobadas,
		"despachos_entregados": despachosEntregados,
	}

	// Pendientes de atención
	var peticionesPendientes, despachosEnTransito, facturasPorPagar int64
	database.DB.Model(&models.PeticionUniforme{}).Where("id_estado = ?", 20).Count(&peticionesPendientes)
	database.DB.Model(&models.Despacho{}).Where("id_estado_despacho = ?", 32).Count(&despachosEnTransito)
	database.DB.Model(&models.Factura{}).Where("id_estado = ?", 60).Count(&facturasPorPagar)

	pendientesAtencion := gin.H{
		"peticiones_pendientes": peticionesPendientes,
		"despachos_en_transito": despachosEnTransito,
		"facturas_por_pagar":    facturasPorPagar,
	}

	// Alertas
	var alertas []gin.H

	// Facturas próximas a vencer (ejemplo: más de 30 días)
	var facturasProximasVencer int64
	treintaDiasAtras := time.Now().AddDate(0, 0, -30)
	database.DB.Model(&models.Factura{}).
		Where("id_estado = ? AND fecha_emision <= ?", 60, treintaDiasAtras).
		Count(&facturasProximasVencer)

	if facturasProximasVencer > 0 {
		alertas = append(alertas, gin.H{
			"tipo":      "warning",
			"mensaje":   "facturas próximas a vencer",
			"prioridad": "media",
			"cantidad":  facturasProximasVencer,
		})
	}

	if despachosEnTransito > 0 {
		alertas = append(alertas, gin.H{
			"tipo":      "info",
			"mensaje":   "despachos en tránsito",
			"prioridad": "baja",
			"cantidad":  despachosEnTransito,
		})
	}

	// Actividad reciente (últimos 10 registros)
	var actividadReciente []gin.H

	// Últimas peticiones modificadas
	var peticionesRecientes []models.PeticionUniforme
	database.DB.Preload("Funcionario").Preload("Estado").
		Order("fecha_modificacion DESC NULLS LAST, fecha_registro DESC").
		Limit(5).
		Find(&peticionesRecientes)

	for _, p := range peticionesRecientes {
		fecha := p.FechaRegistro
		if p.FechaModificacion != nil {
			fecha = *p.FechaModificacion
		}
		actividadReciente = append(actividadReciente, gin.H{
			"tipo":        "peticion",
			"id":          p.IDPeticion,
			"accion":      "Petición actualizada",
			"funcionario": p.Funcionario.NombreCompleto(),
			"fecha":       fecha.Format(time.RFC3339),
		})
	}

	// Últimos despachos
	var despachosRecientes []models.Despacho
	database.DB.Order("fecha_despacho DESC NULLS LAST").
		Limit(5).
		Find(&despachosRecientes)

	for _, d := range despachosRecientes {
		if d.FechaDespacho != nil {
			actividadReciente = append(actividadReciente, gin.H{
				"tipo":   "despacho",
				"id":     d.IDDespacho,
				"accion": "Despacho procesado",
				"guia":   d.GuiaDeDespacho,
				"fecha":  d.FechaDespacho.Format(time.RFC3339),
			})
		}
	}

	// Respuesta
	response := gin.H{
		"resumen_hoy":         resumenHoy,
		"pendientes_atencion": pendientesAtencion,
		"alertas":             alertas,
		"actividad_reciente":  actividadReciente,
	}

	c.JSON(http.StatusOK, utils.SuccessDataResponse(response))
}
