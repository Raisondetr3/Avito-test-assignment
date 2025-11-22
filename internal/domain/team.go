package domain

import "time"

type Team struct {
	TeamName  string
	Members   []*User
	CreatedAt time.Time
}

func NewTeam(teamName string, members []*User) *Team {
	return &Team{
		TeamName:  teamName,
		Members:   members,
		CreatedAt: time.Now(),
	}
}
func (t *Team) GetActiveMembers() []*User {
	activeMembers := make([]*User, 0, len(t.Members))
	for _, member := range t.Members {
		if member.IsActive {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}

func (t *Team) GetActiveMembersExcluding(excludeUserID string) []*User {
	activeMembers := make([]*User, 0, len(t.Members))
	for _, member := range t.Members {
		if member.IsActive && member.UserID != excludeUserID {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}

func (t *Team) HasMember(userID string) bool {
	for _, member := range t.Members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}
