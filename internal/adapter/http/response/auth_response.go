package response

// AuthResponse represents the response for login/register
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// RefreshTokenResponse represents the response for refresh token
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}
