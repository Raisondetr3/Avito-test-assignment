package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/dto"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/middleware"
)

type PRHandler struct {
	prService PRService
}

func NewPRHandler(prService PRService) *PRHandler {
	return &PRHandler{
		prService: prService,
	}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	response := dto.PRResponse{
		PR: mapPRToDTO(pr),
	}

	middleware.WriteJSON(w, http.StatusCreated, response)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req dto.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	response := dto.PRResponse{
		PR: mapPRToDTO(pr),
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

func (h *PRHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	var req dto.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	response := dto.ReassignResponse{
		PR:         mapPRToDTO(pr),
		ReplacedBy: newReviewerID,
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

func mapPRToDTO(pr *domain.PullRequest) *dto.PullRequest {
	return &dto.PullRequest{
		PullRequestID:     pr.PullRequestID,
		PullRequestName:   pr.PullRequestName,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status.String(),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}
