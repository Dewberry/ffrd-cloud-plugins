package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/USACE/filestore"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Config struct {
	config filestore.S3FSConfig
}

func NewS3Conf(accessKeyENV, secretAccessKeyENV, regionENV, bucketENV string) *S3Config {
	s3Conf := filestore.S3FSConfig{}
	s3Conf.S3Id = os.Getenv(accessKeyENV)
	s3Conf.S3Key = os.Getenv(secretAccessKeyENV)
	s3Conf.S3Region = os.Getenv(regionENV)
	s3Conf.S3Bucket = os.Getenv(bucketENV)
	return &S3Config{s3Conf}
}

func (s S3Config) Init() *filestore.S3FS {
	fs, err := filestore.NewFileStore(s.config)
	if err != nil {
		panic(err)
	}
	return fs.(*filestore.S3FS)
}

func FileStoreInit(bucket string) *filestore.FileStore {
	var fs filestore.FileStore
	var err error
	config := filestore.S3FSConfig{
		S3Id:     os.Getenv("AWS_ACCESS_KEY_ID"),
		S3Key:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
		S3Region: os.Getenv("AWS_REGION"),
		S3Bucket: bucket,
	}

	fs, err = filestore.NewFileStore(config)
	if err != nil {
		panic(err)
	}
	return &fs
}

type S3Controller struct {
	Sess  *session.Session
	S3Svc *s3.S3
}

func SessionManager() (S3Controller, error) {
	region := os.Getenv("AWS_REGION")
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	var s3Ctrl S3Controller
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return s3Ctrl, fmt.Errorf("error creating s3 session: %w", err)
	}
	s3Ctrl.Sess = sess
	s3Ctrl.S3Svc = s3.New(sess)
	return s3Ctrl, nil

}
func KeyExists(s3Ctrl *s3.S3, bucket string, key string) (bool, error) {
	_, err := s3Ctrl.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound": // s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we hardwire a string
				return false, nil
			default:
				return false, fmt.Errorf("KeyExists: %w", err)
			}
		}
		return false, fmt.Errorf("KeyExists: %w", err)
	}
	return true, nil
}

func GetDownloadPresignedURL(s3Ctrl *s3.S3, bucket, key string, expDays int) (string, error) {
	duration := time.Duration(expDays) * 24 * time.Hour
	req, _ := s3Ctrl.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return req.Presign(duration)
}
