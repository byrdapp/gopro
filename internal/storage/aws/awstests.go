package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

type BucketRef string

const (
	ImageTestReference = "images"
)

var testTypePath = map[BucketRef]string{"images": "images"}

type S3TestInstance struct {
	TestPathType string
	session      *s3manager.Downloader
}

func NewTest(path BucketRef) (*S3TestInstance, error) {
	pathType, ok := testTypePath[path]
	if !ok {
		return nil, errors.New("bucket reference path for test material not found ")
	}
	log.Infof("getting path %s", pathType)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3Region),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS"), os.Getenv("AWS_SECRET"), ""),
	})
	if err != nil {
		return nil, errors.Errorf("aws session failed: %s", err)
	}
	return &S3TestInstance{
		pathType,
		s3manager.NewDownloader(sess),
	}, nil
}

type TestMaterial *aws.WriteAtBuffer

func (s3test *S3TestInstance) Download(fileName string) (TestMaterial, error) {
	var buf aws.WriteAtBuffer
	_, err := s3test.session.Download(&buf, &s3.GetObjectInput{
		Bucket: aws.String(s3TestBucket + "/" + s3test.TestPathType),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "download from s3 failed")
	}

	return &buf, nil
}
