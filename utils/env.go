package utils

import (
	"fmt"
	"os"
)

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetDefaultServerUrl() string {
	return GetEnv("PUBLIC_DEFAULT_SERVER_URL", "")
}

type S3Data struct {
	BucketPublic                   string
	BucketPrivate                  string
	BucketPrivateOutputQueueFolder string
	Hostname                       string
	PrivateUrl                     string
	AccountId                      string
	AccessKeyId                    string
	SecretKey                      string
}

func GetS3Data() S3Data {
	return S3Data{
		BucketPublic:                   "stablecog",
		BucketPrivate:                  "stablecog-private",
		BucketPrivateOutputQueueFolder: "queue/output",
		Hostname:                       fmt.Sprintf("%s.r2.cloudflarestorage.com", os.Getenv("CLOUDFLARE_ACCOUNT_ID")),
		PrivateUrl:                     os.Getenv("R2_PRIVATE_URL"),
		AccountId:                      os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		AccessKeyId:                    os.Getenv("R2_ACCESS_KEY_ID"),
		SecretKey:                      os.Getenv("R2_SECRET_ACCESS_KEY"),
	}
}
