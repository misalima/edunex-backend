package primary

import (
	"context"

	"github.com/misalima/edunex-backend/internal/core/util"
)

// Authenticator defines the primary port interface for authenticating incoming requests
// and ensuring the user is synchronized with the local database.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*util.TokenClaims, error)
}
