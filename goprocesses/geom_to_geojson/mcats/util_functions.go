package mcats

import (
	"app/utils"
	"fmt"
	"path/filepath"
	"strings"
)

type FinalOutput struct {
	Links   []Link   `json:"links"`
	Results []Result `json:"results"`
}

type Link struct {
	Href  string `json:"Href"`
	Rel   string `json:"rel"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

type Result struct {
	Href  string `json:"href"`
	Title string `json:"title"`
}

type UploadResult struct {
	PresignedURL string
	S3URI        string
	Title        string
}

// UploadGeoJSONToS3AndGeneratePresignedURLs uploads GeoJSON data to S3 and generates presigned URLs.
// It returns a slice of presigned URLs and any error encountered.
func uploadGeoJSONToS3AndGeneratePresignedURLs(collectionJson map[string][]byte, g01Key, outPutPrefix, bucket string, urlExpDay int, s3Ctrl utils.S3Controller) ([]UploadResult, error) {
	var results []UploadResult

	for key, value := range collectionJson {
		g01FileName := strings.TrimSuffix(filepath.Base(g01Key), filepath.Ext(filepath.Base(g01Key)))
		outputKey := fmt.Sprintf("%s/%s_%s.geojson", outPutPrefix, g01FileName, key)

		// Use the UploadToS3 function for uploading
		s3URI, err := utils.UploadToS3(s3Ctrl.Sess, bucket, outputKey, value, "application/octet-stream")
		if err != nil {
			return nil, fmt.Errorf("error uploading %s to S3: %w", outputKey, err)
		}

		// Generate the presigned URL for the uploaded file
		href, err := utils.GetDownloadPresignedURL(s3Ctrl.S3Svc, bucket, outputKey, urlExpDay)
		if err != nil {
			return nil, fmt.Errorf("error generating presigned URL for %s: %w", outputKey, err)
		}

		title := g01FileName + " " + key

		results = append(results, UploadResult{
			PresignedURL: href,
			S3URI:        s3URI,
			Title:        title,
		})
	}

	return results, nil
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

	return nil
}
