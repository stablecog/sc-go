package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stablecog/sc-go/log"
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
			Select(sql.As(rt.C(userrole.FieldRoleName), "role_name"), sql.As(s.C(user.FieldID), "user_id"), sql.As(s.C(user.FieldStripeCustomerID), "stripe_customer_id"), sql.As(s.C(user.FieldActiveProductID), "active_product_id"))
	}).Scan(r.Ctx, &userWithRoles)
	if err != nil {
		log.Error("Error getting user with roles", "err", err)
		return nil, err
	}

	if len(userWithRoles) == 0 {
		return nil, nil
	}

	ret := UserWithRoles{ID: userWithRoles[0].ID, StripeCustomerID: userWithRoles[0].StripeCustomerID, ActiveProductID: userWithRoles[0].ActiveProductID}
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
	ActiveProductID  string    `sql:"active_product_id"`
}

type UserWithRoles struct {
	ID               uuid.UUID
	Roles            []userrole.RoleName
	StripeCustomerID string
	ActiveProductID  string
}

func (r *Repository) GetUserByStripeCustomerId(customerId string) (*ent.User, error) {
	user, err := r.DB.User.Query().Where(user.StripeCustomerIDEQ(customerId)).Only(r.Ctx)
	if err != nil && ent.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		log.Error("Error getting user by stripe customer ID", "err", err)
		return nil, err
	}
	return user, nil
}

func (r *Repository) IsSuperAdmin(userID uuid.UUID) (bool, error) {
	// Check for admin
	roles, err := r.GetRoles(userID)
	if err != nil {
		log.Error("Error getting user roles", "err", err)
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
		log.Error("Error getting user roles", "err", err)
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
		log.Error("Error getting user roles", "err", err)
		return nil, err
	}
	var roleNames []userrole.RoleName
	for _, role := range roles {
		roleNames = append(roleNames, role.RoleName)
	}

	return roleNames, nil
}

// Get count for QueryUsers
func (r *Repository) QueryUsersCount(emailSearch string) (totalCount int, totalCountByProduct map[string]int, err error) {
	query := r.DB.User.Query()
	if emailSearch != "" {
		query = query.Where(user.EmailContains(emailSearch))
	}
	count, err := query.Count(r.Ctx)
	if err != nil {
		log.Error("Error querying users count", "err", err)
		return 0, nil, err
	}

	// Get map of user product_id / count
	var userCreditCount []UserCreditGroupByType
	q := r.DB.User.Query().Where(user.ActiveProductIDNotNil(), user.ActiveProductIDNEQ(""))
	if emailSearch != "" {
		q = q.Where(user.EmailContains(emailSearch))
	}
	q.
		GroupBy(user.FieldActiveProductID).
		Aggregate(ent.Count()).
		Scan(r.Ctx, &userCreditCount)

	// Make it a map
	userCreditCountMap := make(map[string]int, len(userCreditCount))
	for _, userCredit := range userCreditCount {
		userCreditCountMap[userCredit.ActiveProductID] = userCredit.Count
	}

	return count, userCreditCountMap, nil
}

type UserCreditGroupByType struct {
	ActiveProductID string `json:"active_product_id"`
	Count           int    `json:"count"`
}

