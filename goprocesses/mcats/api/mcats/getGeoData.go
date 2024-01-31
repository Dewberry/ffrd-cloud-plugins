package mcats

import (
	"app/ccmock"
	"app/controller"
	sharedutils "app/shared_utils"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"

	"github.com/Dewberry/mcat-ras/tools"
	"github.com/dewberry/gdal"
	log "github.com/sirupsen/logrus"
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

func RefreshGeoDataOnStartUp(ctrl *controller.Controller, p *ccmock.FFRDProject) {
	for _, model := range p.Models {
		// Call refreshGeoCache for 'mesh'
		if err := refreshGeoCache(ctrl, model.GeometryFile, "breakline", model.Projection); err != nil {
			errMsg := fmt.Errorf("failed to refresh mesh cache for %s: %s", model.Name, err.Error())
			log.Error(errMsg)
		}

		// Call refreshGeoCache for 'breakline'
		if err := refreshGeoCache(ctrl, model.GeometryFile, "mesh", model.Projection); err != nil {
			errMsg := fmt.Errorf("failed to refresh breakline cache for %s: %s", model.Name, err.Error())
			log.Error(errMsg)
		}

		// Call refreshGeoCache for 'ras 2d domains'
		if err := refreshGeoCache(ctrl, model.GeometryFile, "twodarea", model.Projection); err != nil {
			errMsg := fmt.Errorf("failed to refresh twodarea cache for %s: %s", model.Name, err.Error())
			log.Error(errMsg)
		}
	}
}

// getFilteredGeoData retrieves geospatial data for a given filePath and filters it based on the geoElement provided.
// The function takes a controller object, the filePath of the geospatial data, and a string indicating the type of
// geographic element (either "mesh" or "breakline").
func getFilteredGeoData(ctrl *controller.Controller, filePath, geoElement, projection string) ([]tools.VectorFeature, error) {
	gd := tools.GeoData{
		Features: make(map[string]tools.Features),
	}
	if projection == "wktUSACEProj" {
		projection = sharedutils.WktUSACEProj
	} else if projection == "wktUSACEProjAlt" {
		projection = sharedutils.WktUSACEProjAlt
	} else if projection == "WktUSACEProjFt37_5" {
		projection = sharedutils.WktUSACEProjFt37_5
	}
	// err := tools.GetGeospatialData(&gd, ctrl.S3FS, filePath, projection, 4326)
	// if err != nil {
	// 	return nil, fmt.Errorf("error in GetGeospatialData for %s: %s", geoElement, err.Error())
	// }

	// Filter features
	features := gd.Features[path.Base(filePath)]
	// Extract the right features based on geoElement
	var specificFeatures []tools.VectorFeature
	switch geoElement {
	case "breakline":
		specificFeatures = features.BreakLines
	case "mesh":
		specificFeatures = features.Mesh
	case "twodarea":
		specificFeatures = features.TwoDAreas
	default:
		return nil, fmt.Errorf("Invalid geoElement provided: %s", geoElement)
	}

	return specificFeatures, nil
}

// refreshGeoCache processes and inserts geospatial data features into the database.
// It takes a controller object, a g01 filepath string, and a string indicating the type
// of geographic element (either "mesh" or "breakline").
//
// For "breaklines", all models are accepted by default. For "mesh", only "mesh_voronoi"
// models are currently supported. To extend support for additional mesh models, separate
// database tables with matching geometry types need to be created. For example, if a new mesh
// model involves point geometry, then a new table with geometry type "point" should be added.
func refreshGeoCache(ctrl *controller.Controller, filePath, geoElement, projection string) error {
	var procedureName string

	// Determine which stored procedure to use based on geoElement
	switch geoElement {
	case "breakline":
		procedureName = "kanawha.insert_break_line"
	case "mesh":
		procedureName = "kanawha.insert_mesh_line"
	case "twodarea":
		procedureName = "kanawha.insert_twod_area"
	default:
		return fmt.Errorf("Invalid geoElement provided: %s", geoElement)
	}

	// Retrieve and filter Geospatial Data
	specificFeatures, err := getFilteredGeoData(ctrl, filePath, geoElement, projection)
	if err != nil {
		log.Errorf("error calling getFilteredGeoData: %s", err.Error())
		return fmt.Errorf("error calling getFilteredGeoData: %s", err.Error())
	}

	// Insert into the appropriate table using the provided procedure
	for _, feature := range specificFeatures {
		if geoElement == "mesh" && feature.FeatureName != "mesh_voronoi" {
			continue
		}
		_, dbErr := ctrl.DB.Exec(fmt.Sprintf(`CALL %s($1, $2, ST_SetSRID(ST_GeomFromWKB($3), 4326))`, procedureName), filePath, feature.FeatureName, feature.Geometry)
		if dbErr != nil {
			log.Errorf("refreshGeoCache for %s: Error calling %s procedure for feature %s: %s", geoElement, procedureName, feature.FeatureName, dbErr.Error())
			return fmt.Errorf("error calling %s procedure for feature %s: %s", procedureName, feature.FeatureName, dbErr.Error())
		}
	}

	log.Infof("refreshGeoCache successfully processed and inserted %s features from %s", geoElement, filePath)
	return nil
}

// convertToGeoJSON converts a slice of VectorFeature objects into a GeoJSON feature collection.
// The function takes a slice of VectorFeature objects as input and returns a GeoJSON Collection
// or an error if the conversion fails.
func convertToGeoJSON(features []tools.VectorFeature) (Collection, error) {
	var geoJSONFeatures []Feature

	for _, feature := range features {
		// Assume that feature.Geometry is already in a format that can be included in a GeoJSON feature
		geometry, err := ConvertWKBToGeoJSON(feature.Geometry)
		if err != nil {
			return Collection{}, fmt.Errorf("Error converting geometry: %s", err.Error())
		}

		geoJSONFeature := Feature{
			Type: "Feature",
			Properties: map[string]interface{}{
				"Name": feature.FeatureName,
			},
			Geometry: geometry,
		}
		geoJSONFeatures = append(geoJSONFeatures, geoJSONFeature)
	}

	return Collection{
		Type:     "FeatureCollection",
		Features: geoJSONFeatures,
	}, nil
}

// ConvertWKBToGeoJSON converts Well-Known Binary (WKB) geometry data to GeoJSON format.
// It takes a slice of uint8 representing the WKB data as input and returns a Geometry object
// representing the corresponding GeoJSON data using the gdal functions.
func ConvertWKBToGeoJSON(wkb []uint8) (Geometry, error) {
	srs := gdal.CreateSpatialReference(sharedutils.WktUSACEProj)

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
