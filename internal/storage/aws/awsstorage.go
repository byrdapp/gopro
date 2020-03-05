package aws

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/blixenkrone/gopro/pkg/logger"
)

var log = logger.NewLogger()

type s3Storage struct {
	session     *session.Session
	ctx         context.Context
	contentType string
}

type AWSStorer interface {
	StoreFile(io.Reader) error
}

func NewSession(s AWSStorer, ctx context.Context, contentType string) (*s3Storage, error) {
	session, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3NorthRegion),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS"), os.Getenv("AWS_SECRET"), ""),
	})
	if err != nil {
		return nil, err
	}
	return &s3Storage{session, ctx, contentType}, nil
}

func (s *s3Storage) StoreFile(file io.Reader, name string) error {
	uploader := s3manager.NewUploader(s.session)
	_, err := uploader.UploadWithContext(s.ctx, &s3manager.UploadInput{
		Body:                 file,
		Bucket:               aws.String("byrd-bookings"),
		Key:                  aws.String("/simontestdir/" + name),
		ServerSideEncryption: aws.String("AES256"),
		ContentType:          aws.String(s.contentType),
	})
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info("storage upload complete")
	return nil
}
