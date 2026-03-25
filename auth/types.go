package auth

// Tokens holds an access/refresh token pair
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserOnlineInfo describes a user's online status
type UserOnlineInfo struct {
	LastAccess int64 `json:"last_access"`
	IsOnline   bool  `json:"is_online"`
}
