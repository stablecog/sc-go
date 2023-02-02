package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent/subscription"
	"github.com/stablecog/go-apps/database/ent/userrole"
	"k8s.io/klog/v2"
)

// ! This will eventually be deprecated to simply deduct credits
func (r *Repository) IsProUser(userID uuid.UUID) (bool, error) {
	subTier, err := r.DB.Debug().Subscription.Query().Where(subscription.UserIDEQ(userID)).QuerySubscriptionTier().First(r.Ctx)
	if err != nil {
		return false, err
	}

	isPro := subTier.Name == "pro"
	if isPro {
		return isPro, nil
	}

	isAdmin, _ := r.IsSuperAdmin(userID)
	return isAdmin, nil
}

func (r *Repository) IsSuperAdmin(userID uuid.UUID) (bool, error) {
	// Check for admin
	roles, err := r.GetRoles(userID)
	if err != nil {
		klog.Errorf("Error getting user roles: %v", err)
		return false, err
	}
	for _, role := range roles {
		if role == userrole.RoleNameSUPER_ADMIN {
			return true, nil
		}
	}

	return false, nil
}

func (r *Repository) GetRoles(userID uuid.UUID) ([]userrole.RoleName, error) {
	roles, err := r.DB.Debug().UserRole.Query().Where(userrole.UserIDEQ(userID)).All(r.Ctx)
	if err != nil {
		klog.Errorf("Error getting user roles: %v", err)
		return nil, err
	}
	var roleNames []userrole.RoleName
	for _, role := range roles {
		roleNames = append(roleNames, role.RoleName)
	}

	return roleNames, nil
}
