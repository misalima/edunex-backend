package iservice

type JWTManager interface {
	GenerateToken(userID string) (string, error)
}
