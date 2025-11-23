package handlers

import (
	"net/http"

	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/middleware"
)

type StatsHandler struct {
	statsService StatsService
}

func NewStatsHandler(statsService StatsService) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
	}
}

func (h *StatsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsService.GetStatistics(r.Context())
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	middleware.WriteJSON(w, http.StatusOK, stats)
}
