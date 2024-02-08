package mcats

import (
	"app/utils"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/Dewberry/mcat-ras/tools"
	plug "github.com/Dewberry/papigoplug/papigoplug"
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

// ProcessGeometry will extract geometry from a g01 file transform it to a geojson and upload the geojson to S3.
// this function will return the presigned URLS of the uploaded geojsons in S3
func ProcessGeometry(fs *filestore.FileStore, s3Ctrl utils.S3Controller, urlExpDay int, g01Key, projection, bucket, outPutPrefix string, geoElement []string) (map[string]interface{}, error) {
	finalOutput := make(map[string]interface{})
	//Check if key and projection are provided
	plug.Log.Infof("validating user input")
	err := validateInputs(g01Key, projection, geoElement, bucket, s3Ctrl)
	if err != nil {
		errMsg := fmt.Errorf("error while validating input parameter: %w", err)
		return finalOutput, errMsg
	}

	plug.Log.Infof("Retrieving and filtering geo data")
	//Retrieve and filter Geospatial Data
	specificFeatures, err := getFilteredGeoData(fs, g01Key, projection, geoElement)
	if err != nil {
		errMsg := fmt.Errorf("error while getting the geo data: %w", err)
		return finalOutput, errMsg
	}
	plug.Log.Infof("converting geo data to geojson format")
	//convert geospatial data to geojson format
	collections, err := convertToGeoJSON(specificFeatures, projection)
	if err != nil {
		errMsg := fmt.Errorf("error while converting the geo data to GeoJSON: %w", err)
		return finalOutput, errMsg
	}
	collectionJson := make(map[string][]byte)
	//convert geojson struct to json
	for key, value := range collections {
		json, err := json.Marshal(value)
		if err != nil {
			errMsg := fmt.Errorf("error while marshalling geojson struct to json: %w", err)
			return finalOutput, errMsg
		}
		collectionJson[key] = json
	}
	plug.Log.Infof("uploading .geoJson file(s) to S3 bucket %s", bucket)
	outPutPrefix = strings.TrimSuffix(outPutPrefix, "/")
	uploadResults, err := uploadGeoJSONToS3AndGeneratePresignedURLs(collectionJson, g01Key, outPutPrefix, bucket, urlExpDay, s3Ctrl)
	if err != nil {
		return nil, err
	}

	links := make([]interface{}, 0)
	results := make([]interface{}, 0)

	for _, result := range uploadResults {
		links = append(links, map[string]interface{}{
			"Href":  result.PresignedURL,
			"rel":   "presigned-url",
			"title": result.S3URI,
			"type":  "application/octet-stream",
		})

		results = append(results, map[string]interface{}{
			"href":  result.S3URI,
			"title": result.Title,
		})
	}

	finalOutput = map[string]interface{}{
		"links":   links,
		"results": results,
	}

	return finalOutput, nil
}

// getFilteredGeoData retrieves geospatial data for a given filePath and filters it based on the geoElement provided.
// The function takes a controller object, the filePath of the geospatial data, and an array of strings indicating the type of
// geographic element.
func getFilteredGeoData(fs *filestore.FileStore, g01Key, projection string, geoElements []string) (map[string]interface{}, error) {
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

	plug.Log.Infof("extracting geo spatial data from %s", g01Key)
	err := tools.GetGeospatialData(&gd, *fs, g01Key, projection, 4326)
	if err != nil {
		return nil, fmt.Errorf("error in GetGeospatialData for %s: %w", g01Key, err)
	}

	// Filter features
	specificFeatures := make(map[string]interface{})

	features := gd.Features[path.Base(g01Key)]

	if utils.StrArrContains(geoElements, "all") {
		plug.Log.Infof("appending all data from g01 file and not filtering")
		v := reflect.ValueOf(features)
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldName := v.Type().Field(i).Name
			if field.Kind() == reflect.Slice && field.Len() > 0 {
				// if fieldName == "Mesh" {
				// 	for _, feature := range features.Mesh {
				// 		var tempArr []tools.VectorFeature
				// 		tempArr = append(tempArr, feature)
				// 		specificFeatures[feature.FeatureName] = tempArr
				// 	}
				// } else {
				specificFeatures[fieldName] = field.Interface()
				// }

			}
		}
		return specificFeatures, nil
	}
	// Extract the right features based on geoElement

	for _, geoElement := range geoElements {
		plug.Log.Infof("filtering and appending %s element", geoElement)
		switch geoElement {
		case "breakline":
			specificFeatures[geoElement] = features.BreakLines
		// case "mesh":
		// 	for _, feature := range features.Mesh {
		// 		var tempArr []tools.VectorFeature
		// 		tempArr = append(tempArr, feature)
		// 		specificFeatures[feature.FeatureName] = tempArr
		// 	}
		case "twodarea":
			specificFeatures[geoElement] = features.TwoDAreas
		default:
			return nil, fmt.Errorf("Invalid geoElement provided: %s", geoElement)
		}

	}

	return specificFeatures, nil
}
