package iservice

type JWTManager interface {
	GenerateToken(userID, role string) (string, error)
}
