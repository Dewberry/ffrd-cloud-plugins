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
		Required: []string{"url_exp_days", "g01_key", "projection", "bucket", "output_prefix", "geo_elements"},
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
	fs, err := utils.FileStoreInit(typedParams.Bucket)
	if err != nil {
		plug.Log.Fatal("Error Initiating FileStore: ", err.Error())
	}
	s3Ctrl, err := utils.SessionManager()
	if err != nil {
		plug.Log.Fatal("Error connecting to s3: ", err.Error())
	}

	results, err := mcats.ProcessGeometry(fs, s3Ctrl, typedParams.UrlExpDay, typedParams.G01key, typedParams.Projection, typedParams.Bucket, typedParams.OutputPrefix, typedParams.GeoElements)
	if err != nil {
		plug.Log.Fatal(err)
	}

	plug.PrintResults(results)
}
