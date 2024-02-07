package mcats

import (
	"app/utils"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/Dewberry/mcat-ras/tools"
	"github.com/USACE/filestore"
)

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

func GetGeoJsonPresignedUrls(fs *filestore.FileStore, s3Ctrl utils.S3Controller, urlExpDay int, g01Key, projection, bucket, outPutPrefix string, geoElement []string) ([]string, error) {
	var presignedUrlArr []string
	// Check if key and projection are provided
	err := validateInputs(g01Key, projection, geoElement, bucket, s3Ctrl)
	if err != nil {
		errMsg := fmt.Errorf("error while validating input parameter: %s", err.Error())
		return presignedUrlArr, errMsg
	}

	//Retrieve and filter Geospatial Data
	specificFeatures, err := getFilteredGeoData(fs, g01Key, projection, geoElement)
	if err != nil {
		errMsg := fmt.Errorf("error while getting the geo data: %s", err)
		return presignedUrlArr, errMsg
	}

	//convert geospatial data to geojson format
	collections, err := convertToGeoJSON(specificFeatures, projection)
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
	presignedUrlArr, err = uploadGeoJSONToS3AndGeneratePresignedURLs(collectionJson, g01Key, outPutPrefix, bucket, urlExpDay, s3Ctrl)

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
				if fieldName == "Mesh" {
					for _, feature := range features.Mesh {
						var tempArr []tools.VectorFeature
						tempArr = append(tempArr, feature)
						specificFeatures[feature.FeatureName] = tempArr
					}
				} else {
					specificFeatures[fieldName] = field.Interface()
				}

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
