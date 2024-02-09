package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	plug "github.com/Dewberry/papigoplug/papigoplug"
)

type ParamTypes struct {
	UrlExpDay    int
	G01key       string
	Projection   string
	Bucket       string
	OutputPrefix string
	GeoElements  []string
}

// assertParams takes a map of parameters with interface{} values and asserts them to expected types.
func AssertParams(params map[string]interface{}) (ParamTypes, error) {
	var typedParams ParamTypes
	var errStrings []string

	if urlExpDayFloat, ok := params["url_exp_days"].(float64); ok {
		typedParams.UrlExpDay = int(urlExpDayFloat)
		plug.Log.Infof("urlExpDay cast from float: %v to int: %v", urlExpDayFloat, typedParams.UrlExpDay)
	} else {
		errStrings = append(errStrings, "url_exp_days must be a number")
	}

	if g01key, ok := params["g01_key"].(string); ok {
		typedParams.G01key = g01key
	} else {
		errStrings = append(errStrings, "g01_key must be a string")
	}

	if projection, ok := params["projection"].(string); ok {
		typedParams.Projection = projection

	} else {
		errStrings = append(errStrings, "projection must be a string")
	}

	if bucket, ok := params["bucket"].(string); ok {
		typedParams.Bucket = bucket
	} else {
		errStrings = append(errStrings, "bucket must be a string")
	}

	if outputPrefix, ok := params["output_prefix"].(string); ok {
		typedParams.OutputPrefix = outputPrefix
	} else {
		errStrings = append(errStrings, "output_prefix must be a string")
	}

	if ge, ok := params["geo_elements"].([]interface{}); ok {
		for _, elem := range ge {
			if str, ok := elem.(string); ok {
				typedParams.GeoElements = append(typedParams.GeoElements, str)
			} else {
				errStrings = append(errStrings, "each element in geo_elements must be a string")
			}
		}
	} else {
		errStrings = append(errStrings, "geo_elements must be an array of strings")
	}

	if len(errStrings) > 0 {
		return typedParams, fmt.Errorf("type assertion error(s): %s", strings.Join(errStrings, ", "))
	}

	return typedParams, nil
}

// returns true if a string is found in a string array
func StrArrContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// validateGeoElements checks if all elements in the input array are allowed.
func ValidateElements(elements []string, allowedElements map[string]bool) error {
	for _, elem := range elements {
		if _, exists := allowedElements[elem]; !exists {
			return fmt.Errorf("invalid geoElement '%s' provided; allowed elements are %v", elem, getKeysFromMap(allowedElements))
		}
	}
	return nil
}

// getKeysFromMap returns a slice of keys from the map
func getKeysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ensureG01Extension checks if the given filePath has a .g01 extension.
func EnsureExtension(key string, ext string) error {
	if filepath.Ext(key) != ext {
		return fmt.Errorf("file must have a %s extension, got: %s", ext, filepath.Ext(key))
	}
	return nil
}
