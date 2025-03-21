package services

import (
	"github.com/bdmehedi/s3-media-resolver/internal/config"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct{}

func NewS3Service() *S3Service {
	return &S3Service{}
}

func (s *S3Service) CreatePresignedURL(path string) (string, error) {
	// Remove leading slash if present
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(config.AppConfig.S3Bucket),
		Key:    aws.String(path),
	}

	presigned, err := config.AppConfig.Presigner.PresignGetObject(
		config.AppConfig.Context,
		input,
		s3.WithPresignExpires(config.AppConfig.Expiry),
	)
	if err != nil {
		return "", err
	}

	return presigned.URL, nil
}
