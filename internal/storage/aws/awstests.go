package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"

	"github.com/blixenkrone/gopro/pkg/file"
)

type BucketRef string

const (
	ImageBucketReference = "images"
	VideoBucketReference = "videos"
)

var testTypePath = map[BucketRef]string{"images": ImageBucketReference, "videos": VideoBucketReference}

type S3TestMaterial struct {
	testPathType string
	session      *session.Session
	fileName     string
	byteValue    []byte
}

func ParseCredentials() error {
	_, err := file.SetEnvFileVars("../../../")
	if err != nil {
		return err
	}
	return nil
}

func GetTestMaterial(path BucketRef, fileName string) (*S3TestMaterial, error) {
	pathType, ok := testTypePath[path]
	if !ok {
		return nil, errors.New("bucket reference path for test material not found for: " + string(path))
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s3NorthRegion),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS"), os.Getenv("AWS_SECRET"), ""),
	})
	if err != nil {
		return nil, errors.Errorf("aws session failed: %s", err)
	}

	bucketPath := s3TestBucket + "/" + testTypePath[path]
	log.Infof("getting AWS bucket path/file %s/%s", bucketPath, fileName)

	var buf aws.WriteAtBuffer
	dl := s3manager.NewDownloader(sess)
	_, err = dl.Download(&buf, &s3.GetObjectInput{
		Bucket: aws.String(bucketPath),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "download from s3 failed")
	}

	return &S3TestMaterial{
		pathType,
		sess,
		fileName,
		buf.Bytes(),
	}, nil
}

func (s3test *S3TestMaterial) Bytes() []byte {
	return s3test.byteValue
}
