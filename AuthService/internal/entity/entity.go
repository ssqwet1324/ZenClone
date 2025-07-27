package entity

type LoginUserInfo struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenRequest struct {
	ID           string `json:"id"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterRequest struct {
	Login     string `json:"login"`
	Password  string `json:"password"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}
