package utils

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitEnvironment(t *testing.T) {
	// Reset environment variables after each test.
	defer os.Clearenv()

	// ! Test disabled since nothing is required currently, due to brekaing other tests
	// t.Run("fails when required variables are missing", func(t *testing.T) {
	// 	// Clearing out any potential set environment variables
	// 	os.Clearenv()

	// 	// Reset instance before each test to ensure fresh start.
	// 	instance = nil
	// 	once = sync.Once{}

	// 	assert.Panics(t, func() {
	// 		GetEnv()
	// 	}, "Expected panic due to missing required environment variables")
	// })

	t.Run("sets defaults when optional variables are unset", func(t *testing.T) {
		// Reset instance before each test to ensure fresh start.
		instance = nil
		once = sync.Once{}
		os.Clearenv()

		os.Setenv("CLIPAPI_SECRET", "test-secret")
		os.Setenv("STRIPE_SECRET_KEY", "stripe-secret")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("REDIS_CONNECTION_STRING", "redis-connection")
		os.Setenv("QDRANT_URL", "qdrant-url")
		os.Setenv("PUBLIC_SUPABASE_REFERENCE_ID", "public-supabase")
		os.Setenv("SUPABASE_ADMIN_KEY", "admin-key")

		env := GetEnv()
		assert.Equal(t, false, env.Production)
		assert.Equal(t, 8000, env.Port)
		assert.Equal(t, "http://localhost:8000", env.PublicApiUrl)
		assert.Equal(t, "invalid", env.ScWorkerWebhookSecret)
		assert.Equal(t, false, env.RunMigrations)
		assert.Equal(t, "https://b.stablecog.com/", env.BucketBaseUrl)
		assert.Equal(t, "https://bvoi.stablecog.com/", env.BucketVoiceverUrl)
		assert.Equal(t, "price_1Mf591ATa0ehBYTA6ggpEEkA", env.StripeUltimatePriceID)
		assert.Equal(t, "price_1Mf50bATa0ehBYTAPOcfnOjG", env.StripeProPriceID)
		assert.Equal(t, "price_1Mf56NATa0ehBYTAHkCUablG", env.StripeStarterPriceID)
		assert.Equal(t, "prod_NTzE0C8bEuIv6F", env.StripeUltimateProductID)
		assert.Equal(t, "prod_NTzCojAHPw6tbX", env.StripeProProductID)
		assert.Equal(t, "prod_NPuwbni7ZNkHDO", env.StripeStarterProductID)
		assert.Equal(t, "1", env.StripeLargePackPriceID)
		assert.Equal(t, "2", env.StripeMediumPackPriceID)
		assert.Equal(t, "3", env.StripeMegaPackPriceID)
		assert.Equal(t, "1", env.StripeLargePackProductID)
		assert.Equal(t, "2", env.StripeMediumPackProductID)
		assert.Equal(t, "3", env.StripeMegaPackProductID)
		assert.Equal(t, false, env.GithubActions)
		assert.Equal(t, "postgres", env.PostgresDB)
		assert.Equal(t, "postgres", env.PostgresUser)
		assert.Equal(t, "127.0.0.1", env.PostgresHost)
		assert.Equal(t, 5432, env.PostgresPort)
		assert.Equal(t, "redis-connection", env.RedisConnectionString)
		assert.Equal(t, false, env.MockRedis)
		assert.Equal(t, "stablecog", env.QdrantCollectionName)
		assert.Equal(t, "TEST.Q", env.RabbitMQQueueName)
		assert.Equal(t, "amqp://guest:guest@localhost:5672/", env.RabbitMQAMQPUrl)
		assert.Equal(t, "us-east-1", env.S3Img2ImgRegion)
		assert.Equal(t, "img2img", env.S3Img2ImgBucketName)
		assert.Equal(t, "", env.S3Img2ImgAccessKey)
		assert.Equal(t, "", env.S3Img2ImgSecretKey)
		assert.Equal(t, "", env.S3Img2ImgEndpoint)
		assert.Equal(t, "us-east-1", env.S3Region)
		assert.Equal(t, "stablecog", env.S3BucketName)
		assert.Equal(t, "insecurePassword", env.DataEncryptionPassword)
		assert.Equal(t, "http://localhost:3000", env.OauthRedirectBase)
	})

	t.Run("sets values based on environment variables", func(t *testing.T) {
		// Reset instance before each test to ensure fresh start.
		instance = nil
		once = sync.Once{}
		os.Clearenv()

		// Set required environment variables
		os.Setenv("CLIPAPI_SECRET", "test-secret")
		os.Setenv("STRIPE_SECRET_KEY", "stripe-secret")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("REDIS_CONNECTION_STRING", "redis-connection")
		os.Setenv("QDRANT_URL", "qdrant-url")
		os.Setenv("PUBLIC_SUPABASE_REFERENCE_ID", "public-supabase")
		os.Setenv("SUPABASE_ADMIN_KEY", "admin-key")

		// Set optional environment variables
		os.Setenv("PRODUCTION", "true")
		os.Setenv("PORT", "9090")
		os.Setenv("PUBLIC_API_URL", "http://example.com")

		env := GetEnv()

		// Assert that values are set based on environment variables
		assert.Equal(t, true, env.Production)
		assert.Equal(t, 9090, env.Port)
		assert.Equal(t, "http://example.com", env.PublicApiUrl)

		// Cleanup: unset optional environment variables for next tests
		os.Unsetenv("PRODUCTION")
		os.Unsetenv("PORT")
		os.Unsetenv("PUBLIC_API_URL")
	})

	t.Run("test GetPathFromS3URL", func(t *testing.T) {
		// Reset instance before each test to ensure fresh start.
		instance = nil
		once = sync.Once{}
		os.Clearenv()

		// Set required environment variables
		os.Setenv("CLIPAPI_SECRET", "test-secret")
		os.Setenv("STRIPE_SECRET_KEY", "stripe-secret")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("REDIS_CONNECTION_STRING", "redis-connection")
		os.Setenv("QDRANT_URL", "qdrant-url")
		os.Setenv("PUBLIC_SUPABASE_REFERENCE_ID", "public-supabase")
		os.Setenv("SUPABASE_ADMIN_KEY", "admin-key")

		os.Setenv("BUCKET_BASE_URL", "http://test.com/")

		env := GetEnv()

		assert.Equal(t, "http://test.com/cc70edec-b6ff-42c5-8726-957bbd8fc212.jpeg", env.GetURLFromImagePath("cc70edec-b6ff-42c5-8726-957bbd8fc212.jpeg"))

		os.Unsetenv("BUCKET_BASE_URL")
	})

	t.Run("test GetUrlFromAudioFilePath", func(t *testing.T) {
		// Reset instance before each test to ensure fresh start.
		instance = nil
		once = sync.Once{}
		os.Clearenv()

		// Set required environment variables
		os.Setenv("CLIPAPI_SECRET", "test-secret")
		os.Setenv("STRIPE_SECRET_KEY", "stripe-secret")
		os.Setenv("POSTGRES_PASSWORD", "password")
		os.Setenv("REDIS_CONNECTION_STRING", "redis-connection")
		os.Setenv("QDRANT_URL", "qdrant-url")
		os.Setenv("PUBLIC_SUPABASE_REFERENCE_ID", "public-supabase")
		os.Setenv("SUPABASE_ADMIN_KEY", "admin-key")

		os.Setenv("BUCKET_VOICEOVER_URL", "http://testv.com/")

		env := GetEnv()

		assert.Equal(t, "http://testv.com/cc70edec-b6ff-42c5-8726-957bbd8fc212.mp3", env.GetURLFromAudioFilePath("cc70edec-b6ff-42c5-8726-957bbd8fc212.mp3"))

		os.Unsetenv("BUCKET_VOICEOVER_URL")
	})
}

func TestGetPathFromS3URL(t *testing.T) {
	path := "s3://stablecog/cc70edec-b6ff-42c5-8726-957bbd8fc212.jpeg"
	parsed, err := GetPathFromS3URL(path)
	assert.Nil(t, err)
	assert.Equal(t, "cc70edec-b6ff-42c5-8726-957bbd8fc212.jpeg", parsed)
}
