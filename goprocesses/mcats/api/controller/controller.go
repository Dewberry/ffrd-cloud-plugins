package controller

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/USACE/filestore"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
)

type Controller struct {
	DB *sqlx.DB
	// S3FS *filestore.S3FS
	// FS   *filestore.FileStore
}

func NewController() (*Controller, error) {
	db := sqlx.MustOpen("pgx", os.Getenv("API_DB_CREDS"))

	// s3Conf := NewS3Conf("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "AWS_S3_BUCKET")
	// s3fs := s3Conf.Init()

	// fs := FileStoreInit("S3")
	crtl := &Controller{db}
	return crtl, crtl.DB.Ping()
}

func (crtl *Controller) Close() error {
	return crtl.DB.Close()
}

func (crtl *Controller) Ping(c echo.Context) error {

	message := map[string]interface{}{"database_healthy": true, "s3_connection_healthy": true}
	msg := make([]string, 0)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := crtl.DB.PingContext(ctx)
	if err != nil {
		msg = append(msg, err.Error())
		message["database_healthy"] = false
	}

	// err = crtl.S3FS.Ping()
	// if err != nil {
	// 	msg = append(msg, "unable to connect, verify credentials are in place and valid")
	// 	// Use to debug if credentials are in place and valid
	// 	// msg = append(msg, err.Error())
	// 	message["s3_connection_healthy"] = false
	// }

	if len(msg) > 0 {
		message["message"] = strings.Join(msg, "; ")
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, message)
}

// MCAT / Filestore
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

// FileStoreInit initializes the filestore object
func FileStoreInit(store string) *filestore.FileStore {

	var fs filestore.FileStore
	var err error
	switch store {
	case "LOCAL":
		fs, err = filestore.NewFileStore(filestore.BlockFSConfig{})
		if err != nil {
			panic(err)
		}
	case "S3":
		config := filestore.S3FSConfig{
			S3Id:     os.Getenv("AWS_ACCESS_KEY_ID"),
			S3Key:    os.Getenv("AWS_SECRET_ACCESS_KEY"),
			S3Region: os.Getenv("AWS_REGION"),
			S3Bucket: os.Getenv("AWS_S3_BUCKET"),
		}

		fs, err = filestore.NewFileStore(config)
		if err != nil {
			panic(err)
		}
	}
	return &fs
}
