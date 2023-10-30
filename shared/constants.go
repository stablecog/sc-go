package shared

import "time"

// Max queued items allowed
const MAX_QUEUED_ITEMS_ULTIMATE = 4
const MAX_QUEUED_ITEMS_PRO = 3
const MAX_QUEUED_ITEMS_STARTER = 2
const MAX_QUEUED_ITEMS_FREE = 1
const MAX_QUEUED_ITEMS_VOICEOVER = 1

// NSFW Error
const NSFW_ERROR = "NSFW"

// Timeout
const TIMEOUT_ERROR = "TIMEOUT"

// After this period, a request will timeout and a user will be refunded
// But the generation/upscale may still go through, if it takes longer than this
const REQUEST_COG_TIMEOUT = 180 * time.Second

const REQUEST_COG_TIMEOUT_VOICEOVER = 180 * time.Second

// Generation related
const MAX_GENERATE_WIDTH = 1536
const MAX_GENERATE_HEIGHT = 1536
const MIN_GENERATE_WIDTH = 256
const MIN_GENERATE_HEIGHT = 256
const MIN_INFERENCE_STEPS = 10
const MAX_GENERATE_INTERFERENCE_STEPS_FREE = 30
const MAX_PRO_PIXEL_STEPS = 1024 * 1024 * 30
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

// The name of the redis stream used to enqueue worker voiceover requests
const COG_REDIS_VOICEOVER_QUEUE = "input_queue_for_voiceover"

// This redis channel our servers publish to when we want to broadcast SSE events to clients
const REDIS_SSE_BROADCAST_CHANNEL = "sse:broadcast_channel"

// This redis channel is when webhook sends an internal request we care about
const REDIS_INTERNAL_COG_CHANNEL = "cog:internal_message"

// This redis channel is when webhook sends a request from API token
const REDIS_APITOKEN_COG_CHANNEL = "cog:apitoken_message"

// This redis channel is for discord bot, when a user has connected their account
const REDIS_DISCORD_COG_CHANNEL = "cog:discord_message"

// Allowed image extensions used by various APIs
type ImageExtension string

const (
	PNG  ImageExtension = "png"
	JPG  ImageExtension = "jpg"
	JPEG ImageExtension = "jpeg"
	WEBP ImageExtension = "webp"
)

// Allowed image extensions for upload
var ALLOWS_IMAGE_EXTENSIONS_UPLOAD = []ImageExtension{WEBP, JPEG, PNG}

// Allowed process type
type ProcessType string

const (
	GENERATE             ProcessType = "generate"
	UPSCALE              ProcessType = "upscale"
	GENERATE_AND_UPSCALE ProcessType = "generate_and_upscale"
	VOICEOVER            ProcessType = "voiceover"
)

// Default image extension for generate
const DEFAULT_PROCESS_TYPE = GENERATE

// Allowed image extensions for upload
var ALLOWED_PROCESS_TYPES = []ProcessType{GENERATE, UPSCALE, GENERATE_AND_UPSCALE}

// Maximum size of a custom image sent to upscale
// 10MB
const MAX_UPSCALE_IMAGE_SIZE = 1024 * 1024 * 10

// Maximum size of a custom image sent to generate for img2img
const MAX_GENERATE_IMAGE_SIZE = 1024 * 1024 * 5

// For custom images
const MAX_UPSCALE_MEGAPIXELS = 1024 * 1024

// ! TODO - Does nothing
const DEFAULT_UPSCALE_SCALE = int32(4)
const DEFAULT_UPSCALE_OUTPUT_EXTENSION = JPEG
const DEFAULT_UPSCALE_OUTPUT_QUALITY = 85

// Free credit replenishments
// How much to give per day
const FREE_CREDIT_AMOUNT_DAILY = 10

// How often to replenish (related to updated_at on credits)
// They get up to FREE_CREDIT_AMOUNT_DAILY in this time period
const FREE_CREDIT_REPLENISHMENT_INTERVAL = 12 * time.Hour

// Last sign in within 7 days
// const FREE_CREDIT_LAST_ACTIVITY_REQUIREMENT = 168 * time.Hour

// ! Auto-upscale
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

// ! Image Generation Defaults
const DEFAULT_GENERATE_OUTPUT_EXTENSION = JPEG
const DEFAULT_GENERATE_OUTPUT_QUALITY = 85
const DEFAULT_GENERATE_NUM_OUTPUTS int32 = 2
const DEFAULT_GENERATE_GUIDANCE_SCALE float32 = 7.0
const DEFAULT_GENERATE_INFERENCE_STEPS int32 = 30
const DEFAULT_GENERATE_PROMPT_STRENGTH float32 = 0.5

// ! Voiceover
// Calculated as math.Ceil(VOICEOVER_CREDIT_COST_PER_CHARACTER * len(text))
const VOICEOVER_CREDIT_COST_PER_CHARACTER = 0.0175

const VOICEOVER_MAX_TEXT_LENGTH = 500

const DEFAULT_VOICEOVER_TEMPERATURE float32 = 0.7
const DEFAULT_VOICEOVER_DENOISE_AUDIO = true
const DEFAULT_VOICEOVER_REMOVE_SILENCE = true

// Discord Bot/Connection
const DISCORD_VERIFY_TOKEN_EXPIRY = 5 * time.Minute

// Tippable credits
// Give credits * TIPPABLE_CREDIT_MULTIPLIER as "tippable" type when replenishing subscription/adhoc purchases
const TIPPABLE_CREDIT_MULTIPLIER = 0.2

// Max upload size allowed for img2img/upscale
const MAX_UPLOAD_SIZE_MB = 10

// Queue priorities
const (
	QUEUE_PRIORITY_1 uint8 = iota + 1
	QUEUE_PRIORITY_2
	QUEUE_PRIORITY_3
	QUEUE_PRIORITY_4
	QUEUE_PRIORITY_5
	QUEUE_PRIORITY_6
	QUEUE_PRIORITY_7
	QUEUE_PRIORITY_8
	QUEUE_PRIORITY_9
	QUEUE_PRIORITY_10
)
