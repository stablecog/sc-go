package shared

// Generation related
const MAX_GENERATE_WIDTH = 768
const MAX_GENERATE_HEIGHT = 768
const MAX_GENERATE_WIDTH_FREE = 512
const MAX_GENERATE_HEIGHT_FREE = 512
const MAX_GENERATE_INTERFERENCE_STEPS_FREE = 30
const MAX_PRO_PIXEL_STEPS = 640 * 640 * 50
const MAX_GENERATE_NUM_OUTPUTS = 4
const MIN_GENERATE_NUM_OUTPUTS = 1

// Prompt related
const MAX_PROMPT_LENGTH = 500

// The name of the redis stream used to enqueue cog requests
const COG_REDIS_QUEUE = "input_queue"

// This is the redis channel that the cog publishes to for generate events
const COG_REDIS_GENERATE_EVENT_CHANNEL = "queue:generate_event"

// Same for upscale
const COG_REDIS_UPSCALE_EVENT_CHANNEL = "queue:upscale_event"

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

// Default image extension for generate
const DEFAULT_PROCESS_TYPE = GENERATE_AND_UPSCALE

// Allowed image extensions for upload
var ALLOWED_PROCESS_TYPES = []ProcessType{GENERATE, UPSCALE, GENERATE_AND_UPSCALE}

// Maximum size of a custom image sent to upscale
const MAX_UPSCALE_IMAGE_SIZE = 1024 * 1024 * 10
