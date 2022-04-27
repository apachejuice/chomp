package auth

type Session struct {
	Account Account
	Token   string
}
