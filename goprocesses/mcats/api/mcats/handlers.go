package mcats

import (
	"app/ccmock"
	"app/controller"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// GeoRefreshHandler returns an echo.HandlerFunc that handles the refreshing of geospatial data cache.
// The function is responsible for validating the request, checking if the file exists in the S3 bucket,
// and invoking the appropriate function to refresh the geospatial data cache based on the endpoint.
// Supported endpoints are "/refresh_mesh_line" and "/refresh_break_line".
func GeoRefreshHandler(p *ccmock.FFRDProject, ctrl *controller.Controller) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Key        string `json:"key"`
			Projection string `json:"projection"`
		}

		// Decode the request body into the GeomRequest struct
		if err := c.Bind(&req); err != nil {
			errMsg := fmt.Errorf("invalid request body: %s", err.Error())
			log.Error(errMsg)
			return c.JSON(http.StatusBadRequest, errMsg.Error())
		}
		// Check if key is provided
		if req.Key == "" || req.Projection == "" {
			errMsg := fmt.Errorf("`key` and `projection` must be provided")
			log.Error(errMsg.Error())
			return c.JSON(http.StatusUnprocessableEntity, errMsg.Error())
		}

		if req.Projection != "wktUSACEProj" && req.Projection != "wktUSACEProjAlt" && req.Projection != "WktUSACEProjFt37_5" {
			errMsg := fmt.Errorf("`projection` can only be `wktUSACEProj` or `wktUSACEProjAlt` or `WktUSACEProjFt37_5`")
			log.Error(errMsg.Error())
			return c.JSON(http.StatusUnprocessableEntity, errMsg.Error())
		}

		// Ensure the file has a .g01 extension
		if err := ensureExtension(req.Key, ".g01"); err != nil {
			errMsg := fmt.Errorf("invalid file extension: %s", err.Error())
			log.Error(errMsg.Error())
			return c.JSON(http.StatusBadRequest, errMsg.Error())
		}

		// Check if the key exists in the S3 bucket
		ok, err := p.S3Ctrl.KeyExists(p.Bucket, req.Key)
		if err != nil {
			errMsg := fmt.Errorf("error returned when invoking KeyExists: %s", err.Error())
			log.Error(errMsg.Error())
			return c.JSON(http.StatusInternalServerError, errMsg.Error())
		}

		if !ok {
			errMsg := fmt.Errorf("the provided object does not exist in the S3 bucket: %s", req.Key)
			log.Error(errMsg.Error())
			return c.JSON(http.StatusBadRequest, errMsg.Error())
		}

		var geoElement string
		switch c.Path() {
		case "/mcat/refresh_mesh_line":
			geoElement = "mesh"
		case "/mcat/refresh_break_line":
			geoElement = "breakline"
		case "/mcat/refresh_twod_area":
			geoElement = "twodarea"
		default:
			errMsg := fmt.Errorf("invalid endpoint")
			log.Error(errMsg)
			return c.JSON(http.StatusUnprocessableEntity, errMsg.Error())
		}

		// Call refreshGeoCache with the appropriate parameters
		if err := refreshGeoCache(ctrl, req.Key, geoElement, req.Projection); err != nil {
			errMsg := fmt.Errorf("failed to refresh cache: %s", err.Error())
			log.Errorf(errMsg.Error())
			return c.JSON(http.StatusInternalServerError, errMsg.Error())
		}

		log.Infof("successfully refreshed cache")
		return c.JSON(http.StatusOK, "successfully refreshed cache")
	}
}

// GetGeoJSONHandler creates and returns an Echo HTTP handler function to retrieve
// geospatial data based on the requested path and provided file path. The data is
// retrieved from an S3 bucket, filtered, converted to GeoJSON, and returned to the
// client. Supported endpoints are "/get_geo_mesh_line" and "/get_geo_break_line".
func GetGeoJSONHandler(p *ccmock.FFRDProject, ctrl *controller.Controller) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req struct {
			Key        string `json:"key"`
			Projection string `json:"projection"`
		}
		// Decode the request body into the req struct
		if err := c.Bind(&req); err != nil {
			errMsg := fmt.Errorf("invalid request body: %s", err)
			log.Error(errMsg.Error())
			return c.JSON(http.StatusBadRequest, errMsg.Error())
		}

		// Check if key is provided
		if req.Key == "" || req.Projection == "" {
			errMsg := fmt.Errorf("`key` and `projection` must be provided")
			log.Error(errMsg.Error())
			return c.JSON(http.StatusUnprocessableEntity, errMsg.Error())
		}

		if req.Projection != "wktUSACEProj" && req.Projection != "wktUSACEProjAlt" && req.Projection != "WktUSACEProjFt37_5" {
			errMsg := fmt.Errorf("`projection` can only be `wktUSACEProj` or `wktUSACEProjAlt` or `WktUSACEProjFt37_5`")
			log.Error(errMsg.Error())
			return c.JSON(http.StatusUnprocessableEntity, errMsg.Error())
		}

		// Ensure the file has a .g01 extension
		if err := ensureExtension(req.Key, ".g01"); err != nil {
			errMsg := fmt.Errorf("invalid file extension: %s", err.Error())
			log.Error(errMsg.Error())
			return c.JSON(http.StatusBadRequest, errMsg.Error())
		}

		// Check if the key exists in the S3 bucket
		exists, err := p.S3Ctrl.KeyExists(p.Bucket, req.Key)
		if err != nil {
			errMsg := fmt.Errorf("Error returned when invoking KeyExists: %s", err.Error())
			log.Error(errMsg.Error())
			return c.JSON(http.StatusInternalServerError, errMsg.Error())
		}

		if !exists {
			errMsg := fmt.Errorf("The provided object does not exist in the S3 bucket: %s", req.Key)
			log.Error(errMsg.Error())
			return c.JSON(http.StatusNotFound, errMsg.Error())
		}

		var geoElement string
		switch c.Path() {
		case "/mcat/break_line":
			geoElement = "mesh"
		case "/mcat/mesh_line":
			geoElement = "breakline"
		default:
			errMsg := "invalid endpoint"
			log.Errorf(errMsg)
			return c.JSON(http.StatusNotFound, errMsg)
		}

		// Retrieve and filter Geospatial Data
		specificFeatures, err := getFilteredGeoData(ctrl, req.Key, geoElement, req.Projection)
		if err != nil {
			errMsg := fmt.Errorf("Error while getting the geo data: %s", err)
			log.Error(errMsg.Error())
			return c.JSON(http.StatusInternalServerError, errMsg.Error())
		}

		collection, err := convertToGeoJSON(specificFeatures)
		if err != nil {
			errMsg := fmt.Errorf("Error while converting the geo data to GeoJSON: %s", err)
			log.Error(errMsg.Error())
			return c.JSON(http.StatusInternalServerError, errMsg.Error())
		}

		return c.JSON(http.StatusOK, collection)
	}
}
