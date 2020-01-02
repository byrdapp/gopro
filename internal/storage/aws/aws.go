package aws

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	s3NorthRegion      = "eu-north-1"
	// s3CentralRegion    = "eu-central-1"
	s3SecretBucket     = "byrd-secrets"
	s3AccountingBucket = "byrd-accounting"
	s3TestBucket       = "/byrd-tests"
)

// NewUpload returns url location for where the file has been placed
func NewUpload(file []byte, dateStamp string) (string, error) {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3NorthRegion),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS"), os.Getenv("AWS_SECRET"), ""),
	})
	if err != nil {
		return "", err
	}
	sess := session.Must(s, err)
	location, err := uploader(sess, file, dateStamp)
	if err != nil {
		return "", err
	}
	return location, nil
}

// Uploader S3 uploader
func uploader(s *session.Session, file []byte, dateStamp string) (string, error) {
	uploader := s3manager.NewUploader(s)
	dir := "/" + dateStamp[:7] + "/"
	fileName := "media-subscriptions_" + dateStamp[:7] + ".pdf"
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:                 bytes.NewBuffer(file),
		Bucket:               aws.String(s3AccountingBucket),
		Key:                  aws.String(dir + string(fileName)),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return "", fmt.Errorf("Failed to upload file:  %v", err)
	}
	fmt.Printf("Successfully uploaded file to: %s\n", aws.StringValue(&result.Location))
	return dir, nil
}

// GetAWSSecrets -
func GetAWSSecrets(fileName string) []byte {
	buf := &aws.WriteAtBuffer{}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3NorthRegion),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS"), os.Getenv("AWS_SECRET"), ""),
	})
	if err != nil {
		log.Errorf("Didnt get aws CC's: %s", err)
	}
	dl := s3manager.NewDownloader(sess)
	_, err = dl.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(s3SecretBucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		log.Errorf("Didnt get aws DL: %s", err)
	}
	return buf.Bytes()
}
