package errors

import "fmt"

type ErrorCode string

const (
	ErrCodeTeamExists ErrorCode = "TEAM_EXISTS"

	ErrCodePRExists ErrorCode = "PR_EXISTS"

	ErrCodePRMerged ErrorCode = "PR_MERGED"

	ErrCodeNotAssigned ErrorCode = "NOT_ASSIGNED"

	ErrCodeNoCandidate ErrorCode = "NO_CANDIDATE"

	ErrCodeNotFound ErrorCode = "NOT_FOUND"
)

type AppError struct {
	Code    ErrorCode
	Message string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func ErrTeamExists(teamName string) *AppError {
	return NewAppError(ErrCodeTeamExists, fmt.Sprintf("team '%s' already exists", teamName))
}

func ErrPRExists(prID string) *AppError {
	return NewAppError(ErrCodePRExists, fmt.Sprintf("pull request '%s' already exists", prID))
}

func ErrPRMerged(prID string) *AppError {
	return NewAppError(ErrCodePRMerged, fmt.Sprintf("pull request '%s' is merged and cannot be modified", prID))
}

func ErrNotAssigned(userID, prID string) *AppError {
	return NewAppError(ErrCodeNotAssigned, fmt.Sprintf("user '%s' is not assigned as reviewer to PR '%s'", userID, prID))
}

func ErrNoCandidate(teamName string) *AppError {
	return NewAppError(ErrCodeNoCandidate, fmt.Sprintf("no active replacement candidate available in team '%s'", teamName))
}

func ErrNotFound(resourceType, identifier string) *AppError {
	return NewAppError(ErrCodeNotFound, fmt.Sprintf("%s '%s' not found", resourceType, identifier))
}

func ErrTeamNotFound(teamName string) *AppError {
	return ErrNotFound("team", teamName)
}

func ErrUserNotFound(userID string) *AppError {
	return ErrNotFound("user", userID)
}

func ErrPRNotFound(prID string) *AppError {
	return ErrNotFound("pull request", prID)
}
