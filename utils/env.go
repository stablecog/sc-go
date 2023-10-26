package utils

import (
	"net/url"
	"sync"

	"github.com/caarlos0/env/v9"
	"github.com/stablecog/sc-go/log"
)

type SCEnv struct {
	Production   bool   `env:"PRODUCTION" envDefault:"false"`
	Port         int    `env:"PORT" envDefault:"8000"`
	PublicApiUrl string `env:"PUBLIC_API_URL" envDefault:"http://localhost:8000"` // Used for thing such as, building the webhook URL for sc-worker to send results to
	// Content moderation and translator
	OpenAIApiKey        string `env:"OPENAI_API_KEY"`
	PrivateLinguaAPIUrl string `env:"PRIVATE_LINGUA_API_URL"` // Corresponds to sc-go/language server
	TranslatorCogURL    string `env:"TRANSLATOR_COG_URL"`     // Corresponds to translation service/cog
	// CLIP Service, built into sc-worker
	ClipAPISecret   string   `env:"CLIPAPI_SECRET"` // required
	ClipAPIURLs     []string `env:"CLIPAPI_URLS" envSeparator:","`
	ClipAPIEndpoint string   `env:"CLIPAPI_ENDPOINT"` // Probably not necessary for most uses, used for internal stuff
	// Shared secret between sc-worker and sc-server
	ScWorkerWebhookSecret string `env:"SC_WORKER_WEBHOOK_SECRET" envDefault:"invalid"`
	// Whether to run DB migrations on startup, can only be done in the local environment (not on a supabase database)
	RunMigrations bool `env:"RUN_MIGRATIONS" envDefault:"false"`
	// CDN URLs for assets
	BucketBaseUrl     string `env:"BUCKET_BASE_URL" envDefault:"https://b.stablecog.com/"` // Public CDN URL
	BucketVoiceverUrl string `env:"BUCKET_VOICEOVER_URL" envDefault:"https://bvoi.stablecog.com/"`
	// Stripe
	StripeSecretKey      string `env:"STRIPE_SECRET_KEY"`      // required
	StripeEndpointSecret string `env:"STRIPE_ENDPOINT_SECRET"` // Required for stripe webhooks
	// Stripe subscription price + product IDs
	StripeUltimatePriceID   string `env:"STRIPE_ULTIMATE_PRICE_ID" envDefault:"price_1Mf591ATa0ehBYTA6ggpEEkA"`
	StripeProPriceID        string `env:"STRIPE_PRO_PRICE_ID" envDefault:"price_1Mf50bATa0ehBYTAPOcfnOjG"`
	StripeStarterPriceID    string `env:"STRIPE_STARTER_PRICE_ID" envDefault:"price_1Mf56NATa0ehBYTAHkCUablG"`
	StripeUltimateProductID string `env:"STRIPE_ULTIMATE_PRODUCT_ID" envDefault:"prod_NTzE0C8bEuIv6F"`
	StripeProProductID      string `env:"STRIPE_PRO_PRODUCT_ID" envDefault:"prod_NTzCojAHPw6tbX"`
	StripeStarterProductID  string `env:"STRIPE_STARTER_PRODUCT_ID" envDefault:"prod_NPuwbni7ZNkHDO"`
	// Stripe ad-hoc purchase price + product IDs
	StripeLargePackPriceID    string `env:"STRIPE_LARGE_PACK_PRICE_ID" envDefault:"1"`
	StripeMediumPackPriceID   string `env:"STRIPE_MEDIUM_PACK_PRICE_ID" envDefault:"2"`
	StripeMegaPackPriceID     string `env:"STRIPE_MEGA_PACK_PRICE_ID" envDefault:"3"`
	StripeLargePackProductID  string `env:"STRIPE_LARGE_PACK_PRODUCT_ID" envDefault:"1"`
	StripeMediumPackProductID string `env:"STRIPE_MEDIUM_PACK_PRODUCT_ID" envDefault:"2"`
	StripeMegaPackProductID   string `env:"STRIPE_MEGA_PACK_PRODUCT_ID" envDefault:"3"`
	// Discord webhooks
	DiscordWebhookUrl       string `env:"DISCORD_WEBHOOK_URL"`        // For health notifications in cron, Optional
	DiscordWebhookUrlDeploy string `env:"DISCORD_WEBHOOK_URL_DEPLOY"` // For deploy notifications in server, Optional
	DiscordWebhookUrlNewSub string `env:"DISCORD_WEBHOOK_URL_NEWSUB"` // For new sub notifications in server, Optional
	GeoIpWebhook            string `env:"GEOIP_WEBHOOK"`              // For geoip notifications in server, Optional
	// Whether running in github actions, basically whether to use Postgres or not in tests (will use SQLite if false)
	GithubActions bool `env:"GITHUB_ACTIONS" envDefault:"false"` // Whether we're running in Github Actions
	// PostgreSQL
	PostgresDB       string `env:"POSTGRES_DB" envDefault:"postgres"`    // Postgres DB Name
	PostgresUser     string `env:"POSTGRES_USER" envDefault:"postgres"`  // Postgres DB User
	PostgresPassword string `env:"POSTGRES_PASSWORD"`                    // Postgres DB Password, required
	PostgresHost     string `env:"POSTGRES_HOST" envDefault:"127.0.0.1"` // Postgres DB Host
	PostgresPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`      // Postgres DB Port
	// Redis
	RedisConnectionString string `env:"REDIS_CONNECTION_STRING" envDefault:"redis://localhost:6379/0"` // Redis connection string, required
	MockRedis             bool   `env:"MOCK_REDIS" envDefault:"false"`                                 // Whether to mock redis for tests
	// Qdrant
	QdrantUrl            string `env:"QDRANT_URL"`                                    // Qdrant URL, required
	QdrantUsername       string `env:"QDRANT_USERNAME"`                               // Qdrant Username, Optional
	QdrantPassword       string `env:"QDRANT_PASSWORD"`                               // Qdrant Password, Optional
	QdrantCollectionName string `env:"QDRANT_COLLECTION_NAME" envDefault:"stablecog"` // Qdrant Collection Name
	// Supabase
	PublicSupabaseReferenceID string `env:"PUBLIC_SUPABASE_REFERENCE_ID"` // Supabase reference ID, required
	SupabaseAdminKey          string `env:"SUPABASE_ADMIN_KEY"`           // Supabase admin key, required
	GotrueURL                 string `env:"GOTRUE_URL"`                   // Gotrue URL, Optional (Only for self-hosted supabase)
	// RabbitMQ
	RabbitMQQueueName string `env:"RABBITMQ_QUEUE_NAME" envDefault:"TEST.Q"`                           // RabbitMQ queue name
	RabbitMQAMQPUrl   string `env:"RABBITMQ_AMQP_URL" envDefault:"amqp://guest:guest@localhost:5672/"` // RabbitMQ AMQP URL
	// S3 img2img bucket, for user-uploaded images
	S3Img2ImgRegion     string `env:"S3_IMG2IMG_REGION" envDefault:"us-east-1"`    // S3 region
	S3Img2ImgBucketName string `env:"S3_IMG2IMG_BUCKET_NAME" envDefault:"img2img"` // S3 bucket
	S3Img2ImgAccessKey  string `env:"S3_IMG2IMG_ACCESS_KEY" envDefault:""`         // S3 access key
	S3Img2ImgSecretKey  string `env:"S3_IMG2IMG_SECRET_KEY" envDefault:""`         // S3 secret key
	S3Img2ImgEndpoint   string `env:"S3_IMG2IMG_ENDPOINT" envDefault:""`           // S3 endpoint
	// S3 bucket, for all sc-worker uploaded assets. Primarily used for deleting user data in cron CLI
	S3Region     string `env:"S3_REGION" envDefault:"us-east-1"`      // S3 region
	S3BucketName string `env:"S3_BUCKET_NAME" envDefault:"stablecog"` // S3 bucket
	S3AccessKey  string `env:"S3_ACCESS_KEY"`                         // S3 access key
	S3SecretKey  string `env:"S3_SECRET_KEY"`
	S3Endpoint   string `env:"S3_ENDPOINT"` // S3 endpoint
	// Analytics (Optional group)
	PosthogApiKey   string `env:"POSTHOG_API_KEY"`  // Posthog API Key, Optional
	PosthogEndpoint string `env:"POSTHOG_ENDPOINT"` // Posthog Endpoint, Optional
	MixpanelApiKey  string `env:"MIXPANEL_API_KEY"` // Mixpanel API Key, Optional
	// Discord bot
	DiscordBotToken string `env:"DISCORD_BOT_TOKEN"` // Discord bot token, only required for discobot
	// Oauth service, used by services such as Raycast
	DataEncryptionPassword string `env:"DATA_ENCRYPTION_PASSWORD" envDefault:"insecurePassword"` // Data encryption password
	OauthRedirectBase      string `env:"OAUTH_REDIRECT_BASE" envDefault:"http://localhost:3000"` // Oauth redirect base
}

// The package-level instance and its initialization controls.
var (
	instance *SCEnv
	once     sync.Once
	mutex    sync.Mutex
)

// initEnvironment is a helper function to initialize the environment.
func initEnvironment() {
	cfg := SCEnv{}
	if err := env.Parse(&cfg); err != nil {
		log.Error("Error parsing environment", "err", err)
		panic("Error parsing environment")
	}
	instance = &cfg
}

// GetEnv provides a thread-safe way to get the environment.
func GetEnv() *SCEnv {
	once.Do(initEnvironment)
	mutex.Lock()
	defer mutex.Unlock()
	return instance
}

func (e *SCEnv) GetURLFromImagePath(s3UrlStr string) string {
	baseUrl := EnsureTrailingSlash(e.BucketBaseUrl)

	return baseUrl + s3UrlStr
}

func (e *SCEnv) GetURLFromAudioFilePath(s3UrlStr string) string {
	baseUrl := EnsureTrailingSlash(e.BucketVoiceverUrl)

	return baseUrl + s3UrlStr
}

func (e *SCEnv) GetCorsOrigins() []string {
	if e.Production {
		return []string{
			"http://localhost:4173",
			"http://localhost:5173",
			"http://localhost:3000",
			"http://localhost:8000",
			"https://stablecog-git-v21-stablecog.vercel.app",
			"https://stablecog-git-v3-stablecog.vercel.app",
			"https://stablecog.com",
		}
	}
	return []string{
		"http://localhost:3000",
		"http://localhost:4173",
		"http://localhost:5173",
		"http://localhost:8000",
		"https://stablecog-git-v21-stablecog.vercel.app",
		"https://stablecog-git-v3-stablecog.vercel.app",
		"https://stablecog.com",
	}
}

func GetPathFromS3URL(s3UrlStr string) (string, error) {
	s3Url, err := url.Parse(s3UrlStr)
	if err != nil {
		return s3UrlStr, err
	}

	if s3Url.Scheme != "s3" {
		return s3UrlStr, nil
	}

	// Remove leading slash from path
	s3Url.Path = s3Url.Path[1:]

	return s3Url.Path, nil
}
