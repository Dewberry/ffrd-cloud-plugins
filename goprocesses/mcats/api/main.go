package main

import (
	"fmt"
	"os"

	"app/mcats"
	"app/utils"

	plug "github.com/Dewberry/papigoplug/papigoplug"
)

func init() {
	requiredEnvVars := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_REGION",
	}

	for _, envVar := range requiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			fmt.Printf("Error: Missing environment variable %s\n", envVar)
			os.Exit(1)
		}
	}
}

func main() {
	//s3Conf := utils.NewS3Conf("AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "AWS_S3_BUCKET")
	//s3fs := s3Conf.Init()
	plug.InitLog("info")
	allowedParams := plug.PluginParams{
		Required: []string{"urlExpDay", "g01key", "projection", "bucket", "outputPrefix", "geoElements"},
	}
	fs := utils.FileStoreInit("S3")
	s3Ctrl, err := utils.SessionManager()
	if err != nil {
		plug.Log.Panic("Error connecting to s3: ", err.Error())
	}
	params, err := plug.ParseInput(os.Args, allowedParams)
	if err != nil {
		plug.Log.Panic(err)
	}
	plug.Log.Infof("Params provided: %s", params) // params[""]
	typedParams, err := utils.AssertParams(params)
	if err != nil {
		plug.Log.Panic(err)
	}

	// Now you can use typedParams with the correct types
	hrefs, err := mcats.GetGeoJsonPresignedUrls(fs, s3Ctrl, typedParams.UrlExpDay, typedParams.G01key, typedParams.Projection, typedParams.Bucket, typedParams.OutputPrefix, typedParams.GeoElements)
	if err != nil {
		plug.Log.Panic(err)
	}
	hrefsMap := make(map[string]interface{})
	hrefsMap["results"] = hrefs
	defer plug.PrintResults(hrefsMap)

}
