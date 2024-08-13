package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/qdrant"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
	"golang.org/x/exp/slices"
)

type BanDomainRequest struct {
	Action     string   `json:"action"`
	DeleteData bool     `json:"delete_data"`
	Domains    []string `json:"domains"`
}

type BannedResponse struct {
	AffectedUsers int `json:"affected_users"`
}

// Get disposable domains
func (c *RestAPI) HandleGetDisposableDomains(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, BanDomainRequest{
		Domains: shared.GetCache().DisposableEmailDomains(),
	})
}

// Bulk ban email domains
func (c *RestAPI) HandleBanDomains(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var banDomainsReq BanDomainRequest
	err := json.Unmarshal(reqBody, &banDomainsReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if banDomainsReq.Action != "ban" && banDomainsReq.Action != "unban" {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Unsupported action %s", banDomainsReq.Action), "")
		return
	}

	// Exec ban
	if banDomainsReq.Action == "ban" {
		affected, err := c.Repo.BanDomains(banDomainsReq.Domains, banDomainsReq.DeleteData)
		if err != nil {
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, BannedResponse{
			AffectedUsers: affected,
		})
		return
	}

	// Unban
	affected, err := c.Repo.UnbanDomains(banDomainsReq.Domains)
	if err != nil {
		responses.ErrInternalServerError(w, r, err.Error())
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, BannedResponse{
		AffectedUsers: affected,
	})
}

// Admin-related routes, these must be behind admin middleware and auth middleware
// HTTP POST - admin ban user
func (c *RestAPI) HandleBanUser(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var banUsersReq requests.BanUsersRequest
	err := json.Unmarshal(reqBody, &banUsersReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if banUsersReq.Action != requests.BanActionBan && banUsersReq.Action != requests.BanActionUnban {
		responses.ErrBadRequest(w, r, fmt.Sprintf("Unsupported action %s", banUsersReq.Action), "")
		return
	}

	var affected int
	if banUsersReq.Action == requests.BanActionBan {
		affected, err = c.Repo.BanUsers(banUsersReq.UserIDs, banUsersReq.DeleteData)
		if err != nil {
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
	} else {
		affected, err = c.Repo.UnbanUsers(banUsersReq.UserIDs)
		if err != nil {
			responses.ErrInternalServerError(w, r, err.Error())
			return
		}
	}

	res := responses.UpdatedResponse{
		Updated: affected,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// HTTP POST - admin approve/reject image in gallery
func (c *RestAPI) HandleReviewGallerySubmission(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var adminGalleryReq requests.ReviewGalleryRequest
	err := json.Unmarshal(reqBody, &adminGalleryReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	galleryStatus := adminGalleryReq.GalleryStatus

	var updateCount int
	if galleryStatus == "" {
		switch adminGalleryReq.Action {
		case requests.GalleryApproveAction:
			galleryStatus = generationoutput.GalleryStatusApproved
		case requests.GalleryRejectAction:
			galleryStatus = generationoutput.GalleryStatusRejected
		}
	}

	if galleryStatus != generationoutput.GalleryStatusApproved &&
		galleryStatus != generationoutput.GalleryStatusRejected &&
		galleryStatus != generationoutput.GalleryStatusWaitingForApproval &&
		galleryStatus != generationoutput.GalleryStatusSubmitted &&
		galleryStatus != generationoutput.GalleryStatusNotSubmitted {
		responses.ErrBadRequest(w, r, "invalid_gallery_status", galleryStatus.String())
		return
	}

	updateCount, err = c.Repo.BulkUpdateGalleryStatusForOutputs(adminGalleryReq.GenerationOutputIDs, galleryStatus)
	if err != nil {
		if ent.IsNotFound(err) {
			responses.ErrBadRequest(w, r, "Generation not found", "")
			return
		}
		responses.ErrInternalServerError(w, r, err.Error())
		return
	}

	res := responses.UpdatedResponse{
		Updated: updateCount,
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// HTTP DELETE - admin delete generation
func (c *RestAPI) HandleDeleteGenerationOutput(w http.ResponseWriter, r *http.Request) {
	// Get user
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Get user_role from context
	userRole, ok := r.Context().Value("user_role").(string)
	if !ok || userRole != "SUPER_ADMIN" {
		responses.ErrUnauthorized(w, r)
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

	count, err := c.Repo.MarkGenerationOutputsForDeletion(deleteReq.GenerationOutputIDs)
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

func (c *RestAPI) HandleQueryGenerationsForAdmin(w http.ResponseWriter, r *http.Request) {
	user, email := c.GetUserIDAndEmailIfAuthenticated(w, r)
	if user == nil || email == "" {
		return
	}

	// Get user_role from context
	userRole, ok := r.Context().Value("user_role").(string)
	if !ok {
		responses.ErrUnauthorized(w, r)
		return
	}
	superAdmin := userRole == "SUPER_ADMIN"

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

	// Make sure non-super admin can't get private generations
	if !superAdmin {
		if len(filters.GalleryStatus) == 0 {
			filters.GalleryStatus = []generationoutput.GalleryStatus{
				generationoutput.GalleryStatusApproved,
				generationoutput.GalleryStatusRejected,
				generationoutput.GalleryStatusSubmitted,
				generationoutput.GalleryStatusWaitingForApproval,
			}
		} else if slices.Contains(filters.GalleryStatus, generationoutput.GalleryStatusNotSubmitted) {
			responses.ErrUnauthorized(w, r)
			return
		}
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

		if filters != nil && len(filters.Username) > 0 {
			userIDs, err := c.Repo.GetUserIDsByUsernames(filters.Username)
			if err != nil {
				log.Error("Error getting user ids by usernames", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error occurred")
				return
			}
			if len(userIDs) == 0 {
				render.Status(r, http.StatusOK)
				render.JSON(w, r, repository.GenerationQueryWithOutputsMeta[*uint]{})
				return
			}
			shouldFilter := []qdrant.SCMatchCondition{}
			for _, userID := range userIDs {
				shouldFilter = append(shouldFilter, qdrant.SCMatchCondition{
					Key:   "user_id",
					Match: &qdrant.SCValue{Value: userID.String()},
				})
			}
			if len(shouldFilter) > 0 {
				qdrantFilters.Must = append(qdrantFilters.Must, qdrant.SCMatchCondition{
					Should: shouldFilter,
				})
			}
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
		generationsUnsorted, err := c.Repo.RetrieveGalleryDataWithOutputIDs(outputIds, user, repository.GalleryDataFromHistory)
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
		totalI, err := c.Repo.GetGenerationCountAdmin(filters)
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

// HTTP Get - users for admin
func (c *RestAPI) HandleQueryUsers(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
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

	var cursor *time.Time
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		cursorTime, err := utils.ParseIsoTime(cursorStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "cursor must be a valid iso time string", "")
			return
		}
		cursor = &cursorTime
	}

	var productIds []string
	if productIdsStr := r.URL.Query().Get("active_product_ids"); productIdsStr != "" {
		productIds = strings.Split(productIdsStr, ",")
	}

	// Ban status
	var banned *bool
	if banStatusStr := r.URL.Query().Get("banned"); banStatusStr != "" {
		bannedBool, err := strconv.ParseBool(banStatusStr)
		if err != nil {
			responses.ErrBadRequest(w, r, "banned must be a boolean", "")
			return
		}
		banned = &bannedBool
	}

	// Get users
	users, err := c.Repo.QueryUsers(r.URL.Query().Get("search"), perPage, cursor, productIds, banned)
	if err != nil {
		log.Error("Error getting users", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting users")
		return
	}

	// Return generations
	render.Status(r, http.StatusOK)
	render.JSON(w, r, users)
}

// Get available credit types admin can gift to user
func (c *RestAPI) HandleQueryCreditTypes(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Get credit types
	creditTypes, err := c.Repo.GetCreditTypeList()
	if err != nil {
		log.Error("Error getting credit types", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	resp := make([]responses.QueryCreditTypesResponse, len(creditTypes))
	for i, ct := range creditTypes {
		resp[i].ID = ct.ID
		resp[i].Amount = ct.Amount
		resp[i].Name = ct.Name
		resp[i].Description = ct.Name
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// Add credits to user
func (c *RestAPI) HandleAddCreditsToUser(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var addReq requests.CreditAddRequest
	err := json.Unmarshal(reqBody, &addReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	// Get credit type
	creditType, err := c.Repo.GetCreditTypeByID(addReq.CreditTypeID)
	if err != nil {
		log.Error("Error getting credit type", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	} else if err == nil && creditType == nil {
		responses.ErrNotFound(w, r, fmt.Sprintf("Invalid credit type %s", addReq.CreditTypeID.String()))
		return
	}

	err = c.Repo.AddCreditsToUser(creditType, addReq.UserID)
	if err != nil {
		log.Error("Error adding credits to user", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"added": true,
	})
}

// Embed text
type EmbedTextRequest struct {
	Text      string `json:"text"`
	Translate bool   `json:"translate"`
}

type EmbedImagePathRequest struct {
	ImagePath string `json:"image_path"`
	NoCache   bool   `json:"no_cache"`
}

type EmbedTextResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (c *RestAPI) HandleEmbedText(w http.ResponseWriter, r *http.Request) {
	if user, email := c.GetUserIDAndEmailIfAuthenticated(w, r); user == nil || email == "" {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var embedReq EmbedTextRequest
	err := json.Unmarshal(reqBody, &embedReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	embeddings, err := c.Clip.GetEmbeddingFromText(embedReq.Text, embedReq.Translate)
	if err != nil {
		log.Errorf("Error getting embeddings %v", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, EmbedTextResponse{
		Embedding: embeddings,
	})
}
