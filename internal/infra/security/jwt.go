package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/misalima/edunex-backend/internal/core/util"
)

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token expirado")
)

type JWTService struct {
	secretKey string
	issuer    string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secretKey: secret,
		issuer:    "edunex-api",
	}
}

func (s *JWTService) GenerateToken(userID, role string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    s.issuer,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  claims.Subject,
		"iss":  claims.Issuer,
		"exp":  claims.ExpiresAt.Unix(),
		"iat":  claims.IssuedAt.Unix(),
		"role": role,
	})

	return token.SignedString([]byte(s.secretKey))
}

func (s *JWTService) ValidateToken(tokenString string) (*util.TokenClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	issuer, ok := claims["iss"].(string)
	if !ok || issuer != s.issuer {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return nil, ErrInvalidToken
	}

	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return nil, ErrInvalidToken
	}

	return &util.TokenClaims{
		UserID: userID,
		Role:   role,
	}, nil
}
