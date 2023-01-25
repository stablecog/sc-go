package shared

// Generation related
const MAX_GENERATE_WIDTH = 768
const MAX_GENERATE_HEIGHT = 768
const MAX_PRO_PIXEL_STEPS = 640 * 640 * 50

// Prompt related
const MAX_PROMPT_LENGTH = 500

// The name of the redis stream used to enqueue cog requests
const COG_REDIS_QUEUE = "input_queue"
