package security

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/gommon/log"
	"github.com/misalima/edunex-backend/internal/core/util"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken = errors.New("token inválido")
	ErrExpiredToken = errors.New("token expirado")
)

type JWTService struct {
	secretKey   string
	issuer      string
	supabaseURL string
	anonKey     string
}

func NewJWTService(secret, issuer, supabaseURL, anonKey string) *JWTService {
	return &JWTService{
		secretKey:   secret,
		issuer:      issuer,
		supabaseURL: supabaseURL,
		anonKey:     anonKey,
	}
}

func (s *JWTService) ValidateToken(tokenString string) (*util.TokenClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Warn("invalid signing method")
			return nil, ErrInvalidToken
		}
		return []byte(s.secretKey), nil
	})

	if err != nil || !token.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Warn("token expired")
			return nil, ErrExpiredToken
		}
		log.Warn("invalid token")
		return nil, ErrInvalidToken
	}

	iss, ok := claims["iss"].(string)
	if !ok || iss != s.issuer {
		log.Warn("invalid issuer")
		return nil, ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		log.Warn("invalid user id")
		return nil, ErrInvalidToken
	}

	email, _ := claims["email"].(string)

	return &util.TokenClaims{
		UserID: userID,
		Email:  email,
	}, nil
}

func (s *JWTService) ValidateTokenViaAPI(tokenString string) (*util.TokenClaims, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	log.Debugf("supabase url: %s", s.supabaseURL)
	req, _ := http.NewRequest("GET", s.supabaseURL+"/auth/v1/user", nil)

	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("apikey", s.anonKey)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil, errors.New("token inválido ou usuário revogado")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("failed to close response body", zap.Error(err))
		}
	}(resp.Body)

	var user struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		log.Error("failed to decode response body", zap.Error(err))
		return nil, err
	}

	return &util.TokenClaims{UserID: user.ID, Email: user.Email}, nil
}
