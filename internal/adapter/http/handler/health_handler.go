package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang-chi-hexagonal-base/internal/adapter/postgres"
	"golang-chi-hexagonal-base/internal/adapter/whatsapp"
)

type HealthHandler struct {
	db            *postgres.DB
	clientManager *whatsapp.ClientManager
}

func NewHealthHandler(db *postgres.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// SetClientManager sets the WhatsApp client manager for health checks
func (h *HealthHandler) SetClientManager(cm *whatsapp.ClientManager) {
	h.clientManager = cm
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]Status `json:"services"`
}

type Status struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]Status),
	}

	dbStatus := h.checkDatabase()
	response.Services["database"] = dbStatus

	// Check WhatsApp service if available
	if h.clientManager != nil {
		whatsappStatus := h.checkWhatsApp()
		response.Services["whatsapp"] = whatsappStatus
		
		// WhatsApp being down doesn't make the whole service unavailable
		// It's an optional feature
		if whatsappStatus.Status == "down" {
			response.Status = "degraded"
		}
	}

	if dbStatus.Status == "down" {
		response.Status = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	dbStatus := h.checkDatabase()

	if dbStatus.Status == "down" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "not ready",
			"reason": "database unavailable",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}

func (h *HealthHandler) checkDatabase() Status {
	if h.db == nil {
		return Status{
			Status: "down",
			Details: map[string]string{
				"error": "database connection not initialized",
			},
		}
	}

	err := h.db.Ping()
	if err != nil {
		return Status{
			Status: "down",
			Details: map[string]string{
				"error": err.Error(),
			},
		}
	}

	stats := h.db.Stats()
	return Status{
		Status: "up",
		Details: map[string]string{
			"open_connections": string(rune(stats.OpenConnections)),
			"in_use":           string(rune(stats.InUse)),
			"idle":             string(rune(stats.Idle)),
		},
	}
}

func (h *HealthHandler) checkWhatsApp() Status {
	if h.clientManager == nil {
		return Status{
			Status: "down",
			Details: map[string]string{
				"error": "WhatsApp client manager not initialized",
			},
		}
	}

	clientCount := h.clientManager.GetClientCount()
	clients := h.clientManager.GetAllClients()
	
	connectedCount := 0
	disconnectedCount := 0
	
	for _, client := range clients {
		if client.IsConnected() {
			connectedCount++
		} else {
			disconnectedCount++
		}
	}

	return Status{
		Status: "up",
		Details: map[string]string{
			"total_clients":       fmt.Sprintf("%d", clientCount),
			"connected_clients":   fmt.Sprintf("%d", connectedCount),
			"disconnected_clients": fmt.Sprintf("%d", disconnectedCount),
		},
	}
}
