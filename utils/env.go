package utils

import (
	"net/url"
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

func ParseS3UrlToURL(s3UrlStr string) (string, error) {
	baseUrl := EnsureTrailingSlash(GetEnv("BUCKET_BASE_URL", "https://b.stablecog.com/"))

	s3Url, err := url.Parse(s3UrlStr)
	if err != nil {
		return s3UrlStr, err
	}

	if s3Url.Scheme != "s3" {
		return s3UrlStr, nil
	}

	// Remove leading slash from path
	s3Url.Path = s3Url.Path[1:]

	return baseUrl + s3Url.Path, nil
}
