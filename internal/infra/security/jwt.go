package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/gommon/log"
	"github.com/misalima/edunex-backend/internal/core/util"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken  = errors.New("token inválido")
	ErrExpiredToken  = errors.New("token expirado")
	ErrInvalidClaims = errors.New("claims inválidos")
)

type JWTService struct {
	publicKey   *ecdsa.PublicKey
	issuer      string
	supabaseURL string
	anonKey     string
}

func NewJWTService(xBase64, yBase64, issuer, supabaseURL, anonKey string) (*JWTService, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(xBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding X: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding Y: %w", err)
	}

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	return &JWTService{
		publicKey:   pubKey,
		issuer:      issuer,
		supabaseURL: supabaseURL,
		anonKey:     anonKey,
	}, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*util.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("expected method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		iss, ok := claims["iss"].(string)
		if !ok || iss != s.issuer {
			return nil, ErrInvalidToken
		}
		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return nil, ErrInvalidClaims
		}
		email, ok := claims["email"].(string)
		if !ok || email == "" {
			return nil, ErrInvalidClaims
		}

		return &util.TokenClaims{
			UserID: userID,
			Email:  email,
		}, nil
	}

	return nil, ErrInvalidClaims
}

func (s *JWTService) ValidateTokenViaAPI(tokenString string) (*util.TokenClaims, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, _ := http.NewRequest("GET", s.supabaseURL+"/auth/v1/user", nil)

	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("apikey", s.anonKey)

	resp, err := client.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("failed to close response body", zap.Error(err))
		}
	}(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		return nil, errors.New("token inválido ou usuário revogado")
	}

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
