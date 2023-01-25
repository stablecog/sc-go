package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("MY_ENV", "value")
	defer os.Unsetenv("MY_ENV")

	assert.Equal(t, "value", GetEnv("MY_ENV", "default"))
	assert.Equal(t, "default", GetEnv("MY_ENV_UNKNOWN", "default"))
}

func TestGetDefaultServerUrl(t *testing.T) {
	os.Setenv("PUBLIC_DEFAULT_SERVER_URL", "testgetdefaultserverurl")
	defer os.Unsetenv("PUBLIC_DEFAULT_SERVER_URL")
	assert.Equal(t, "testgetdefaultserverurl", GetDefaultServerUrl())
	os.Setenv("PUBLIC_DEFAULT_SERVER_URL", "different")
	assert.NotEqual(t, "testgetdefaultserverurl", GetDefaultServerUrl())
}

func TestGetS3Data(t *testing.T) {
	os.Setenv("CLOUDFLARE_ACCOUNT_ID", "cloudflare_id")
	defer os.Unsetenv("CLOUDFLARE_ACCOUNT_ID")
	os.Setenv("R2_PRIVATE_URL", "private_url")
	defer os.Unsetenv("R2_PRIVATE_URL")
	os.Setenv("R2_ACCESS_KEY_ID", "access_key_id")
	defer os.Unsetenv("R2_ACCESS_KEY_ID")
	os.Setenv("R2_SECRET_ACCESS_KEY", "r2_secret_access_key")
	defer os.Unsetenv("R2_SECRET_ACCESS_KEY")

	data := GetS3Data()
	assert.Equal(t, "stablecog", data.BucketPublic)
	assert.Equal(t, "stablecog-private", data.BucketPrivate)
	assert.Equal(t, "queue/output", data.BucketPrivateOutputQueueFolder)
	assert.Equal(t, "cloudflare_id.r2.cloudflarestorage.com", data.Hostname)
	assert.Equal(t, "private_url", data.PrivateUrl)
	assert.Equal(t, "cloudflare_id", data.AccountId)
	assert.Equal(t, "access_key_id", data.AccessKeyId)
	assert.Equal(t, "r2_secret_access_key", data.SecretKey)
}
