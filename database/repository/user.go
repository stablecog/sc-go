package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent/subscription"
	"github.com/stablecog/go-apps/database/ent/userrole"
)

// ! This will eventually be deprecated to simply deduct credits
func (r *Repository) IsProUser(userID uuid.UUID) (bool, error) {
	sub, err := r.DB.Subscription.Query().Where(subscription.UserIDEQ(userID)).WithSubscriptionTier().First(r.Ctx)
	if err != nil {
		return false, err
	}

	isPro := sub.Edges.SubscriptionTier.Name == "pro"
	if !isPro {
		// Check for admin
		roles, err := r.DB.UserRole.Query().Where(userrole.UserIDEQ(userID)).All(r.Ctx)
		if err != nil {
			return false, err
		}
		for _, role := range roles {
			if role.RoleName == userrole.RoleNameADMIN {
				return true, nil
			}
		}
	}

	return isPro, nil
}
