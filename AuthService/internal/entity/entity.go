package entity

type UserInfo struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenRequest struct {
	ID           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
}
