package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"k8s.io/klog/v2"
)

func (r *Repository) GetUser(id uuid.UUID) (*ent.User, error) {
	user, err := r.DB.User.Query().Where(user.IDEQ(id)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *Repository) GetUserWithRoles(id uuid.UUID) (*UserWithRoles, error) {
	var userWithRoles []UserWithRolesRaw
	err := r.DB.User.Query().Where(user.IDEQ(id)).Modify(func(s *sql.Selector) {
		rt := sql.Table(userrole.Table)
		s.LeftJoin(rt).On(rt.C(userrole.FieldUserID), s.C(user.FieldID)).
			Select(sql.As(rt.C(userrole.FieldRoleName), "role_name"), sql.As(s.C(user.FieldID), "user_id"), sql.As(s.C(user.FieldStripeCustomerID), "stripe_customer_id"))
	}).Scan(r.Ctx, &userWithRoles)
	if err != nil {
		klog.Errorf("Error getting user with roles: %v", err)
		return nil, err
	}

	if len(userWithRoles) == 0 {
		return nil, nil
	}

	ret := UserWithRoles{ID: userWithRoles[0].ID, StripeCustomerID: userWithRoles[0].StripeCustomerID}
	for _, userWithRole := range userWithRoles {
		if userWithRole.RoleName == "" {
			continue
		}
		ret.Roles = append(ret.Roles, userrole.RoleName(userWithRole.RoleName))
	}
	return &ret, nil
}

type UserWithRolesRaw struct {
	ID               uuid.UUID `sql:"user_id"`
	RoleName         string    `sql:"role_name"`
	StripeCustomerID string    `sql:"stripe_customer_id"`
}

type UserWithRoles struct {
	ID               uuid.UUID
	Roles            []userrole.RoleName
	StripeCustomerID string
}

func (r *Repository) GetUserByStripeCustomerId(customerId string) (*ent.User, error) {
	user, err := r.DB.User.Query().Where(user.StripeCustomerIDEQ(customerId)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		klog.Errorf("Error getting user by stripe customer ID: %v", err)
		return nil, err
	}
	return user, nil
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

func (r *Repository) GetSuperAdminUserIDs() ([]uuid.UUID, error) {
	// Query all super  admins
	admins, err := r.DB.UserRole.Query().Select(userrole.FieldUserID).Where(userrole.RoleNameEQ(userrole.RoleNameSUPER_ADMIN)).All(r.Ctx)
	if err != nil {
		klog.Errorf("Error getting user roles: %v", err)
		return nil, err
	}
	var adminIDs []uuid.UUID
	for _, admin := range admins {
		adminIDs = append(adminIDs, admin.UserID)
	}
	return adminIDs, nil
}

func (r *Repository) GetRoles(userID uuid.UUID) ([]userrole.RoleName, error) {
	roles, err := r.DB.UserRole.Query().Where(userrole.UserIDEQ(userID)).All(r.Ctx)
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

// Get count for QueryUsers
func (r *Repository) QueryUsersCount() (totalCount int, countByCreditName map[string]int, err error) {
	count, err := r.DB.User.Query().Count(r.Ctx)
	if err != nil {
		klog.Errorf("Error querying users count: %v", err)
		return 0, nil, err
	}

	// Get count of users with credits grouped by credit_type
	// ! TODO - it would be better to have SQL do the aggregations for this data
	var rawResult []UserCreditGroupByType
	err = r.DB.Credit.Query().Where(credit.ExpiresAtGT(time.Now())).
		Modify(func(s *sql.Selector) {
			ct := sql.Table(credittype.Table)
			s.Join(ct).On(ct.C(credittype.FieldID), s.C(credit.FieldCreditTypeID)).
				Select(sql.As(ct.C(credittype.FieldName), "credit_name"), sql.As(s.C(credit.FieldUserID), "user_id")).
				GroupBy(s.C(credit.FieldUserID), "credit_name")
		}).Scan(r.Ctx, &rawResult)
	if err != nil {
		klog.Errorf("Error querying users count: %v", err)
		return 0, nil, err
	}

	userCreditCountMap := make(map[string]int)
	for _, result := range rawResult {
		if _, ok := userCreditCountMap[result.CreditName]; !ok {
			userCreditCountMap[result.CreditName] = 1
			continue
		}
		userCreditCountMap[result.CreditName]++
	}

	return count, userCreditCountMap, nil
}

type UserCreditGroupByType struct {
	CreditName string    `json:"credit_name" sql:"credit_name"`
	UserID     uuid.UUID `json:"user_id" sql:"user_id"`
}

// Query all users with filters
// per_page is how many rows to return
// cursor is created_at on users, will return items with created_at less than cursor
func (r *Repository) QueryUsers(per_page int, cursor *time.Time) (*UserQueryMeta, error) {
	selectFields := []string{
		user.FieldID,
		user.FieldEmail,
		user.FieldStripeCustomerID,
		user.FieldCreatedAt,
	}

	var query *ent.UserQuery

	query = r.DB.User.Query().Select(selectFields...).Order(ent.Desc(user.FieldCreatedAt))
	if cursor != nil {
		query = query.Where(user.CreatedAtLT(*cursor))
	}

	query = query.Limit(per_page + 1)

	// Include non-expired credits and type
	query.WithCredits(func(s *ent.CreditQuery) {
		s.Where(credit.ExpiresAtGT(time.Now())).WithCreditType().Order(ent.Asc(credit.FieldExpiresAt))
	})

	// Include user roles
	query.WithUserRoles()

	res, err := query.All(r.Ctx)
	if err != nil {
		klog.Errorf("Error querying users: %v", err)
		return nil, err
	}

	// Check if there is a next page
	var next *time.Time
	if len(res) > per_page {
		next = &res[per_page-1].CreatedAt
		res = res[:per_page]
	}

	// Build meta
	meta := &UserQueryMeta{
		Next: next,
	}
	if cursor == nil {
		total, totalByType, err := r.QueryUsersCount()
		if err != nil {
			klog.Errorf("Error querying users count: %v", err)
			return nil, err
		}
		meta.Total = &total
		meta.TotalByCreditName = totalByType
	}

	for _, user := range res {
		formatted := UserQueryResult{
			ID:               user.ID,
			Email:            user.Email,
			StripeCustomerID: user.StripeCustomerID,
			CreatedAt:        user.CreatedAt,
		}
		for _, role := range user.Edges.UserRoles {
			formatted.Roles = append(formatted.Roles, role.RoleName)
		}
		for _, credit := range user.Edges.Credits {
			creditType := UserQueryCreditType{Name: credit.Edges.CreditType.Name}
			if credit.Edges.CreditType.StripeProductID != nil {
				creditType.StripeProductId = *credit.Edges.CreditType.StripeProductID
			}
			formatted.Credits = append(formatted.Credits, UserQueryCredits{
				RemainingAmount: credit.RemainingAmount,
				ExpiresAt:       credit.ExpiresAt,
				CreditType:      creditType,
			})
		}
		meta.Users = append(meta.Users, formatted)
	}

	return meta, nil
}

// Paginated meta for querying generations
type UserQueryMeta struct {
	Total             *int              `json:"total_count,omitempty"`
	TotalByCreditName map[string]int    `json:"total_count_by_name,omitempty"`
	Next              *time.Time        `json:"next,omitempty"`
	Users             []UserQueryResult `json:"users"`
}

type UserQueryCreditType struct {
	Name            string `json:"name"`
	StripeProductId string `json:"stripe_product_id,omitempty"`
}

type UserQueryCredits struct {
	RemainingAmount int32 `json:"remaining_amount"`
	ExpiresAt       time.Time
	CreditType      UserQueryCreditType `json:"credit_type"`
}

type UserQueryResult struct {
	ID               uuid.UUID           `json:"id"`
	Email            string              `json:"email"`
	StripeCustomerID string              `json:"stripe_customer_id"`
	Roles            []userrole.RoleName `json:"role,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	Credits          []UserQueryCredits
}
