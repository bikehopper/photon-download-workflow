package app_config

import "os"

type AppConfig struct {
	ArchiveUrl    string
	S3Region      string
	S3EndpointUrl string
	Bucket        string
	ArchiveKey    string
	TemporalUrl   string
}

func New() *AppConfig {
	return &AppConfig{
		ArchiveUrl:    getEnv("ARCHIVE_URL", ""),
		S3Region:      getEnv("S3_REGION", "us-east-1"),
		S3EndpointUrl: getEnv("S3_ENDPOINT_URL", ""),
		Bucket:        getEnv("BUCKET", ""),
		ArchiveKey:    getEnv("ARCHIVE_KEY", ""),
		TemporalUrl:   getEnv("TEMPORAL_URL", "localhost:7233"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
