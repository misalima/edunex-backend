package security

import "github.com/misalima/edunex-backend/internal/core/util"

type JWTValidator interface {
	ValidateToken(token string) (*util.TokenClaims, error)
	ValidateTokenViaAPI(token string) (*util.TokenClaims, error)
}
