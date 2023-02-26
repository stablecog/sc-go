package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"github.com/stripe/stripe-go/v74"
	"k8s.io/klog/v2"
)

// HTTP Get - user info
func (c *RestAPI) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	userID, email := c.GetUserIDAndEmailIfAuthenticated(w, r)
	if userID == nil || email == "" {
		return
	}

	// Get customer ID for user
	user, err := c.Repo.GetUser(*userID)
	if err != nil {
		klog.Errorf("Error getting user: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	} else if user == nil {
		// Handle create user flow
		freeCreditType, err := c.Repo.GetOrCreateFreeCreditType()
		if err != nil {
			klog.Errorf("Error getting free credit type: %v", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}
		if freeCreditType == nil {
			klog.Errorf("Server misconfiguration: a credit_type with the name 'free' must exist")
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}

		var customer *stripe.Customer
		if err := c.Repo.WithTx(func(tx *ent.Tx) error {
			client := tx.Client()

			customer, err = c.StripeClient.Customers.New(&stripe.CustomerParams{
				Email: stripe.String(email),
				Params: stripe.Params{
					Metadata: map[string]string{
						"supabase_id": (*userID).String(),
					},
				},
			})
			if err != nil {
				klog.Errorf("Error creating stripe customer: %v", err)
				return err
			}

			u, err := c.Repo.CreateUser(*userID, email, customer.ID, client)
			if err != nil {
				klog.Errorf("Error creating user: %v", err)
				return err
			}

			// Add free credits
			_, err = c.Repo.ReplenishFreeCreditsIfEligible(u.ID, time.Now().AddDate(0, 0, 30), client)
			if err != nil {
				klog.Errorf("Error adding free credits: %v", err)
				return err
			}

			return nil
		}); err != nil {
			klog.Errorf("Error creating user: %v", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			// Delete stripe customer
			if customer != nil {
				_, err := c.StripeClient.Customers.Del(customer.ID, nil)
				if err != nil {
					klog.Errorf("Error deleting stripe customer: %v", err)
				}
			}
			return
		}
	} else {
		_, err := c.Repo.ReplenishFreeCreditsIfEligible(*userID, time.Now().AddDate(0, 0, 30), nil)
		if err != nil {
			klog.Errorf("Error replenishing free credits: %v", err)
		}
	}

	// Get total credits
	totalRemaining, err := c.Repo.GetNonExpiredCreditTotalForUser(*userID, nil)
	if err != nil {
		klog.Errorf("Error getting credits for user: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	customer, err := c.StripeClient.Customers.Get(user.StripeCustomerID, &stripe.CustomerParams{
		Params: stripe.Params{
			Expand: []*string{
				stripe.String("subscriptions"),
			},
		},
	})
	stripeHadError := false
	if err != nil {
		klog.Errorf("Error getting customer from stripe, unknown error: %v", err)
		stripeHadError = true
	}

	// Get current time in ms since epoch
	now := time.Now().UnixNano() / int64(time.Second)
	var highestProduct string
	var cancelsAt *time.Time
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
				highestProduct = item.Price.Product.ID
				// If not scheduled to be cancelled, we are done
				if subscription.CancelAt == 0 {
					cancelsAt = nil
					break
				}
				cancelsAsTime := utils.SecondsSinceEpochToTime(subscription.CancelAt)
				cancelsAt = &cancelsAsTime
			}
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.GetUserResponse{
		TotalRemainingCredits: totalRemaining,
		ProductID:             highestProduct,
		CancelsAt:             cancelsAt,
		StripeHadError:        stripeHadError,
	})
}

// HTTP Get - generations for user
// Takes query paramers for pagination
// per_page: number of generations to return
// cursor: cursor for pagination, it is an iso time string in UTC
func (c *RestAPI) HandleQueryGenerations(w http.ResponseWriter, r *http.Request) {
	userID, _ := c.GetUserIDAndEmailIfAuthenticated(w, r)
	if userID == nil {
		return
	}

	// Validate query parameters
	perPage := DEFAULT_PER_PAGE
	var err error
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		perPage, err = strconv.Atoi(perPageStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "per_page must be an integer")
			return
		} else if perPage < 1 || perPage > MAX_PER_PAGE {
			responses.ErrBadRequest(w, r, fmt.Sprintf("per_page must be between 1 and %d", MAX_PER_PAGE))
			return
		}
	}

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string")
			return
		}
		cursor = &cursorTime
	}

	filters := &requests.QueryGenerationFilters{}
	err = filters.ParseURLQueryParameters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error())
		return
	}

	// Ensure user ID is set to only include this users generations
	filters.UserID = userID

	// Get generaions
	generations, err := c.Repo.QueryGenerations(perPage, cursor, filters)
	if err != nil {
		klog.Errorf("Error getting generations for user: %s", err)
		responses.ErrInternalServerError(w, r, "Error getting generations")
		return
	}

	// Return generations
	render.Status(r, http.StatusOK)
	render.JSON(w, r, generations)
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
		klog.Errorf("Error getting credits for user: %s", err)
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

// HTTP DELETE - admin delete generation
func (c *RestAPI) HandleDeleteGenerationOutputForUser(w http.ResponseWriter, r *http.Request) {
	// Get user id (of admin)
	userID, _ := c.GetUserIDAndEmailIfAuthenticated(w, r)
	if userID == nil {
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

	count, err := c.Repo.MarkGenerationOutputsForDeletionForUser(deleteReq.GenerationOutputIDs, *userID)
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
