package mcats

import (
	"encoding/json"
	"fmt"

	"github.com/Dewberry/mcat-ras/tools"
	"github.com/dewberry/gdal"
)

// convertToGeoJSON converts a slice of VectorFeature objects into a GeoJSON feature collection.
// The function takes a slice of VectorFeature objects as input and returns a GeoJSON Collection
// or an error if the conversion fails.
func convertToGeoJSON(features map[string]interface{}, projection string) (map[string]Collection, error) {
	collections := make(map[string]Collection) // Initialize the map

	for key, value := range features {
		var geoJSONFeatures []Feature
		if slice, ok := value.([]tools.VectorFeature); ok {
			for _, feature := range slice {
				// Assume that feature.Geometry is already in a format that can be included in a GeoJSON feature
				geometry, err := convertWKBToGeoJSON(feature.Geometry, projection)
				if err != nil {
					return nil, fmt.Errorf("error converting geometry: %w", err)
				}

				geoJSONFeature := Feature{
					Type:       "Feature",
					Properties: map[string]interface{}{"Name": feature.FeatureName},
					Geometry:   geometry,
				}
				geoJSONFeatures = append(geoJSONFeatures, geoJSONFeature)
			}
		}
		collections[key] = Collection{
			Type:     "FeatureCollection",
			Features: geoJSONFeatures,
		}
	}
	return collections, nil
}

// ConvertWKBToGeoJSON converts Well-Known Binary (WKB) geometry data to GeoJSON format.
// It takes a slice of uint8 representing the WKB data as input and returns a Geometry object
// representing the corresponding GeoJSON data using the gdal functions.
func convertWKBToGeoJSON(wkb []uint8, projection string) (Geometry, error) {
	srs := gdal.CreateSpatialReference(projection)

	geom, err := gdal.CreateFromWKB(wkb, srs, len(wkb))
	if err != nil {
		return Geometry{}, fmt.Errorf("error creating a geometry object from its WKB:  %w", err)
	}

	var geojson Geometry
	err = json.Unmarshal([]byte(geom.ToJSON()), &geojson)
	if err != nil {
		return Geometry{}, fmt.Errorf("error unmarshalling geometry to JSON:  %w", err)
	}
	return geojson, nil
}
