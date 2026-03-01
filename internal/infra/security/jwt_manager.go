package security

import "github.com/misalima/edunex-backend/internal/core/util"

type JWTManager interface {
	ValidateToken(token string) (*util.TokenClaims, error)
	ValidateTokenViaAPI(token string) (*util.TokenClaims, error)
}
