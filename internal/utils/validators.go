package utils

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ValidateEstadoPeticion valida que el ID de estado esté en el rango 20-25
func ValidateEstadoPeticion(id int) error {
	if id < 20 || id > 25 {
		return errors.New("ID de estado de petición inválido. Debe estar entre 20 y 25")
	}
	return nil
}

// ValidateEstadoDespacho valida que el ID de estado esté en el rango 30-34
func ValidateEstadoDespacho(id int) error {
	if id < 30 || id > 34 {
		return errors.New("ID de estado de despacho inválido. Debe estar entre 30 y 34")
	}
	return nil
}

// ValidateEstadoFactura valida que el ID de estado esté en el rango 60-62
func ValidateEstadoFactura(id int) error {
	if id < 60 || id > 62 {
		return errors.New("ID de estado de factura inválido. Debe estar entre 60 y 62")
	}
	return nil
}

// ValidateStateTransition valida si una transición de estado es permitida
// TODO: Implementar lógica específica de transiciones según reglas de negocio
func ValidateStateTransition(from, to int, entity string) error {
	switch entity {
	case "peticion":
		if err := ValidateEstadoPeticion(to); err != nil {
			return err
		}
		// Validaciones específicas de transiciones de petición
		// Por ejemplo: no se puede pasar de Completada (24) a Pendiente (20)
		if from == 24 && to == 20 {
			return errors.New("no se puede cambiar de 'Completada' a 'Pendiente'")
		}
		if from == 25 { // Cancelada
			return errors.New("no se puede cambiar el estado de una petición cancelada")
		}

	case "despacho":
		if err := ValidateEstadoDespacho(to); err != nil {
			return err
		}
		// Validaciones específicas de transiciones de despacho
		if from == 33 && to == 30 {
			return errors.New("no se puede cambiar de 'Entregado' a 'Pendiente'")
		}

	case "factura":
		if err := ValidateEstadoFactura(to); err != nil {
			return err
		}
		// Validaciones específicas de transiciones de factura
		if from == 61 && to == 60 {
			return errors.New("no se puede cambiar de 'Pagada' a 'Pendiente'")
		}
	}

	return nil
}

// ParsePaginationParams extrae y valida los parámetros de paginación
func ParsePaginationParams(c *gin.Context) (page int, limit int) {
	page = 1
	limit = 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			// Límite máximo de 100
			if limit > 100 {
				limit = 100
			}
		}
	}

	return page, limit
}

// CalculateOffset calcula el offset para la consulta de base de datos
func CalculateOffset(page, limit int) int {
	return (page - 1) * limit
}
