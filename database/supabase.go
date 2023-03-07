package database

import (
	"errors"

	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"github.com/supabase-community/gotrue-go"
)

var SupabaseAuthUnauthorized = errors.New("Unauthorized")

type SupabaseAuth struct {
	client gotrue.Client
}

// Returns gotrue client with keys
func NewSupabaseAuth() *SupabaseAuth {
	client := gotrue.New(utils.GetEnv("PUBLIC_SUPABASE_REFERENCE_ID", ""), utils.GetEnv("SUPABASE_ADMIN_KEY", ""))
	return &SupabaseAuth{client: client}
}

func (s *SupabaseAuth) GetSupabaseUserIdFromAccessToken(accessToken string) (id, email string, err error) {
	if accessToken == "" {
		return "", "", SupabaseAuthUnauthorized
	}

	user, err := s.client.WithToken(accessToken).GetUser()
	if err != nil {
		log.Error("Error getting user from Supabase", "err", err)
		return "", "", err
	}

	if user == nil {
		log.Info("User not found in Supabase (unauthorized)")
		return "", "", SupabaseAuthUnauthorized
	}

	if user.EmailConfirmedAt == nil {
		log.Info("User not confirmed in Supabase (unauthorized))")
		return "", "", SupabaseAuthUnauthorized
	}

	// Check disposable email
	if shared.GetCache().IsDisposableEmail(user.Email) {
		log.Info("User is using disposable email (unauthorized)", "email", user.Email)
		return "", "", SupabaseAuthUnauthorized
	}

	return user.ID.String(), user.Email, nil
}
