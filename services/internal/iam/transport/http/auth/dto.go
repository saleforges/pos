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
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int    `json:"expiresIn"`
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