// Query all users with filters
// per_page is how many rows to return
// cursor is created_at on users, will return items with created_at less than cursor
func (r *Repository) QueryUsers(
	emailSearch string,
	per_page int,
	cursor *time.Time,
	productIds []string,
	banned *bool,
) (*UserQueryMeta, error) {
	selectFields := []string{
		user.FieldID,
		user.FieldEmail,
		user.FieldActiveProductID,
		user.FieldStripeCustomerID,
		user.FieldCreatedAt,
		user.FieldLastSignInAt,
		user.FieldLastSeenAt,
		user.FieldBannedAt,
		user.FieldDataDeletedAt,
		user.FieldScheduledForDeletionOn,
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

	if productIds != nil && len(productIds) > 0 {
		query = query.Where(user.ActiveProductIDIn(productIds...))
	}

	if banned != nil {
		if *banned {
			query = query.Where(user.BannedAtNotNil())
		} else {
			query = query.Where(user.BannedAtIsNil())
		}
	}

	if emailSearch != "" {
		query = query.Where(user.EmailContains(emailSearch))
	}

	// Include user roles
	query.WithUserRoles()

	res, err := query.All(r.Ctx)
	if err != nil {
		log.Error("Error querying users", "err", err)
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
		Next:  next,
		Users: make([]UserQueryResult, len(res)),
	}
	if cursor == nil {
		total, totalByProduct, err := r.QueryUsersCount(emailSearch)
		if err != nil {
			log.Error("Error querying users count", "err", err)
			return nil, err
		}
		meta.Total = &total
		meta.TotalByProductID = totalByProduct
	}

	for i, user := range res {
		formatted := UserQueryResult{
			ID:                     user.ID,
			Email:                  user.Email,
			StripeCustomerID:       user.StripeCustomerID,
			CreatedAt:              user.CreatedAt,
			StripeProductID:        user.ActiveProductID,
			LastSignInAt:           user.LastSignInAt,
			LastSeenAt:             user.LastSeenAt,
			BannedAt:               user.BannedAt,
			DataDeletedAt:          user.DataDeletedAt,
			ScheduledForDeletionOn: user.ScheduledForDeletionOn,
		}
		for _, role := range user.Edges.UserRoles {
			formatted.Roles = append(formatted.Roles, role.RoleName)
		}

		formatted.Credits = make([]UserQueryCredits, len(user.Edges.Credits))
		for i, credit := range user.Edges.Credits {
			creditType := UserQueryCreditType{ID: credit.Edges.CreditType.ID, Name: credit.Edges.CreditType.Name}
			if credit.Edges.CreditType.StripeProductID != nil {
				creditType.StripeProductId = *credit.Edges.CreditType.StripeProductID
			}
			formatted.Credits[i] = UserQueryCredits{
				RemainingAmount: credit.RemainingAmount,
				ExpiresAt:       credit.ExpiresAt,
				CreditType:      creditType,
				ReplenishedAt:   credit.ReplenishedAt,
			}
		}
		meta.Users[i] = formatted
	}

	return meta, nil
}

// Paginated meta for querying generations
type UserQueryMeta struct {
	Total            *int              `json:"total_count,omitempty"`
	TotalByProductID map[string]int    `json:"total_count_by_product_id,omitempty"`
	Next             *time.Time        `json:"next,omitempty"`
	Users            []UserQueryResult `json:"users"`
}

type UserQueryCreditType struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	StripeProductId string    `json:"stripe_product_id,omitempty"`
}

type UserQueryCredits struct {
	RemainingAmount int32               `json:"remaining_amount"`
	ExpiresAt       time.Time           `json:"expires_at"`
	CreditType      UserQueryCreditType `json:"credit_type"`
	ReplenishedAt   time.Time           `json:"replenished_at,omitempty"`
}

type UserQueryResult struct {
	ID                     uuid.UUID           `json:"id"`
	Email                  string              `json:"email"`
	StripeCustomerID       string              `json:"stripe_customer_id"`
	Roles                  []userrole.RoleName `json:"role,omitempty"`
	CreatedAt              time.Time           `json:"created_at"`
	Credits                []UserQueryCredits  `json:"credits,omitempty"`
	LastSignInAt           *time.Time          `json:"last_sign_in_at,omitempty"`
	LastSeenAt             time.Time           `json:"last_seen_at"`
	BannedAt               *time.Time          `json:"banned_at,omitempty"`
	DataDeletedAt          *time.Time          `json:"data_deleted_at,omitempty"`
	ScheduledForDeletionOn *time.Time          `json:"scheduled_for_deletion_on,omitempty"`
	StripeProductID        *string             `json:"product_id,omitempty"`
}

// For credit replenishment
func (r *Repository) GetUsersThatSignedInSince(since time.Duration) ([]*ent.User, error) {
	// Subtract since from now to get users signed in since then
	return r.DB.User.Query().Where(user.LastSeenAtGT(time.Now().Add(-since)), user.ActiveProductIDIsNil()).All(r.Ctx)
}

// Get N subscribers
func (r *Repository) GetNSubscribers() (int, error) {
	return r.DB.User.Query().Where(user.ActiveProductIDNotNil(), user.ActiveProductIDNEQ("")).Count(r.Ctx)
}

// Get is banned
func (r *Repository) IsBanned(userId uuid.UUID) (bool, error) {
	return r.DB.User.Query().Where(user.IDEQ(userId), user.BannedAtNotNil()).Exist(r.Ctx)
}

// Get banned users to delete
func (r *Repository) GetBannedUsersToDelete() ([]*ent.User, error) {
	return r.DB.User.Query().Where(user.BannedAtNotNil(), user.DataDeletedAtIsNil(), user.ScheduledForDeletionOnLT(time.Now())).All(r.Ctx)
}

// Get non-banned users to delete
func (r *Repository) GetUsersToDelete() ([]*ent.User, error) {
	return r.DB.User.Query().Where(user.DataDeletedAtIsNil(), user.ScheduledForDeletionOnLT(time.Now())).All(r.Ctx)
}
