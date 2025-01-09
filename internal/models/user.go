package models

type AuthInfo struct {
	IsAuthenticated bool
	Token           string
}

type User struct {
	ID          int    `json:"id"`
	Login       string `json:"login"`
	Password    string `json:"omitempty"`
	AuthInfo    *AuthInfo
	BalanceInfo *Balance
}

type UserLogPassRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
