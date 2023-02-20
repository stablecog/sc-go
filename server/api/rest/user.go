package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
	"github.com/stripe/stripe-go"
	"k8s.io/klog/v2"
)

const DEFAULT_PER_PAGE = 50
const MAX_PER_PAGE = 100

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
		freeCreditType, err := c.Repo.GetFreeCreditType()
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
			_, err = c.Repo.AddCreditsIfEligible(freeCreditType, u.ID, time.Now().AddDate(0, 0, 30), client)
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
	credits, err := c.Repo.GetCreditsForUser(*userID)
	if err != nil {
		klog.Errorf("Error getting credits for user: %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	var totalRemaining int32
	for _, credit := range credits {
		totalRemaining += credit.RemainingAmount
	}

	customer, err := c.StripeClient.Customers.Get(user.StripeCustomerID, nil)
	customerNotFound := false
	stripeHadError := false
	if err != nil {
		// Try to decode error as map
		var stripeErr map[string]interface{}
		marshalErr := json.Unmarshal([]byte(err.Error()), &stripeErr)
		if marshalErr == nil {
			status, ok := stripeErr["status"]
			if ok && status == 404 {
				customerNotFound = true
			} else {
				klog.Errorf("Error getting customer from stripe: %v", err)
				stripeHadError = true
			}
		} else {
			klog.Errorf("Error getting customer from stripe, unknown error: %v", err)
			stripeHadError = true
		}
	} else if customer == nil || customer.Subscriptions == nil || customer.Subscriptions.Data == nil {
		customerNotFound = true
	}

	// Get current time in ms since epoch
	now := time.Now().UnixNano() / int64(time.Second)
	var highestProduct string
	var cancelsAt *time.Time
	if !customerNotFound && !stripeHadError {
		// Find highest subscription tier
		for _, subscription := range customer.Subscriptions.Data {
			if subscription.Plan == nil || subscription.Plan.Product == nil {
				continue
			}

			// Not expired or cancelled
			if now > subscription.CurrentPeriodEnd || subscription.CanceledAt > subscription.CurrentPeriodEnd {
				continue
			}
			highestProduct = subscription.Plan.Product.ID
			// If not scheduled to be cancelled, we are done
			if subscription.CancelAt == 0 {
				cancelsAt = nil
				break
			}
			cancelsAsTime := utils.SecondsSinceEpochToTime(subscription.CancelAt)
			cancelsAt = &cancelsAsTime
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.GetUserResponse{
		TotalRemainingCredits: totalRemaining,
		Product:               highestProduct,
		CancelsAt:             cancelsAt,
		StripeHadError:        stripeHadError,
		CustomerNotFound:      customerNotFound,
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

	filters, err := ParseQueryGenerationFilters(r.URL.Query())
	if err != nil {
		responses.ErrBadRequest(w, r, err.Error())
		return
	}

	// Get generaions
	generations, err := c.Repo.GetUserGenerations(*userID, perPage, cursor, filters)
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

// Parse filters from query parameters
func ParseQueryGenerationFilters(rawQuery url.Values) (*requests.QueryGenerationFilters, error) {
	filters := &requests.QueryGenerationFilters{}
	for key, value := range rawQuery {
		// model_ids
		if key == "model_ids" {
			if strings.Contains(value[0], ",") {
				for _, modelId := range strings.Split(value[0], ",") {
					parsed, err := uuid.Parse(modelId)
					if err != nil {
						return nil, fmt.Errorf("invalid model id: %s", modelId)
					}
					filters.ModelIDs = append(filters.ModelIDs, parsed)
				}
			} else {
				parsed, err := uuid.Parse(value[0])
				if err != nil {
					return nil, fmt.Errorf("invalid model id: %s", value[0])
				}
				filters.ModelIDs = []uuid.UUID{parsed}
			}
		}
		// scheduler_ids
		if key == "scheduler_ids" {
			if strings.Contains(value[0], ",") {
				for _, schedulerId := range strings.Split(value[0], ",") {
					parsed, err := uuid.Parse(schedulerId)
					if err != nil {
						return nil, fmt.Errorf("invalid scheduler id: %s", schedulerId)
					}
					filters.SchedulerIDs = append(filters.SchedulerIDs, parsed)
				}
			} else {
				parsed, err := uuid.Parse(value[0])
				if err != nil {
					return nil, fmt.Errorf("invalid scheduler id: %s", value[0])
				}
				filters.SchedulerIDs = []uuid.UUID{parsed}
			}
		}
		// Min and max height
		if key == "min_height" {
			minHeight, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid min height: %s", value[0])
			}
			if minHeight > math.MaxInt32 {
				return nil, fmt.Errorf("min height too large: %d", minHeight)
			}
			filters.MinHeight = int32(minHeight)
		}
		if key == "max_height" {
			maxHeight, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid max height: %s", value[0])
			}
			if maxHeight > math.MaxInt32 {
				return nil, fmt.Errorf("max height too large: %d", maxHeight)
			}
			filters.MaxHeight = int32(maxHeight)
		}
		// Min and max width
		if key == "min_width" {
			minWidth, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid min width: %s", value[0])
			}
			if minWidth > math.MaxInt32 {
				return nil, fmt.Errorf("min width too large: %d", minWidth)
			}
			filters.MinWidth = int32(minWidth)
		}
		if key == "max_width" {
			maxWidth, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid max width: %s", value[0])
			}
			if maxWidth > math.MaxInt32 {
				return nil, fmt.Errorf("max width too large: %d", maxWidth)
			}
			filters.MaxWidth = int32(maxWidth)
		}
		// Min and max inference steps
		if key == "min_inference_steps" {
			minInferenceSteps, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid min inference steps: %s", value[0])
			}
			if minInferenceSteps > math.MaxInt32 {
				return nil, fmt.Errorf("min inference steps too large: %d", minInferenceSteps)
			}
			filters.MinInferenceSteps = int32(minInferenceSteps)
		}
		if key == "max_inference_steps" {
			maxInferenceSteps, err := strconv.Atoi(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid max inference steps: %s", value[0])
			}
			if maxInferenceSteps > math.MaxInt32 {
				return nil, fmt.Errorf("max inference steps too large: %d", maxInferenceSteps)
			}
			filters.MaxInferenceSteps = int32(maxInferenceSteps)
		}
		// Min and max guidance scale, the same but float32 not int32
		if key == "min_guidance_scale" {
			minGuidanceScale, err := strconv.ParseFloat(value[0], 32)
			if err != nil {
				return nil, fmt.Errorf("invalid min guidance scale: %s", value[0])
			}
			filters.MinGuidanceScale = float32(minGuidanceScale)
		}
		if key == "max_guidance_scale" {
			maxGuidanceScale, err := strconv.ParseFloat(value[0], 32)
			if err != nil {
				return nil, fmt.Errorf("invalid max guidance scale: %s", value[0])
			}
			filters.MaxGuidanceScale = float32(maxGuidanceScale)
		}
		// Widths
		if key == "widths" {
			if strings.Contains(value[0], ",") {
				for _, width := range strings.Split(value[0], ",") {
					parsed, err := strconv.Atoi(width)
					if err != nil {
						return nil, fmt.Errorf("invalid width: %s", width)
					}
					if parsed > math.MaxInt32 {
						return nil, fmt.Errorf("width too large: %d", parsed)
					}
					filters.Widths = append(filters.Widths, int32(parsed))
				}
			} else {
				parsed, err := strconv.Atoi(value[0])
				if err != nil {
					return nil, fmt.Errorf("invalid width: %s", value[0])
				}
				if parsed > math.MaxInt32 {
					return nil, fmt.Errorf("width too large: %d", parsed)
				}
				filters.Widths = []int32{int32(parsed)}
			}
		}
		// Heights
		if key == "heights" {
			if strings.Contains(value[0], ",") {
				for _, height := range strings.Split(value[0], ",") {
					parsed, err := strconv.Atoi(height)
					if err != nil {
						return nil, fmt.Errorf("invalid height: %s", height)
					}
					if parsed > math.MaxInt32 {
						return nil, fmt.Errorf("height too large: %d", parsed)
					}
					filters.Heights = append(filters.Heights, int32(parsed))
				}
			} else {
				parsed, err := strconv.Atoi(value[0])
				if err != nil {
					return nil, fmt.Errorf("invalid height: %s", value[0])
				}
				if parsed > math.MaxInt32 {
					return nil, fmt.Errorf("height too large: %d", parsed)
				}
				filters.Heights = []int32{int32(parsed)}
			}
		}
		// Inference Steps
		if key == "inference_steps" {
			if strings.Contains(value[0], ",") {
				for _, inferenceStep := range strings.Split(value[0], ",") {
					parsed, err := strconv.Atoi(inferenceStep)
					if err != nil {
						return nil, fmt.Errorf("invalid inference step: %s", inferenceStep)
					}
					if parsed > math.MaxInt32 {
						return nil, fmt.Errorf("inference step too large: %d", parsed)
					}
					filters.InferenceSteps = append(filters.InferenceSteps, int32(parsed))
				}
			} else {
				parsed, err := strconv.Atoi(value[0])
				if err != nil {
					return nil, fmt.Errorf("invalid inference step: %s", value[0])
				}
				if parsed > math.MaxInt32 {
					return nil, fmt.Errorf("inference step too large: %d", parsed)
				}
				filters.InferenceSteps = []int32{int32(parsed)}
			}
		}
		// Guidance Scales
		if key == "guidance_scales" {
			if strings.Contains(value[0], ",") {
				for _, guidanceScale := range strings.Split(value[0], ",") {
					parsed, err := strconv.ParseFloat(guidanceScale, 32)
					if err != nil {
						return nil, fmt.Errorf("invalid guidance scale: %s", guidanceScale)
					}
					filters.GuidanceScales = append(filters.GuidanceScales, float32(parsed))
				}
			} else {
				parsed, err := strconv.ParseFloat(value[0], 32)
				if err != nil {
					return nil, fmt.Errorf("invalid guidance scale: %s", value[0])
				}
				filters.GuidanceScales = []float32{float32(parsed)}
			}
		}

		// Order
		if key == "order" {
			if strings.ToLower(value[0]) == string(requests.SortOrderAscending) {
				filters.Order = requests.SortOrderAscending
			} else if strings.ToLower(value[0]) == string(requests.SortOrderDescending) {
				filters.Order = requests.SortOrderDescending
			} else {
				return nil, fmt.Errorf("invalid order: '%s' expected '%s' or '%s'", value[0], requests.SortOrderAscending, requests.SortOrderDescending)
			}
		}
		// Upscale status
		if key == "upscaled" {
			if strings.ToLower(value[0]) == string(requests.UpscaleStatusAny) {
				filters.UpscaleStatus = requests.UpscaleStatusAny
			} else if strings.ToLower(value[0]) == string(requests.UpscaleStatusNot) {
				filters.UpscaleStatus = requests.UpscaleStatusNot
			} else if strings.ToLower(value[0]) == string(requests.UpscaleStatusOnly) {
				filters.UpscaleStatus = requests.UpscaleStatusOnly
			} else {
				return nil, fmt.Errorf("invalid upscaled: '%s' expected '%s', '%s', or '%s'", value[0], requests.UpscaleStatusAny, requests.UpscaleStatusNot, requests.UpscaleStatusOnly)
			}
		}
		// Start and end date
		if key == "start_dt" {
			startDt, err := utils.ParseIsoTime(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid start_dt: %s", value[0])
			}
			filters.StartDt = &startDt
		}
		if key == "end_dt" {
			endDt, err := utils.ParseIsoTime(value[0])
			if err != nil {
				return nil, fmt.Errorf("invalid end_dt: %s", value[0])
			}
			filters.EndDt = &endDt
		}
	}
	// Descending default
	if filters.Order == "" {
		filters.Order = requests.SortOrderDescending
	}
	// Upscale status any default
	if filters.UpscaleStatus == "" {
		filters.UpscaleStatus = requests.UpscaleStatusAny
	}
	return filters, nil
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
