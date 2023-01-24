package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent/userrole"
)

func (r *Repository) IsProUser(userID uuid.UUID) (bool, error) {
	roles, err := r.DB.UserRole.Query().Where(userrole.UserIDEQ(userID)).All(r.Ctx)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.RoleName == userrole.RoleNameADMIN || role.RoleName == userrole.RoleNamePRO {
			return true, nil
		}
	}
	return false, nil
}
