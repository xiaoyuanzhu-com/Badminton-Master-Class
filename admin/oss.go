package main

import (
	"context"
	"log"
	"os"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// ossConfig holds Aliyun OSS configuration from environment variables.
type ossConfig struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	AccessKeySecret string
	ObjectKey       string
}

// loadOSSConfig reads OSS settings from environment variables.
// Returns nil if any required variable is missing.
func loadOSSConfig() *ossConfig {
	endpoint := os.Getenv("BMC_OSS_ENDPOINT")
	bucket := os.Getenv("BMC_OSS_BUCKET")
	keyID := os.Getenv("BMC_OSS_ACCESS_KEY_ID")
	keySecret := os.Getenv("BMC_OSS_ACCESS_KEY_SECRET")

	if endpoint == "" || bucket == "" || keyID == "" || keySecret == "" {
		return nil
	}

	objectKey := os.Getenv("BMC_OSS_OBJECT_KEY")
	if objectKey == "" {
		objectKey = "bmc.db"
	}

	return &ossConfig{
		Endpoint:        endpoint,
		Bucket:          bucket,
		AccessKeyID:     keyID,
		AccessKeySecret: keySecret,
		ObjectKey:       objectKey,
	}
}

// uploadToOSS is the default implementation that uploads a file to Aliyun OSS.
// It is a variable so tests can replace it with a mock.
var uploadToOSS = uploadToOSSImpl

func uploadToOSSImpl(cfg *ossConfig, filePath string) error {
	client := oss.NewClient(&oss.Config{
		Endpoint:            oss.Ptr(cfg.Endpoint),
		CredentialsProvider: credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret),
	})

	_, err := client.PutObjectFromFile(context.Background(), &oss.PutObjectRequest{
		Bucket: oss.Ptr(cfg.Bucket),
		Key:    oss.Ptr(cfg.ObjectKey),
	}, filePath)
	return err
}

// tryUploadToOSS attempts to upload the DB file to OSS.
// It skips silently if OSS is not configured and logs errors on failure.
func tryUploadToOSS(dbPath string) {
	cfg := loadOSSConfig()
	if cfg == nil {
		return
	}

	if err := uploadToOSS(cfg, dbPath); err != nil {
		log.Printf("OSS upload failed: %v", err)
	} else {
		log.Printf("OSS upload succeeded: %s -> %s/%s", dbPath, cfg.Bucket, cfg.ObjectKey)
	}
}
