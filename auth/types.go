package auth

// Tokens holds an access/refresh token pair
type Tokens struct {
	AccessToken  string
	RefreshToken string
}

// UserOnlineInfo describes a user's online status
type UserOnlineInfo struct {
	LastAccess int64 `json:"last_access"`
	IsOnline   bool  `json:"is_online"`
}
