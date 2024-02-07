package mcats

import (
	"app/utils"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// UploadGeoJSONToS3AndGeneratePresignedURLs uploads GeoJSON data to S3 and generates presigned URLs.
// It returns a slice of presigned URLs and any error encountered.
func uploadGeoJSONToS3AndGeneratePresignedURLs(collectionJson map[string][]byte, g01Key, outPutPrefix, bucket string, urlExpDay int, s3Ctrl utils.S3Controller) ([]string, error) {
	var presignedUrlArr []string

	for key, value := range collectionJson {
		// Generate the output file name based on the key and the original file name.
		g01FileName := strings.TrimSuffix(filepath.Base(g01Key), filepath.Ext(filepath.Base(g01Key)))
		outputKey := fmt.Sprintf("%s/%s_%s.geojson", outPutPrefix, g01FileName, key)

		// Create a new uploader using the S3 session.
		uploader := s3manager.NewUploader(s3Ctrl.Sess)

		// Upload the JSON value to S3.
		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(outputKey),
			Body:        bytes.NewReader(value),
			ContentType: aws.String("binary/octet-stream"),
		})
		if err != nil {
			return nil, fmt.Errorf("error uploading %s to S3: %w", outputKey, err)
		}

		// Generate a presigned URL for the uploaded file.
		href, err := utils.GetDownloadPresignedURL(s3Ctrl.S3Svc, bucket, outputKey, urlExpDay)
		if err != nil {
			return nil, fmt.Errorf("error generating presigned URL for %s: %w", outputKey, err)
		}

		// Append the presigned URL to the result slice.
		presignedUrlArr = append(presignedUrlArr, href)
	}

	return presignedUrlArr, nil
}

// validateInputs is used to validate input parameters and existence of g01 key in S3
func validateInputs(g01Key string, projection string, geoElement []string, bucket string, s3Ctrl utils.S3Controller) error {

	if g01Key == "" || projection == "" {
		return fmt.Errorf("`key` and `projection` must be provided")
	}

	if projection != "wktUSACEProj" && projection != "wktUSACEProjAlt" && projection != "WktUSACEProjFt37_5" {
		return fmt.Errorf("`projection` can only be `wktUSACEProj`, `wktUSACEProjAlt`, or `WktUSACEProjFt37_5`")
	}

	err := utils.ValidateElements(geoElement, utils.AllowedGeoElements)
	if err != nil {
		return fmt.Errorf("error returned while validating geoElements: %w", err)
	}

	if err := utils.EnsureExtension(g01Key, ".g01"); err != nil {
		return fmt.Errorf("invalid input file extension: %w", err)
	}

	exists, err := utils.KeyExists(s3Ctrl.S3Svc, bucket, g01Key)
	if err != nil {
		return fmt.Errorf("error returned when invoking KeyExists: %w", err)
	}
	if !exists {
		return fmt.Errorf("the provided object does not exist in the S3 bucket: %s", g01Key)
	}

	// If everything is valid, return an empty presignedUrlArr and nil error
	return nil
}
