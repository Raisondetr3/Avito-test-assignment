package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Raisondetr3/Avito-test-assignment/internal/domain"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/dto"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/middleware"
)

type UserHandler struct {
	userService UserService
	prService   PRService
}

func NewUserHandler(userService UserService, prService PRService) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
	}
}

func (h *UserHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	var req dto.SetActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := h.userService.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	response := dto.UserResponse{
		User: mapUserToDTO(user),
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		middleware.WriteJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	prs, err := h.prService.GetPRsByReviewer(r.Context(), userID)
	if err != nil {
		middleware.WriteError(w, err)
		return
	}

	prDTOs := make([]*dto.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		prDTOs = append(prDTOs, &dto.PullRequestShort{
			PullRequestID:   pr.PullRequestID,
			PullRequestName: pr.PullRequestName,
			AuthorID:        pr.AuthorID,
			Status:          pr.Status.String(),
		})
	}

	response := dto.GetReviewResponse{
		UserID:       userID,
		PullRequests: prDTOs,
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

func mapUserToDTO(user *domain.User) *dto.User {
	return &dto.User{
		UserID:   user.UserID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}
