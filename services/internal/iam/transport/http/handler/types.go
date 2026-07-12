package handler

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

type userResponse struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	CreatedAt string     `json:"createdAt"`
	UpdatedAt string     `json:"updatedAt"`
}

type merchantDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type branchDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type roleResponse struct {
	ID       int64        `json:"id"`
	Name     string       `json:"name"`
	Merchant *merchantDTO `json:"merchant"`
	Branch   *branchDTO   `json:"branch"`
}

type meResponse struct {
	ID             int64          `json:"id"`
	Username       string         `json:"username"`
	Email          string         `json:"email"`
	Type           string         `json:"type"`
	Status         string         `json:"status"`
	Roles          []roleResponse `json:"roles"`
	DefaultBranch  *branchDTO     `json:"defaultBranch"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
}

type createUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type updateUserRequest struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Status   *string `json:"status"`
}

type createRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type updateRoleRequest struct {
	Description *string `json:"description"`
}

type createPermissionRequest struct {
	Permission string `json:"permission"`
}
