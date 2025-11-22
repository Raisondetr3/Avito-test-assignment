package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/dto"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/middleware"
)

type TeamHandler struct {
	teamService TeamService
}

func NewTeamHandler(teamService TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	members := make([]*domain.User, 0, len(req.Members))
	for _, m := range req.Members {
		user := domain.NewUser(m.UserID, m.Username, req.TeamName, m.IsActive)
		members = append(members, user)
	}

	team, err := h.teamService.CreateTeam(r.Context(), req.TeamName, members)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	response := dto.TeamResponse{
		Team: mapTeamToDTO(team),
	}

	middleware.WriteJSON(w, http.StatusCreated, response)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	response := mapTeamToDTO(team)
	middleware.WriteJSON(w, http.StatusOK, response)
}

func mapTeamToDTO(team *domain.Team) *dto.Team {
	members := make([]*dto.TeamMember, 0, len(team.Members))
	for _, m := range team.Members {
		members = append(members, &dto.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	return &dto.Team{
		TeamName: team.TeamName,
		Members:  members,
	}
}
