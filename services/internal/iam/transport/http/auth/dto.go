package auth

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

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

type switchContextRequest struct {
	UserRoleID int64 `json:"userRoleId"`
}

type setDefaultRoleRequest struct {
	RoleID int64 `json:"userRoleId"`
}

type introspectRequest struct {
	Token string `json:"token"`
}
