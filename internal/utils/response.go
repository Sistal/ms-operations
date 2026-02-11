package utils

import "github.com/gin-gonic/gin"

// SuccessResponse retorna una respuesta exitosa con datos y metadata de paginación
func SuccessResponse(data interface{}, meta gin.H) gin.H {
	return gin.H{
		"success": true,
		"data":    data,
		"meta":    meta,
	}
}

// SuccessDataResponse retorna una respuesta exitosa solo con datos (sin paginación)
func SuccessDataResponse(data interface{}) gin.H {
	return gin.H{
		"success": true,
		"data":    data,
	}
}

// SuccessMessageResponse retorna una respuesta exitosa con mensaje y datos opcionales
func SuccessMessageResponse(message string, data interface{}) gin.H {
	response := gin.H{
		"success": true,
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	return response
}

// ErrorResponse retorna una respuesta de error con mensaje
func ErrorResponse(message string) gin.H {
	return gin.H{
		"success": false,
		"message": message,
	}
}

// ErrorResponseWithCode retorna una respuesta de error con código y mensaje
func ErrorResponseWithCode(code string, message string) gin.H {
	return gin.H{
		"success": false,
		"message": message,
		"error": gin.H{
			"code": code,
		},
	}
}

// ValidationErrorResponse retorna una respuesta de error de validación con detalles
func ValidationErrorResponse(message string, errors []gin.H) gin.H {
	return gin.H{
		"success": false,
		"message": message,
		"errors":  errors,
	}
}

// PaginationMeta calcula y retorna metadata de paginación
func PaginationMeta(page, limit, total int) gin.H {
	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return gin.H{
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	}
}
