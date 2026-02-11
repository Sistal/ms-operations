package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware valida el header Authorization Bearer token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Header de autorización requerido",
			})
			c.Abort()
			return
		}

		// Verificar formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Formato de autorización inválido. Usar: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token de autorización vacío",
			})
			c.Abort()
			return
		}

		// TODO: Validar token JWT con MS-Authentication
		// Por ahora, aceptamos cualquier token y extraemos claims simulados
		// En producción, esto debe validarse contra el servicio de autenticación

		// Simular extracción de claims (reemplazar con validación JWT real)
		c.Set("user_id", 1)
		c.Set("role", "admin")
		c.Set("token", token)

		c.Next()
	}
}

// RequireRole middleware para validar que el usuario tenga uno de los roles permitidos
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "No se pudo determinar el rol del usuario",
				"error": gin.H{
					"code": "PERMISSION_DENIED",
				},
			})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Rol de usuario inválido",
				"error": gin.H{
					"code": "PERMISSION_DENIED",
				},
			})
			c.Abort()
			return
		}

		// Verificar si el rol del usuario está en la lista de roles permitidos
		allowed := false
		for _, role := range roles {
			if roleStr == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Sin permisos suficientes",
				"error": gin.H{
					"code": "PERMISSION_DENIED",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
