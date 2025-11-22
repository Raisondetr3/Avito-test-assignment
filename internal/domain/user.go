package domain

import "time"

type User struct {
	UserID    string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(userID, username, teamName string, isActive bool) *User {
	now := time.Now()
	return &User{
		UserID:    userID,
		Username:  username,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) CanBeAssignedAsReviewer() bool {
	return u.IsActive
}

func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}
