package user

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

type authResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int    `json:"expiresIn"`
	UserID       int64  `json:"userId,omitempty"`
}
