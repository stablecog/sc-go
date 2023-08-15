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

func GetURLFromImagePath(s3UrlStr string) string {
	baseUrl := EnsureTrailingSlash(GetEnv("BUCKET_BASE_URL", "https://b.stablecog.com/"))

	return baseUrl + s3UrlStr
}

func GetURLFromAudioFilePath(s3UrlStr string) string {
	baseUrl := EnsureTrailingSlash(GetEnv("BUCKET_VOICEOVER_URL", "https://bvoi.stablecog.com/"))

	return baseUrl + s3UrlStr
}

func GetCorsOrigins() []string {
	if GetEnv("PRODUCTION", "false") == "true" {
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
