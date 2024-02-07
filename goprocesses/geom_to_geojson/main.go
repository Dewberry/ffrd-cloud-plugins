package main

import (
	"fmt"
	"os"

	"app/mcats"
	"app/utils"

	plug "github.com/Dewberry/papigoplug/papigoplug"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load("../.env")
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
	plug.InitLog("info")
	allowedParams := plug.PluginParams{
		Required: []string{"urlExpDay", "g01Key", "projection", "bucket", "outputPrefix", "geoElements"},
	}
	params, err := plug.ParseInput(os.Args, allowedParams)
	if err != nil {
		plug.Log.Fatal(err)
	}
	plug.Log.Infof("Params provided: %s", params)
	typedParams, err := utils.AssertParams(params)
	if err != nil {
		plug.Log.Fatal(err)
	}
	fs := utils.FileStoreInit(typedParams.Bucket)
	s3Ctrl, err := utils.SessionManager()
	if err != nil {
		plug.Log.Fatal("Error connecting to s3: ", err.Error())
	}

	// Now you can use typedParams with the correct types
	hrefs, err := mcats.GetGeoJsonPresignedUrls(fs, s3Ctrl, typedParams.UrlExpDay, typedParams.G01key, typedParams.Projection, typedParams.Bucket, typedParams.OutputPrefix, typedParams.GeoElements)
	if err != nil {
		plug.Log.Fatal(err)
	}
	hrefsMap := make(map[string]interface{})
	hrefsMap["results"] = hrefs
	defer plug.PrintResults(hrefsMap)

}
