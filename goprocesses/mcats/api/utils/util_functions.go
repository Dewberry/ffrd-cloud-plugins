package utils

import (
	"fmt"
	"strings"
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

	if urlExpDay, ok := params["urlExpDay"].(int); ok {
		typedParams.UrlExpDay = urlExpDay
	} else {
		errStrings = append(errStrings, "urlExpDay must be an integer")
	}

	if g01key, ok := params["g01key"].(string); ok {
		typedParams.G01key = g01key
	} else {
		errStrings = append(errStrings, "g01key must be a string")
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

	if outputPrefix, ok := params["outputPrefix"].(string); ok {
		typedParams.OutputPrefix = outputPrefix
	} else {
		errStrings = append(errStrings, "outputPrefix must be a string")
	}

	if geoElements, ok := params["geoElements"].([]string); ok {
		typedParams.GeoElements = geoElements
	} else {
		errStrings = append(errStrings, "geoElements must be a slice of strings")
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
