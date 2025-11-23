package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/Raisondetr3/Avito-test-assignment/internal/errors"
	"github.com/Raisondetr3/Avito-test-assignment/internal/transport/http/dto"
)

func WriteError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		WriteJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	statusCode := getStatusCode(appErr.Code)
	WriteJSONError(w, statusCode, string(appErr.Code), appErr.Message)
}

func WriteJSONError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck
}

func getStatusCode(code errors.ErrorCode) int {
	switch code {
	case errors.ErrCodeTeamExists:
		return http.StatusBadRequest
	case errors.ErrCodePRExists:
		return http.StatusConflict
	case errors.ErrCodePRMerged:
		return http.StatusConflict
	case errors.ErrCodeNotAssigned:
		return http.StatusConflict
	case errors.ErrCodeNoCandidate:
		return http.StatusConflict
	case errors.ErrCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data) //nolint:errcheck
}
