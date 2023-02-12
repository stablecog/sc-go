package database

import (
	"errors"
	"log"

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

func (s *SupabaseAuth) GetSupabaseUserIdFromAccessToken(accessToken string) (string, error) {
	if accessToken == "" {
		return "", SupabaseAuthUnauthorized
	}

	user, err := s.client.WithToken(accessToken).GetUser()
	if err != nil {
		log.Printf("Error getting user from Supabase: %v", err)
		return "", err
	}

	if user == nil {
		log.Printf("User not found in Supabase")
		return "", SupabaseAuthUnauthorized
	}

	return user.ID.String(), nil
}
