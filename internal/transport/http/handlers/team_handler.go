package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/dto"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/middleware"
)

type TeamHandler struct {
	teamService             TeamService
	bulkDeactivationService BulkDeactivationService
}

func NewTeamHandler(teamService TeamService, bulkDeactivationService BulkDeactivationService) *TeamHandler {
	return &TeamHandler{
		teamService:             teamService,
		bulkDeactivationService: bulkDeactivationService,
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

func (h *TeamHandler) BulkDeactivateUsers(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkDeactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.TeamName == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	result, err := h.bulkDeactivationService.DeactivateUsersAndReassignPRs(r.Context(), req.TeamName, req.UserIDs)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	response := dto.BulkDeactivateResponse{
		DeactivatedUsers: result.DeactivatedUsers,
		ReassignedPRs:    mapReassignedPRsToDTO(result.ReassignedPRs),
		SkippedPRs:       mapSkippedPRsToDTO(result.SkippedPRs),
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

func mapReassignedPRsToDTO(prs []domain.ReassignedPR) []dto.ReassignedPRInfo {
	result := make([]dto.ReassignedPRInfo, 0, len(prs))
	for _, pr := range prs {
		result = append(result, dto.ReassignedPRInfo{
			PullRequestID: pr.PullRequestID,
			OldReviewerID: pr.OldReviewerID,
			NewReviewerID: pr.NewReviewerID,
		})
	}
	return result
}

func mapSkippedPRsToDTO(prs []domain.SkippedPR) []dto.SkippedPRInfo {
	result := make([]dto.SkippedPRInfo, 0, len(prs))
	for _, pr := range prs {
		result = append(result, dto.SkippedPRInfo{
			PullRequestID: pr.PullRequestID,
			Reason:        pr.Reason,
		})
	}
	return result
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
