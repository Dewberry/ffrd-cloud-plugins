package utils

import (
	"fmt"
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

	if urlExpDayFloat, ok := params["urlExpDay"].(float64); ok {
		typedParams.UrlExpDay = int(urlExpDayFloat)
		plug.Log.Infof("urlExpDay cast from float: %v to int: %v", urlExpDayFloat, typedParams.UrlExpDay)
	} else {
		errStrings = append(errStrings, "urlExpDay must be a number")
	}

	if g01key, ok := params["g01Key"].(string); ok {
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

	if ge, ok := params["geoElements"].([]interface{}); ok {
		for _, elem := range ge {
			if str, ok := elem.(string); ok {
				typedParams.GeoElements = append(typedParams.GeoElements, str)
			} else {
				errStrings = append(errStrings, "each element in geoElements must be a string")
			}
		}
	} else {
		errStrings = append(errStrings, "geoElements must be an array of strings")
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
