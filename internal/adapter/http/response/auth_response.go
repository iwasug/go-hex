package response

// AuthResponse represents the response for login
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse represents the response for refresh token
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}
