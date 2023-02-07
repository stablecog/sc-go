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

// This is the name of the channel in redis, that is used to broadcast webhook events
// We use this, because we may have multiple replicas of our service, and we want to
// broadcast these events to all of them to make sure the consumer that cares about
// the message gets it
const COG_REDIS_WEBHOOK_QUEUE_CHANNEL = "queue:webhook"

// Allowed image extensions used by various APIs
type ImageExtension string

const (
	PNG  ImageExtension = "png"
	JPG  ImageExtension = "jpg"
	JPEG ImageExtension = "jpeg"
	WEBP ImageExtension = "webp"
)

// Default image extension for generate
const DEFAULT_GENERATE_OUTPUT_EXTENSION = WEBP
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
