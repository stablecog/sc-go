package rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"github.com/stripe/stripe-go/v74"
)

// HTTP Get - user info
func (c *RestAPI) HandleGetUserV2(w http.ResponseWriter, r *http.Request) {
	s := time.Now()

	userID, email := c.GetUserIDAndEmailIfAuthenticated(w, r)

	if userID == nil || email == "" {
		return
	}

	var lastSignIn *time.Time
	lastSignInStr, ok := r.Context().Value("user_last_sign_in").(string)
	if ok {
		lastSignInP, err := time.Parse(time.RFC3339, lastSignInStr)
		if err == nil {
			lastSignIn = &lastSignInP
		}
	}

	// Get user with roles
	user, err := c.Repo.GetUserWithRoles(*userID)

	if err != nil {
		log.Error("HandleGetUserV2 - Error getting user", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	} else if user == nil {
		// Handle create user flow
		m := time.Now()
		err := createNewUser(email, userID, lastSignIn, c)
		if err != nil {
			log.Error("HandleGetUserV2 - Error creating user", err)
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
		go c.Track.SignUp(*userID, email, utils.GetIPAddress(r), utils.GetClientDeviceInfo(r))
		log.Infof("HandleGetUserV2 - createNewUser: %dms", time.Since(m).Milliseconds())

		user, err = c.Repo.GetUserWithRoles(*userID)
		if err != nil {
			log.Error("HandleGetUserV2 - Error getting user with roles", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}
	}

	// Update last_seen_at in a separate GO routine, it's not critical to the result of this function
	go func() {
		err := c.Repo.UpdateLastSeenAt(*userID)
		if err != nil {
			log.Warn("HandleGetUserV2 - Error updating last seen at", "err", err, "user", userID.String())
		}
	}()

	type result struct {
		highestProductID string
		highestPriceID   string
		cancelsAt        *time.Time
		renewsAt         *time.Time
		totalRemaining   int
		paidCreditCount  int
		err              error
		stripeHadError   bool
		hasPurchase      bool
		duration         time.Duration
		operation        string
	}

	goroutineCount := 4
	ch := make(chan result, goroutineCount)
	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	// Get total credits
	go func() {
		defer wg.Done()
		m := time.Now()
		totalRemaining, err := c.Repo.GetNonExpiredCreditTotalForUser(*userID, nil)
		ch <- result{totalRemaining: totalRemaining, err: err, duration: time.Since(m), operation: "GO routine - GetNonExpiredCreditTotalForUser"}
	}()

	// Get customer from Stripe
	go func() {
		defer wg.Done()
		m := time.Now()
		var highestProductID string
		var highestPriceID string
		var cancelsAt *time.Time
		var renewsAt *time.Time
		var stripeHadError bool
		var stripeErr error
		var operation string

		// If the user has info synced to the DB already, use that
		if user.StripeSyncedAt != nil {
			operation = "GO routine - StripeSubscriptionInfoFromUserObject"
			if user.StripeHighestProductID != nil {
				highestProductID = *user.StripeHighestProductID
			}
			if user.StripeHighestPriceID != nil {
				highestPriceID = *user.StripeHighestPriceID
			}
			cancelsAt = user.StripeCancelsAt
			renewsAt = user.StripeRenewsAt
		} else {
			operation = "GO routine - GetAndSyncStripeSubscriptionInfo"
			highestProductID, highestPriceID, cancelsAt, renewsAt, stripeErr = c.GetAndSyncStripeSubscriptionInfo(user.StripeCustomerID)
			stripeHadError = stripeErr != nil
		}

		ch <- result{
			highestProductID: highestProductID,
			highestPriceID:   highestPriceID,
			cancelsAt:        cancelsAt,
			renewsAt:         renewsAt,
			stripeHadError:   stripeHadError,
			duration:         time.Since(m),
			operation:        operation}
	}()

	// Get paid credits
	go func() {
		defer wg.Done()
		m := time.Now()
		paidCreditCount, err := c.Repo.GetNonFreeCreditSum(*userID)
		ch <- result{paidCreditCount: paidCreditCount, err: err, duration: time.Since(m), operation: "GO routine - GetNonFreeCreditSum"}
	}()

	// Get payments made by customer
	go func() {
		defer wg.Done()
		m := time.Now()
		hasPurchase := getHasPurchaseForUser(user.ID, c)
		ch <- result{hasPurchase: hasPurchase, duration: time.Since(m), operation: "GO routine - getHasPurchaseForUser"}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	var res result
	for goroutineResult := range ch {
		if goroutineResult.err != nil {
			log.Error("HandleGetUserV2 - Error in goroutine", "err", goroutineResult.err, "operation", goroutineResult.operation)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}
		if goroutineResult.totalRemaining != 0 {
			res.totalRemaining = goroutineResult.totalRemaining
		}
		if goroutineResult.highestProductID != "" {
			res.highestProductID = goroutineResult.highestProductID
		}
		if goroutineResult.highestPriceID != "" {
			res.highestPriceID = goroutineResult.highestPriceID
		}
		if goroutineResult.cancelsAt != nil {
			res.cancelsAt = goroutineResult.cancelsAt
		}
		if goroutineResult.renewsAt != nil {
			res.renewsAt = goroutineResult.renewsAt
		}
		if goroutineResult.paidCreditCount != 0 {
			res.paidCreditCount = goroutineResult.paidCreditCount
		}
		if goroutineResult.stripeHadError {
			res.stripeHadError = true
		}
		if goroutineResult.hasPurchase {
			res.hasPurchase = goroutineResult.hasPurchase
		}
	}

	moreCreditsAt, moreCreditsAtAmount, renewsAtAmount, freeCreditAmount := getMoreCreditsInfo(*userID, res.highestProductID, res.renewsAt, res.stripeHadError, c)

	roles := make([]string, len(user.Edges.Roles))
	for i, role := range user.Edges.Roles {
		roles[i] = role.Name
	}

	log.Infof("HandleGetUserV2 - Total: %dms", time.Since(s).Milliseconds())

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.GetUserResponseV2{
		UserID:                  userID,
		TotalRemainingCredits:   res.totalRemaining,
		HasNonfreeCredits:       res.paidCreditCount > 0,
		ProductID:               res.highestProductID,
		PriceID:                 res.highestPriceID,
		CancelsAt:               res.cancelsAt,
		RenewsAt:                res.renewsAt,
		RenewsAtAmount:          renewsAtAmount,
		FreeCreditAmount:        freeCreditAmount,
		StripeHadError:          res.stripeHadError,
		Roles:                   roles,
		MoreCreditsAt:           moreCreditsAt,
		MoreCreditsAtAmount:     moreCreditsAtAmount,
		MoreFreeCreditsAt:       moreCreditsAt,
		MoreFreeCreditsAtAmount: moreCreditsAtAmount,
		WantsEmail:              user.WantsEmail,
		Username:                user.Username,
		CreatedAt:               user.CreatedAt,
		UsernameChangedAt:       user.UsernameChangedAt,
		HasPurchase:             res.hasPurchase,
		ScheduledForDeletionOn:  user.ScheduledForDeletionOn,
	})
}

func (c *RestAPI) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	s := time.Now()
	m := time.Now()

	userID, email := c.GetUserIDAndEmailIfAuthenticated(w, r)
	log.Infof("HandleGetUser - GetUserIDAndEmailIfAuthenticated: %dms", time.Since(m).Milliseconds())

	if userID == nil || email == "" {
		return
	}
	var lastSignIn *time.Time
	lastSignInStr, ok := r.Context().Value("user_last_sign_in").(string)
	if ok {
		lastSignInP, err := time.Parse(time.RFC3339, lastSignInStr)
		if err == nil {
			lastSignIn = &lastSignInP
		}
	}

	// Get customer ID for user
	m = time.Now()
	user, err := c.Repo.GetUserWithRoles(*userID)
	log.Infof("HandleGetUser - GetUserWithRoles: %dms", time.Since(m).Milliseconds())

	if err != nil {
		log.Error("Error getting user", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	} else if user == nil {
		// Handle create user flow
		err := createNewUser(email, userID, lastSignIn, c)
		if err != nil {
			log.Error("Error creating user", err)
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
		go c.Track.SignUp(*userID, email, utils.GetIPAddress(r), utils.GetClientDeviceInfo(r))
	}

	if user == nil {
		user, err = c.Repo.GetUserWithRoles(*userID)
		if err != nil {
			log.Error("Error getting user with roles", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}
	}

	// Get total credits
	m = time.Now()
	totalRemaining, err := c.Repo.GetNonExpiredCreditTotalForUser(*userID, nil)
	if err != nil {
		log.Error("Error getting credits for user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}
	log.Infof("HandleGetUser - GetNonExpiredCreditTotalForUser: %dms", time.Since(m).Milliseconds())

	m = time.Now()
	customer, err := c.StripeClient.Customers.Get(user.StripeCustomerID, &stripe.CustomerParams{
		Params: stripe.Params{
			Expand: []*string{
				stripe.String("subscriptions"),
			},
		},
	})
	stripeHadError := false
	if err != nil {
		log.Error("Error getting customer from stripe, unknown error", "err", err)
		stripeHadError = true
	}
	log.Infof("HandleGetUser - GetStripeCustomer: %dms", time.Since(m).Milliseconds())

	// Get subscription info
	highestProductID, highestPriceID, cancelsAt, renewsAt := extractSubscriptionInfoFromCustomer(customer)

	m = time.Now()
	err = c.Repo.UpdateLastSeenAt(*userID)
	if err != nil {
		log.Warn("Error updating last seen at", "err", err, "user", userID.String())
	}
	log.Infof("HandleGetUser - UpdateLastSeenAt: %dms", time.Since(m).Milliseconds())

	// Figure out when free credits will be replenished
	m = time.Now()
	moreCreditsAt, moreCreditsAtAmount, renewsAtAmount, freeCreditAmount := getMoreCreditsInfo(*userID, highestProductID, renewsAt, stripeHadError, c)
	log.Infof("HandleGetUser - getMoreCreditsInfo: %dms", time.Since(m).Milliseconds())

	// Get paid credits for user
	m = time.Now()
	paidCreditCount, err := c.Repo.GetNonFreeCreditSum(*userID)
	if err != nil {
		log.Error("Error getting paid credits for user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}
	log.Infof("HandleGetUser - GetNonFreeCreditSum: %dms", time.Since(m).Milliseconds())

	roles := make([]string, len(user.Edges.Roles))
	for i, role := range user.Edges.Roles {
		roles[i] = role.Name
	}

	m = time.Now()
	purchaseCount := getPurchaseCountForCustomer(user.StripeCustomerID, c)
	log.Infof("HandleGetUser - getPurchaseCountForCustomer: %dms", time.Since(m).Milliseconds())

	log.Infof("HandleGetUser - Total: %dms", time.Since(s).Milliseconds())

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.GetUserResponse{
		UserID:                  userID,
		TotalRemainingCredits:   totalRemaining,
		HasNonfreeCredits:       paidCreditCount > 0,
		ProductID:               highestProductID,
		PriceID:                 highestPriceID,
		CancelsAt:               cancelsAt,
		RenewsAt:                renewsAt,
		RenewsAtAmount:          renewsAtAmount,
		FreeCreditAmount:        freeCreditAmount,
		StripeHadError:          stripeHadError,
		Roles:                   roles,
		MoreCreditsAt:           moreCreditsAt,
		MoreCreditsAtAmount:     moreCreditsAtAmount,
		MoreFreeCreditsAt:       moreCreditsAt,
		MoreFreeCreditsAtAmount: moreCreditsAtAmount,
		WantsEmail:              user.WantsEmail,
		Username:                user.Username,
		CreatedAt:               user.CreatedAt,
		UsernameChangedAt:       user.UsernameChangedAt,
		PurchaseCount:           purchaseCount,
	})
}

func getHasPurchaseForUser(userID uuid.UUID, c *RestAPI) bool {
	hasPurchase := false
	credits, err := c.Repo.GetAllCreditsForUser(userID)
	if err != nil {
		log.Error("Error getting credits for user", err)
		return hasPurchase
	}
	if credits == nil {
		log.Error("No credits found for user", "user", userID.String())
		return hasPurchase
	}
	for _, credit := range credits {
		if credit.Edges.CreditType != nil && credit.Edges.CreditType.StripeProductID != nil {
			hasPurchase = true
			break
		}
	}
	return hasPurchase
}

func getMoreCreditsInfo(userID uuid.UUID, highestProductID string, renewsAt *time.Time, stripeHadError bool, c *RestAPI) (*time.Time, *int, *int, *int) {
	var moreCreditsAt *time.Time
	var moreCreditsAtAmount *int
	var renewsAtAmount *int
	var fcredit *ent.Credit
	var ctype *ent.CreditType
	var freeCreditAmount *int
	var err error
	if highestProductID == "" && !stripeHadError {
		moreCreditsAt, fcredit, ctype, err = c.Repo.GetFreeCreditReplenishesAtForUser(userID)
		if err != nil {
			log.Error("Error getting next free credit replenishment time", "err", err, "user", userID.String())
		}
		moreCreditsAtAmount = utils.ToPtr(shared.FREE_CREDIT_AMOUNT_DAILY)

		if fcredit != nil && ctype != nil {
			if shared.FREE_CREDIT_AMOUNT_DAILY+fcredit.RemainingAmount > ctype.Amount {
				am := int(shared.FREE_CREDIT_AMOUNT_DAILY + fcredit.RemainingAmount - ctype.Amount)
				freeCreditAmount = &am
			} else {
				am := shared.FREE_CREDIT_AMOUNT_DAILY
				freeCreditAmount = &am
			}
		}
	} else if !stripeHadError && renewsAt != nil {
		creditType, err := c.Repo.GetCreditTypeByStripeProductID(highestProductID)
		if err != nil {
			log.Warnf("Error getting credit type from product id '%s' %v", highestProductID, err)
		} else {
			renewsAtAmount = utils.ToPtr(int(creditType.Amount))
		}
	}
	return moreCreditsAt, moreCreditsAtAmount, renewsAtAmount, freeCreditAmount
}

func createNewUser(email string, userID *uuid.UUID, lastSignIn *time.Time, c *RestAPI) error {
	s := time.Now()
	m := time.Now()
	unknownError := errors.New("An unknown error has occurred")
	freeCreditType, err := c.Repo.GetOrCreateFreeCreditType(nil)
	log.Infof("createNewUser - GetOrCreateFreeCreditType: %dms", time.Since(m).Milliseconds())
	if err != nil {
		log.Error("Error getting free credit type", "err", err)
		return unknownError
	}
	if freeCreditType == nil {
		log.Error("Server misconfiguration: a credit_type with the name 'free' must exist")
		return unknownError
	}
	m = time.Now()
	tippableCreditType, err := c.Repo.GetOrCreateTippableCreditType(nil)
	log.Infof("createNewUser - GetOrCreateTippableCreditType: %dms", time.Since(m).Milliseconds())
	if err != nil {
		log.Error("Error getting tippable credit type", "err", err)
		return unknownError
	}
	if tippableCreditType == nil {
		log.Error("Server misconfiguration: a credit_type with the name 'tippable' must exist")
		return unknownError
	}

	// See if email exists
	m = time.Now()
	_, exists, err := c.Repo.CheckIfEmailExistsV2(email)
	if err != nil {
		log.Error("Error checking if email exists", "err", err)
		return unknownError
	} else if exists {
		log.Error("Email already exists", email)
		return errors.New("Email already exists")
	}
	log.Infof("createNewUser - CheckIfEmailExists: %dms", time.Since(m).Milliseconds())

	var customer *stripe.Customer
	if err := c.Repo.WithTx(func(tx *ent.Tx) error {
		client := tx.Client()

		// Create stripe customer
		m = time.Now()
		customer, err = c.StripeClient.Customers.New(&stripe.CustomerParams{
			Email: stripe.String(email),
			Params: stripe.Params{
				Metadata: map[string]string{
					"supabase_id": (*userID).String(),
				},
			},
		})
		if err != nil {
			log.Error("Error creating stripe customer", err)
			return err
		}
		log.Infof("createNewUser - CreateStripeCustomer: %dms", time.Since(m).Milliseconds())

		m = time.Now()
		u, err := c.Repo.CreateUser(*userID, email, customer.ID, lastSignIn, client)
		if err != nil {
			log.Error("Error creating user", err)
			return err
		}
		log.Infof("createNewUser - CreateUser: %dms", time.Since(m).Milliseconds())

		// Add free credits
		m = time.Now()
		added, err := c.Repo.GiveFreeCredits(u.ID, client)
		if err != nil || !added {
			log.Error("Error adding free credits", "err", err)
			return err
		}
		log.Infof("createNewUser - GiveFreeCredits: %dms", time.Since(m).Milliseconds())

		// Add free tippable credits
		m = time.Now()
		added, err = c.Repo.GiveFreeTippableCredits(u.ID, client)
		if err != nil || !added {
			log.Error("Error adding free tippable credits", err)
			return err
		}
		log.Infof("createNewUser - GiveFreeTippableCredits: %dms", time.Since(m).Milliseconds())

		return nil
	}); err != nil {
		log.Error("Error creating user", err)
		// Delete stripe customer
		if customer != nil {
			_, err := c.StripeClient.Customers.Del(customer.ID, nil)
			if err != nil {
				log.Error("Error deleting stripe customer", "err", err)
			}
		}
		return unknownError
	}
	log.Infof("createNewUser - Total: %dms", time.Since(s).Milliseconds())
	return nil
}

func getPurchaseCountForCustomer(customerId string, c *RestAPI) int {
	purchaseCount := 0
	paymentIntents := c.StripeClient.PaymentIntents.List(&stripe.PaymentIntentListParams{
		Customer: stripe.String(customerId),
	})
	for paymentIntents.Next() {
		intent := paymentIntents.PaymentIntent()
		if intent != nil && intent.Status == stripe.PaymentIntentStatusSucceeded {
			purchaseCount++
		}
	}
	return purchaseCount
}

func extractSubscriptionInfoFromCustomer(customer *stripe.Customer) (string, string, *time.Time, *time.Time) {
	now := time.Now().UnixNano() / int64(time.Second)

	var highestProductID string
	var highestPriceID string
	var cancelsAt *time.Time
	var renewsAt *time.Time

	if customer != nil && customer.Subscriptions != nil && customer.Subscriptions.Data != nil {
		// Find highest subscription tier
		for _, subscription := range customer.Subscriptions.Data {
			if subscription.Items == nil || subscription.Items.Data == nil {
				continue
			}

			for _, item := range subscription.Items.Data {
				if item.Price == nil || item.Price.Product == nil {
					continue
				}
				// Not expired or cancelled
				if now > subscription.CurrentPeriodEnd || subscription.CanceledAt > subscription.CurrentPeriodEnd {
					continue
				}
				highestPriceID = item.Price.ID
				highestProductID = item.Price.Product.ID
				// If not scheduled to be cancelled, we are done
				if !subscription.CancelAtPeriodEnd {
					cancelsAt = nil
					break
				}
				cancelsAsTime := utils.SecondsSinceEpochToTime(subscription.CancelAt)
				cancelsAt = &cancelsAsTime
			}
			if cancelsAt == nil && highestProductID != "" {
				renewsAtTime := utils.SecondsSinceEpochToTime(subscription.CurrentPeriodEnd)
				renewsAt = &renewsAtTime
				break
			}
		}
	}
	return highestProductID, highestPriceID, cancelsAt, renewsAt
}

// HTTP Get - generations for user
// Takes query paramers for pagination
// per_page: number of generations to return
// cursor: cursor for pagination, it is an iso time string in UTC
func (c *RestAPI) HandleQueryGenerations(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Validate query parameters
	perPage := DEFAULT_PER_PAGE
	var err error
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "per_page must be an integer", "")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE), "")
			return
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	search := r.URL.Query().Get("search")

	filters := &requests.QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// For search, use qdrant semantic search
	if search != "" {
		// get embeddings from clip service
		e, err := c.Clip.GetEmbeddingFromText(search, true)
		if err != nil {
			log.Error("Error getting embedding from clip service", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}

		// Parse as qdrant filters
		qdrantFilters, scoreThreshold := filters.ToQdrantFilters(false)
		// Append user_id requirement, unless liked
		if filters.IsLiked == nil {
			qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
				Key:   "user_id",
				Match: &qdrant.SCValue{Value: user.ID.String()},
			})
		} else {
			// Get this users likes
			likedIds, err := c.Repo.GetGenerationOutputIDsLikedByUser(user.ID, 10000)
			if err != nil {
				log.Error("Error getting liked ids", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error has occurred")
				return
			}
			qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
				HasId: likedIds,
			})
		}
		// Deleted at not empty
		qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
			IsEmpty: &qdrant.SCIsEmpty{Key: "deleted_at"},
		})

		// Get cursor str as uint
		var offset *uint
		var total *uint
		if cursorStr != "" {
			cursoru64, err := strconv.ParseUint(cursorStr, 10, 64)
			if err != nil {
				responses.ErrBadRequest(w, r, "cursor must be a valid uint", "")
				return
			}
			cursoru := uint(cursoru64)
			offset = &cursoru
		} else {
			count, err := c.Qdrant.CountWithFilters(qdrantFilters, false)
			if err != nil {
				log.Error("Error counting qdrant", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error has occurred")
				return
			}
			total = &count
		}

		// Query qdrant
		qdrantRes, err := c.Qdrant.QueryGenerations(e, perPage, offset, scoreThreshold, filters.Oversampling, qdrantFilters, false, false)
		if err != nil {
			log.Error("Error querying qdrant", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}

		// Get generation output ids
		var outputIds []uuid.UUID
		for _, hit := range qdrantRes.Result {
			outputId, err := uuid.Parse(hit.Id)
			if err != nil {
				log.Error("Error parsing uuid", "err", err)
				continue
			}
			outputIds = append(outputIds, outputId)
		}

		// Get user generation data in correct format
		generationsUnsorted, err := c.Repo.RetrieveGalleryDataWithOutputIDs(outputIds, utils.ToPtr(user.ID), repository.GalleryDataFromHistory)
		if err != nil {
			log.Error("Error getting generations", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}

		gDataMap := make(map[uuid.UUID]repository.GalleryData)
		for _, gData := range generationsUnsorted {
			gDataMap[gData.ID] = gData
		}
		generationsSorted := make([]repository.GalleryData, len(qdrantRes.Result))

		for i, hit := range qdrantRes.Result {
			outputId, err := uuid.Parse(hit.Id)
			if err != nil {
				log.Error("Error parsing uuid", "err", err)
				continue
			}
			item, ok := gDataMap[outputId]
			if !ok {
				log.Error("Error retrieving gallery data", "output_id", outputId)
				continue
			}
			generationsSorted[i] = item
		}

		// Return generations
		render.Status(r, http.StatusOK)
		render.JSON(w, r, GalleryResponseV3[*uint]{
			Next:    qdrantRes.Next,
			Outputs: c.Repo.ConvertRawGalleryDataToV3Results(generationsSorted),
			Total:   total,
		})
		return
	}

	// Otherwise, query postgres
	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string", "")
			return
		}
		cursor = &cursorTime
	}

	// Ensure user ID is set to only include this users generations
	filters.UserID = &user.ID
	filters.ForHistory = true

	// Test flag
	generations, nextCursor, _, err := c.Repo.RetrieveMostRecentGalleryDataV3(filters, filters.UserID, perPage, cursor, nil)
	if err != nil {
		log.Error("Error getting generations for user", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting generations")
		return
	}

	// Presign init image URLs
	signedMap := make(map[string]string)
	for _, g := range generations {
		if g.InitImageURL != nil {
			// See if we have already signed this URL
			signedInitImageUrl, ok := signedMap[*g.InitImageURL]
			if !ok {
				g.InitImageURLSigned = &signedInitImageUrl
				continue
			}
			// remove s3:// prefix
			if strings.HasPrefix(*g.InitImageURL, "s3://") {
				prefixRemoved := (*g.InitImageURL)[5:]
				// Sign object URL to pass to worker
				req, _ := c.S3.GetObjectRequest(&s3.GetObjectInput{
					Bucket: aws.String(utils.GetEnv().S3Img2ImgBucketName),
					Key:    aws.String(prefixRemoved),
				})
				urlStr, err := req.Presign(1 * time.Hour)
				if err != nil {
					log.Error("Error signing init image URL", "err", err)
					continue
				}
				// Add to map
				signedMap[*g.InitImageURL] = urlStr
				g.InitImageURLSigned = &urlStr
			}
		}
	}

	// Get total if no cursor
	var total *uint
	if cursor == nil {
		totalI, err := c.Repo.GetGenerationCount(filters)
		if err != nil {
			log.Error("Error getting user generation count", "err", err)
			responses.ErrInternalServerError(w, r, "Error getting generations")
			return
		}
		// Convert int to uint
		totalUInt := uint(totalI)
		// Assign the address of the uint to the total pointer
		total = &totalUInt
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, GalleryResponseV3[*time.Time]{
		Next:    nextCursor,
		Outputs: c.Repo.ConvertRawGalleryDataToV3Results(generations),
		Total:   total,
	})
}

// HTTP Get - credits for user
func (c *RestAPI) HandleQueryCredits(w http.ResponseWriter, r *http.Request) {
	// See if authenticated
	userIDStr, authenticated := r.Context().Value("user_id").(string)
	// This should always be true because of the auth middleware, but check it anyway
	if !authenticated || userIDStr == "" {
		responses.ErrUnauthorized(w, r)
		return
	}
	// Parse to UUID
	userId, err := uuid.Parse(userIDStr)
	if err != nil {
		responses.ErrUnauthorized(w, r)
		return
	}

	// Get credits
	credits, err := c.Repo.GetCreditsForUser(userId)
	if err != nil {
		log.Error("Error getting credits for user", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting credits")
		return
	}

	// Format as a nicer response
	var totalRemaining int32
	for _, credit := range credits {
		totalRemaining += credit.RemainingAmount
	}

	creditsFormatted := make([]responses.Credit, len(credits))
	for i, credit := range credits {
		creditsFormatted[i] = responses.Credit{
			ID:              credit.ID,
			RemainingAmount: credit.RemainingAmount,
			ExpiresAt:       credit.ExpiresAt,
			Type: responses.CreditType{
				ID:          credit.CreditTypeID,
				Name:        credit.CreditTypeName,
				Description: credit.CreditTypeDescription,
				Amount:      credit.CreditTypeAmount,
			},
		}
	}

	creditsResponse := responses.QueryCreditsResponse{
		TotalRemainingCredits: totalRemaining,
		Credits:               creditsFormatted,
	}

	// Return credits
	render.Status(r, http.StatusOK)
	render.JSON(w, r, creditsResponse)
}

// HTTP DELETE - delete generation
func (c *RestAPI) HandleDeleteGenerationOutputForUser(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	if user.BannedAt != nil {
		responses.ErrForbidden(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var deleteReq requests.DeleteGenerationRequest
	err := json.Unmarshal(reqBody, &deleteReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	count, err := c.Repo.MarkGenerationOutputsForDeletionForUser(deleteReq.GenerationOutputIDs, user.ID)
	if err != nil {
		responses.ErrInternalServerError(w, r, err.Error())
		return
	}

	res := responses.DeletedResponse{
		Deleted: count,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// HTTP POST - favorite generation
func (c *RestAPI) HandleFavoriteGenerationOutputsForUser(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	if user.BannedAt != nil {
		responses.ErrForbidden(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var favReq requests.FavoriteGenerationRequest
	err := json.Unmarshal(reqBody, &favReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if favReq.Action != requests.AddFavoriteAction && favReq.Action != requests.RemoveFavoriteAction {
		responses.ErrBadRequest(w, r, "action must be either 'add' or 'remove'", "")
		return
	}

	count, err := c.Repo.SetFavoriteGenerationOutputsForUser(favReq.GenerationOutputIDs, user.ID, favReq.Action)
	if err != nil {
		responses.ErrInternalServerError(w, r, err.Error())
		return
	}

	res := responses.FavoritedResponse{
		Favorited: count,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// HTTP DELETE - delete voiceover
func (c *RestAPI) HandleDeleteVoiceoverOutputForUser(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	if user.BannedAt != nil {
		responses.ErrForbidden(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var deleteReq requests.DeleteVoiceoverRequest
	err := json.Unmarshal(reqBody, &deleteReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	count, err := c.Repo.MarkVoiceoverOutputsForDeletionForUser(deleteReq.OutputIDs, user.ID)
	if err != nil {
		responses.ErrInternalServerError(w, r, err.Error())
		return
	}

	res := responses.DeletedResponse{
		Deleted: count,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// HTTP POST - set email preferences
func (c *RestAPI) HandleUpdateEmailPreferences(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	if user.BannedAt != nil {
		responses.ErrForbidden(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var emailReq requests.EmailPreferencesRequest
	err := json.Unmarshal(reqBody, &emailReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Update email preferences
	err = c.Repo.SetWantsEmail(user.ID, emailReq.WantsEmail)
	if err != nil {
		log.Error("Error setting email preferences", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	res := responses.UpdatedResponse{
		Updated: 1,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// HTTP POST - set email preferences
func (c *RestAPI) HandleUpdateUsername(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	if user.BannedAt != nil {
		responses.ErrForbidden(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var usernameReq requests.ChangeUsernameRequest
	err := json.Unmarshal(reqBody, &usernameReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Check if valid
	if err := utils.IsValidUsername(usernameReq.Username); err != nil {
		responses.ErrBadRequest(w, r, err.Error(), "")
		return
	}

	// Update username
	err = c.Repo.SetUsername(user.ID, usernameReq.Username)
	if err != nil {
		if errors.Is(err, repository.UsernameExistsErr) {
			responses.ErrBadRequest(w, r, "username_taken", "")
			return
		}
		log.Error("Error setting username", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"username": usernameReq.Username,
	})
}

// HTTP POST - like/unlike generation
func (c *RestAPI) HandleLikeGenerationOutputsForUser(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	if user.BannedAt != nil {
		responses.ErrForbidden(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var likeReq requests.LikeUnlikeActionRequest
	err := json.Unmarshal(reqBody, &likeReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if likeReq.Action != requests.LikeAction && likeReq.Action != requests.UnlikeAction {
		responses.ErrBadRequest(w, r, "action must be either 'like' or 'unlike'", "")
		return
	}

	err = c.Repo.SetOutputsLikedForUser(likeReq.GenerationOutputIDs, user.ID, likeReq.Action)
	// Error check required due to https://github.com/ent/ent/issues/2176
	// This shouldn't return an error if it fails, due to on conflict do nothing behavior
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error("Error setting outputs liked for user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"success": true,
	})
}

// HTTP POST - schedule for deletion
func (c *RestAPI) HandleScheduleUserForDeletion(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	if user.BannedAt != nil {
		responses.ErrForbidden(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var deleteReq requests.DeleteUserRequest
	err := json.Unmarshal(reqBody, &deleteReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	switch deleteReq.Action {
	case requests.DeleteAction:
		if user.ScheduledForDeletionOn != nil {
			render.Status(r, http.StatusOK)
			render.JSON(w, r, map[string]interface{}{
				"success":                   true,
				"scheduled_for_deletion_on": *user.ScheduledForDeletionOn,
			})
			return
		}
		scheduledAt, err := c.Repo.MarkUserForDeletion(user.ID)

		if err != nil {
			log.Error("Error marking user for deletion", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"success":                   true,
			"scheduled_for_deletion_on": scheduledAt,
		})
		return
	case requests.UndeleteAction:
		if user.ScheduledForDeletionOn == nil {
			render.Status(r, http.StatusOK)
			render.JSON(w, r, map[string]interface{}{
				"success": true,
			})
			return
		}
		_, err := c.Repo.UnmarkUserForDeletion(user.ID)
		if err != nil {
			log.Error("Error unmarking user for deletion", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"success": true,
	})
}
