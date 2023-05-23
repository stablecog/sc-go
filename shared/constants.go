package shared

import "time"

// Max queued items allowed
const MAX_QUEUED_ITEMS_ULTIMATE = 4
const MAX_QUEUED_ITEMS_PRO = 3
const MAX_QUEUED_ITEMS_STARTER = 2
const MAX_QUEUED_ITEMS_FREE = 1

// NSFW Error
const NSFW_ERROR = "NSFW"

// Timeout
const TIMEOUT_ERROR = "TIMEOUT"

// After this period, a request will timeout and a user will be refunded
// But the generation/upscale may still go through, if it takes longer than this
const REQUEST_COG_TIMEOUT = 60 * time.Second

// Generation related
const MAX_GENERATE_WIDTH = 1280
const MAX_GENERATE_HEIGHT = 1280
const MAX_GENERATE_INTERFERENCE_STEPS_FREE = 30
const MAX_PRO_PIXEL_STEPS = 640 * 640 * 50
const MAX_GENERATE_NUM_OUTPUTS = 4
const MIN_GENERATE_NUM_OUTPUTS = 1
const MAX_GUIDANCE_SCALE = 30.0
const MIN_GUIDANCE_SCALE = 1.0
const MAX_PROMPT_STRENGTH = 1.0
const MIN_PROMPT_STRENGTH = 0.0

// Prompt related
const MAX_PROMPT_LENGTH = 500

// The name of the redis stream used to enqueue worker requests
const COG_REDIS_QUEUE = "input_queue"

// This redis channel our servers publish to when we want to broadcast SSE events to clients
const REDIS_SSE_BROADCAST_CHANNEL = "sse:broadcast_channel"

// This redis channel is when webhook sends an internal request we care about
const REDIS_INTERNAL_COG_CHANNEL = "cog:internal_message"

// This redis channel is when webhook sends a request from API token
const REDIS_APITOKEN_COG_CHANNEL = "cog:apitoken_message"

// Allowed image extensions used by various APIs
type ImageExtension string

const (
	PNG  ImageExtension = "png"
	JPG  ImageExtension = "jpg"
	JPEG ImageExtension = "jpeg"
	WEBP ImageExtension = "webp"
)

// Default image extension for generate
const DEFAULT_GENERATE_OUTPUT_EXTENSION = JPEG
const DEFAULT_GENERATE_NUM_OUTPUTS = 4
const DEFAULT_GENERATE_OUTPUT_QUALITY = 85

// Allowed image extensions for upload
var ALLOWS_IMAGE_EXTENSIONS_UPLOAD = []ImageExtension{WEBP, JPEG, PNG}

// Allowed process type
type ProcessType string

const (
	GENERATE             ProcessType = "generate"
	UPSCALE              ProcessType = "upscale"
	GENERATE_AND_UPSCALE ProcessType = "generate_and_upscale"
)

// Source type for API requests
type OperationSourceType string

const (
	OperationSourceTypeAPI   OperationSourceType = "api"
	OperationSourceTypeWebUI OperationSourceType = "web-ui"
)

// Default image extension for generate
const DEFAULT_PROCESS_TYPE = GENERATE

// Allowed image extensions for upload
var ALLOWED_PROCESS_TYPES = []ProcessType{GENERATE, UPSCALE, GENERATE_AND_UPSCALE}

// Maximum size of a custom image sent to upscale
// 10MB
const MAX_UPSCALE_IMAGE_SIZE = 1024 * 1024 * 10

// ! TODO - Does nothing
const DEFAULT_UPSCALE_SCALE = int32(4)
const DEFAULT_UPSCALE_OUTPUT_EXTENSION = JPEG
const DEFAULT_UPSCALE_OUTPUT_QUALITY = 85

// Free credit replenishments
// How much to give per day
const FREE_CREDIT_AMOUNT_DAILY = 8

// How often to replenish (related to updated_at on credits)
// They get up to FREE_CREDIT_AMOUNT_DAILY in this time period
const FREE_CREDIT_REPLENISHMENT_INTERVAL = 12 * time.Hour

// Last sign in within 7 days
const FREE_CREDIT_LAST_ACTIVITY_REQUIREMENT = 168 * time.Hour

// ! Auto-upscale
// Only trigger upscale if queue length is not greater than this
const AUTO_UPSCALE_QUEUE_SIZE_QSIZE_REQUIRED = 1

// Evaluate time in queue over this period
const AUTO_UPSCALE_AVG_TIME_IN_QUEUE_SINCE = 5 * time.Minute

// Only trigger upscale if avg time in queue is less than this
const AUTO_UPSCALE_AVG_TIME_IN_QUEUE_REQUIRED = 0.8

// If criteria not met for auto-upscale, retry after this period
const AUTO_UPSCALE_RETRY_DURATION = 30 * time.Second

// ! Deleting user data
const DELETE_BANNED_USER_DATA_AFTER = 24 * time.Hour

// ! API Tokens
// Maximum number of tokens a user can have at any given time
const MAX_API_TOKENS_PER_USER = 10

// Default name for API tokens
const DEFAULT_API_TOKEN_NAME = "Secret key"

// Prefix on all API tokens
const API_TOKEN_PREFIX = "sc-"

// Max chars in an API token name
const MAX_TOKEN_NAME_SIZE = 50

// ! API Queue Overflow
// Max items in overflow queue
const QUEUE_OVERFLOW_MAX = 50

// Penalty for queue overflow
// Computed as time.Sleep(QUEUE_OVERFLOW_PENALTY * QUEUE_OVERFLOW_SIZE)
const QUEUE_OVERFLOW_PENALTY_MS = 150
