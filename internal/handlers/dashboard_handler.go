package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"ms-operations/internal/database"
	"ms-operations/internal/models"

	"github.com/gin-gonic/gin"
)

// Response structures
type DashboardStatsResponse struct {
	UserID                     int          `json:"user_id"` // Refers to id_funcionario
	TotalSolicitudes           int64        `json:"total_solicitudes"`
	EntregasEntregadas         int64        `json:"entregas_entregadas"` // Delivered
	EntregasProximas           int64        `json:"entregas_proximas"`   // In progress / Upcoming
	SolicitudesPendientes      int64        `json:"solicitudes_pendientes"`
	SolicitudesRequierenAccion int64        `json:"solicitudes_requieren_accion"`
	RecentItems                []RecentItem `json:"recent_items"`
}

type RecentItem struct {
	Fecha       string   `json:"fecha"`
	Estado      string   `json:"estado"`
	Items       []string `json:"items"`
	Description string   `json:"description"`
}

// GET /operaciones/dashboard
func GetDashboard(c *gin.Context) {
	// 1. Get id_funcionario parameter
	// The requirement says "receive the id of a functionary". Using "user_id" as query param to match common patterns,
	// or checking for "id_funcionario". The user example has "user_id" in response.
	// I'll check both for flexibility, prioritizing "id_funcionario".
	idFuncionarioStr := c.Query("id_funcionario")
	if idFuncionarioStr == "" {
		// Fallback to user_id if id_funcionario is not present
		idFuncionarioStr = c.Query("user_id")
	}

	if idFuncionarioStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_funcionario or user_id is required"})
		return
	}

	idFuncionario, err := strconv.Atoi(idFuncionarioStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id_funcionario"})
		return
	}

	// 2. Fetch Aggregated Stats
	// We need to count PeticionUniforme by State for this official.
	// Since we don't know the exact IDs, we join with Estado and group by NombreEstado.

	type StateCount struct {
		NombreEstado string
		Count        int64
	}

	var stateCounts []StateCount
	if err := database.DB.Model(&models.PeticionUniforme{}).
		Select("est.nombre_estado, count(*) as count").
		Joins("join \"Estado\" est on est.id_estado = \"Petición Uniforme\".id_estado").
		Where("\"Petición Uniforme\".id_funcionario = ?", idFuncionario).
		Group("est.nombre_estado").
		Scan(&stateCounts).Error; err != nil {
		log.Printf("Error fetching dashboard stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching statistics"})
		return
	}

	// 3. Process Counts
	var stats DashboardStatsResponse
	stats.UserID = idFuncionario

	for _, sc := range stateCounts {
		stats.TotalSolicitudes += sc.Count

		lowerState := strings.ToLower(sc.NombreEstado)

		// Map states to categories based on name keywords
		if strings.Contains(lowerState, "entregad") || strings.Contains(lowerState, "finalizad") || strings.Contains(lowerState, "recibido") {
			stats.EntregasEntregadas += sc.Count
		} else if strings.Contains(lowerState, "pendiente") || strings.Contains(lowerState, "ingresad") || strings.Contains(lowerState, "solicitado") {
			stats.SolicitudesPendientes += sc.Count
		} else if strings.Contains(lowerState, "rechazad") || strings.Contains(lowerState, "accion") || strings.Contains(lowerState, "revisión") || strings.Contains(lowerState, "deuelt") { // devuelto/deuelta
			stats.SolicitudesRequierenAccion += sc.Count
		} else {
			// Default to "Proximas" / "En Proceso" for everything else (En camino, Aprobado, Preparando, etc.)
			// Or strictly "En proceso"
			stats.EntregasProximas += sc.Count
		}
	}

	// 4. Fetch Recent Items
	var petitions []models.PeticionUniforme
	if err := database.DB.Preload("Estado").Preload("Tallajes.Prenda").Preload("Uniforme").
		Where("id_funcionario = ?", idFuncionario).
		Order("fecha_registro desc").
		Limit(5).
		Find(&petitions).Error; err != nil {
		log.Printf("Error fetching recent items: %v", err)
		// We continue, returning empty list
	}

	stats.RecentItems = make([]RecentItem, 0)
	for _, p := range petitions {
		item := RecentItem{
			Fecha:       p.FechaRegistro.Format("2006-01-02T15:04:05Z07:00"),
			Estado:      p.Estado.NombreEstado,
			Description: p.Uniforme.NombreUniforme,
			Items:       make([]string, 0),
		}

		// Build items list from Tallajes
		if len(p.Tallajes) > 0 {
			for _, t := range p.Tallajes {
				// E.g. "Zapatos de seguridad"
				item.Items = append(item.Items, t.Prenda.NombrePrenda)
			}
		} else {
			// Fallback if no tallajes
			item.Items = append(item.Items, p.Uniforme.NombreUniforme)
		}

		stats.RecentItems = append(stats.RecentItems, item)
	}

	c.JSON(http.StatusOK, stats)
}
