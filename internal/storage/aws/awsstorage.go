package storage

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/blixenkrone/gopro/pkg/logger"
)

var log = logger.NewLogger()

type S3Storage struct {
}

func StoreFiles(file io.Reader) error {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3Region),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS"), os.Getenv("AWS_SECRET"), ""),
	})
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(s)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body:                 file,
		Bucket:               aws.String(s3Bucket),
		Key:                  aws.String(string("fileName")),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info("success to aws s3")
	return nil
}
