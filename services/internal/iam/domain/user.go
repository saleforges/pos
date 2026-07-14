package domain

import "time"

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type BranchScope string

const (
	BranchScopeAll      BranchScope = "all"
	BranchScopeAssigned BranchScope = "assigned"
	BranchScopeNone     BranchScope = "none"
)

type UserRoleAssignment struct {
	ID           int64
	Role         Role
	MerchantID   int64
	MerchantName string
	BranchID     *int64
	BranchName   string
	IsDefault    bool
	BranchScope  BranchScope
}

type User struct {
	ID         int64                  `json:"id"`
	Username   string                 `json:"username"`
	Email      string                 `json:"email"`
	Password   string                 `json:"-"`
	SystemRole *Role                  `json:"-"`
	Roles      []UserRoleAssignment   `json:"roles"`
	Type       UserType               `json:"type"`
	Status     UserStatus             `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type RefreshToken struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type Session struct {
	ID               string     `json:"id"`
	UserID           int64      `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	ActiveUserRoleID int64      `json:"active_user_role_id"`
	UserAgent        string     `json:"user_agent"`
	IPAddress        string     `json:"ip_address"`
	LastUsedAt       time.Time  `json:"last_used_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type APIKey struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key,omitempty"`
	UserID    int64      `json:"user_id"`
	Roles     []string   `json:"roles"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type LoginAudit struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	Success   bool      `json:"success"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Reason    string    `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
