package mcats

import (
	"app/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Dewberry/mcat-ras/tools"
	"github.com/USACE/filestore"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dewberry/gdal"
	"github.com/labstack/echo/v4"
)

type RefLayer struct {
	Type       string     `json:"type"`
	Collection Collection `json:"collection"`
}

type Collection struct {
	Type     string    `json:"type" default:"FeatureCollection"`
	Features []Feature `json:"features,omitempty"`
}
type Feature struct {
	Type       string                 `json:"type,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Geometry   Geometry               `json:"geometry,omitempty"`
}
type Geometry struct {
	Type        string        `json:"type"`
	Coordinates []interface{} `json:"coordinates"`
}
type GeomRequest struct {
	Key         string `json:"key" db:"s3_key"`
	FeatureName string `json:"feature_name" db:"feature_name"`
	Geom        string `json:"geom" db:"geom"`
}

func Handler(fs *filestore.FileStore, s3Ctrl utils.S3Controller) echo.HandlerFunc {
	return func(c echo.Context) error {
		var geoElements []string
		geoElements = append(geoElements, "all")
		href, err := GetGeoJsonPresignedUrls(fs, s3Ctrl, 7, "model-library/FFRD_Kanawha_Compute/ras/Bluestone Local/Kanawha_0505_Bluest.g01", "wktUSACEProj", "ffrd-pilot", "logs/anton", geoElements)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, fmt.Sprintf("error: %s", err.Error()))
		}
		return c.JSON(http.StatusOK, href)
	}
}

func GetGeoJsonPresignedUrls(fs *filestore.FileStore, s3Ctrl utils.S3Controller, urlExpDay int, g01Key, projection, bucket, outPutPrefix string, geoElement []string) ([]string, error) {
	var presignedUrlArr []string
	// Check if key and projection are provided
	if g01Key == "" || projection == "" {
		errMsg := fmt.Errorf("`key` and `projection` must be provided")
		return presignedUrlArr, errMsg
	}

	if projection != "wktUSACEProj" && projection != "wktUSACEProjAlt" && projection != "WktUSACEProjFt37_5" {
		errMsg := fmt.Errorf("`projection` can only be `wktUSACEProj` or `wktUSACEProjAlt` or `WktUSACEProjFt37_5`")
		return presignedUrlArr, errMsg
	}

	//TODO: enforce what kind of types are allowed in teh array and enforce
	// if geoElement != "mesh" && geoElement != "breakline" && geoElement != "twodarea" {
	// 	errMsg := fmt.Errorf("`geoElement` can only be `mesh` or `breakline` or `twodarea`")
	// 	return "", errMsg
	// }

	// Ensure the file has a .g01 extension
	if err := ensureExtension(g01Key, ".g01"); err != nil {
		errMsg := fmt.Errorf("invalid input file extension: %s", err.Error())
		return presignedUrlArr, errMsg
	}

	// // Ensure the output file name had has a .geojson extension
	// if err := ensureExtension(outputGeoJsonName, ".geojson"); err != nil {
	// 	errMsg := fmt.Errorf("invalid output file extension: %s", err.Error())
	// 	return "", errMsg
	// }

	// Check if the key exists in the S3 bucket
	exists, err := utils.KeyExists(s3Ctrl.S3Svc, bucket, g01Key)
	if err != nil {
		errMsg := fmt.Errorf("error returned when invoking KeyExists: %s", err.Error())
		return presignedUrlArr, errMsg
	}

	if !exists {
		errMsg := fmt.Errorf("the provided object does not exist in the S3 bucket: %s", g01Key)
		return presignedUrlArr, errMsg
	}

	// Retrieve and filter Geospatial Data
	specificFeatures, err := getFilteredGeoData(fs, g01Key, projection, geoElement)
	if err != nil {
		errMsg := fmt.Errorf("error while getting the geo data: %s", err)
		return presignedUrlArr, errMsg
	}

	//convert geospatial data to geojson format
	collections, err := convertToGeoJSON(specificFeatures)
	if err != nil {
		errMsg := fmt.Errorf("error while converting the geo data to GeoJSON: %s", err)
		return presignedUrlArr, errMsg
	}
	collectionJson := make(map[string][]byte)
	//convert geojson struct to json
	for key, value := range collections {
		json, err := json.Marshal(value)
		if err != nil {
			errMsg := fmt.Errorf("error while marshalling geojson struct to json: %s", err)
			return presignedUrlArr, errMsg
		}
		collectionJson[key] = json
	}
	outPutPrefix = strings.TrimSuffix(outPutPrefix, "/")
	for key, value := range collectionJson {
		g01FileName := strings.TrimSuffix(filepath.Base(g01Key), filepath.Ext(filepath.Base(g01Key)))
		outputKey := fmt.Sprintf("%s/%s_%s.geojson", outPutPrefix, g01FileName, key)
		uploader := s3manager.NewUploader(s3Ctrl.Sess)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(outputKey),
			Body:        bytes.NewReader([]byte(value)),
			ContentType: aws.String("binary/octet-stream"),
		})
		if err != nil {
			errMsg := fmt.Errorf("error uploading %s to S3: %s", outputKey, err.Error())
			return presignedUrlArr, errMsg
		}
		href, err := utils.GetDownloadPresignedURL(s3Ctrl.S3Svc, bucket, outputKey, urlExpDay)
		if err != nil {
			errMsg := fmt.Errorf("error generating presigned URL for %s: %s", outputKey, err)
			return presignedUrlArr, errMsg
		}
		presignedUrlArr = append(presignedUrlArr, href)
	}

	return presignedUrlArr, err
}

// getFilteredGeoData retrieves geospatial data for a given filePath and filters it based on the geoElement provided.
// The function takes a controller object, the filePath of the geospatial data, and a string indicating the type of
// geographic element (either "mesh" or "breakline").
func getFilteredGeoData(fs *filestore.FileStore, filePath, projection string, geoElements []string) (map[string]interface{}, error) {
	gd := tools.GeoData{
		Features: make(map[string]tools.Features),
	}
	if projection == "wktUSACEProj" {
		projection = utils.WktUSACEProj
	} else if projection == "wktUSACEProjAlt" {
		projection = utils.WktUSACEProjAlt
	} else if projection == "WktUSACEProjFt37_5" {
		projection = utils.WktUSACEProjFt37_5
	}
	err := tools.GetGeospatialData(&gd, *fs, filePath, projection, 4326)
	if err != nil {
		return nil, fmt.Errorf("error in GetGeospatialData for %s: %s", filePath, err.Error())
	}

	// Filter features
	specificFeatures := make(map[string]interface{})

	features := gd.Features[path.Base(filePath)]

	if utils.StrArrContains(geoElements, "all") {
		v := reflect.ValueOf(features)
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldName := v.Type().Field(i).Name
			if field.Kind() == reflect.Slice && field.Len() > 0 {
				specificFeatures[fieldName] = field.Interface()
			}
		}
		return specificFeatures, nil
	}
	// Extract the right features based on geoElement

	for _, geoElement := range geoElements {
		switch geoElement {
		case "breakline":
			specificFeatures[geoElement] = features.BreakLines
		case "mesh":
			for _, feature := range features.Mesh {
				var tempArr []tools.VectorFeature
				tempArr = append(tempArr, feature)
				specificFeatures[feature.FeatureName] = tempArr
			}
		case "twodarea":
			specificFeatures[geoElement] = features.TwoDAreas
		default:
			return nil, fmt.Errorf("Invalid geoElement provided: %s", geoElement)
		}

	}

	return specificFeatures, nil
}

// convertToGeoJSON converts a slice of VectorFeature objects into a GeoJSON feature collection.
// The function takes a slice of VectorFeature objects as input and returns a GeoJSON Collection
// or an error if the conversion fails.
func convertToGeoJSON(features map[string]interface{}) (map[string]Collection, error) {
	collections := make(map[string]Collection) // Initialize the map

	for key, value := range features {
		var geoJSONFeatures []Feature
		if slice, ok := value.([]tools.VectorFeature); ok {
			for _, feature := range slice {
				// Assume that feature.Geometry is already in a format that can be included in a GeoJSON feature
				geometry, err := ConvertWKBToGeoJSON(feature.Geometry)
				if err != nil {
					return nil, fmt.Errorf("error converting geometry: %s", err.Error())
				}

				geoJSONFeature := Feature{
					Type:       "Feature",
					Properties: map[string]interface{}{"Name": feature.FeatureName},
					Geometry:   geometry,
				}
				geoJSONFeatures = append(geoJSONFeatures, geoJSONFeature)
			}
		}
		collections[key] = Collection{
			Type:     "FeatureCollection",
			Features: geoJSONFeatures,
		}
	}
	return collections, nil
}

// ConvertWKBToGeoJSON converts Well-Known Binary (WKB) geometry data to GeoJSON format.
// It takes a slice of uint8 representing the WKB data as input and returns a Geometry object
// representing the corresponding GeoJSON data using the gdal functions.
func ConvertWKBToGeoJSON(wkb []uint8) (Geometry, error) {
	srs := gdal.CreateSpatialReference(utils.WktUSACEProj)

	err := srs.FromEPSG(4326)
	if err != nil {
		return Geometry{}, fmt.Errorf("error initializing SRS based on EPSG code: %s", err.Error())
	}

	geom, err := gdal.CreateFromWKB(wkb, srs, len(wkb))
	if err != nil {
		return Geometry{}, fmt.Errorf("error creating a geometry object from its WKB: %s", err.Error())
	}

	var geojson Geometry
	err = json.Unmarshal([]byte(geom.ToJSON()), &geojson)
	if err != nil {
		return Geometry{}, fmt.Errorf("error unmarshalling geometry to JSON: %s", err.Error())
	}
	return geojson, nil
}

// ensureG01Extension checks if the given filePath has a .g01 extension.
func ensureExtension(key string, ext string) error {
	if filepath.Ext(key) != ext {
		return fmt.Errorf("file must have a %s extension, got: %s", ext, filepath.Ext(key))
	}
	return nil
}
