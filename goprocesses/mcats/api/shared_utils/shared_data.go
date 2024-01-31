package shared_utils

const WktUSACEProjAlt string = `PROJCS["USA_Contiguous_Albers_Equal_Area_Conic_USGS_version",
								GEOGCS["NAD83",
									DATUM["North_American_Datum_1983",
										SPHEROID["GRS 1980",6378137,298.257222101,
											AUTHORITY["EPSG","7019"]],
										AUTHORITY["EPSG","6269"]],
									PRIMEM["Greenwich",0,
										AUTHORITY["EPSG","8901"]],
									UNIT["degree",0.0174532925199433,
										AUTHORITY["EPSG","9122"]],
									AUTHORITY["EPSG","4269"]],
								PROJECTION["Albers_Conic_Equal_Area"],
								PARAMETER["latitude_of_center",23],
								PARAMETER["longitude_of_center",-96],
								PARAMETER["standard_parallel_1",29.5],
								PARAMETER["standard_parallel_2",45.5],
								PARAMETER["false_easting",0],
								PARAMETER["false_northing",0],
								UNIT["metre",1,
									AUTHORITY["EPSG","9001"]],
								AXIS["Easting",EAST],
								AXIS["Northing",NORTH],
								AUTHORITY["EPSG","5070"]]`

const WktUSACEProj string = `PROJCRS["USA_Contiguous_Albers_Equal_Area_Conic_USGS_version",
							BASEGEOGCRS["NAD83",
								DATUM["North American Datum 1983",
									ELLIPSOID["GRS 1980",6378137,298.257222101,
										LENGTHUNIT["metre",1]],
									ID["EPSG",6269]],
								PRIMEM["Greenwich",0,
									ANGLEUNIT["Degree",0.0174532925199433]]],
							CONVERSION["unnamed",
								METHOD["Albers Equal Area",
									ID["EPSG",9822]],
								PARAMETER["Latitude of false origin",23,
									ANGLEUNIT["Degree",0.0174532925199433],
									ID["EPSG",8821]],
								PARAMETER["Longitude of false origin",-96,
									ANGLEUNIT["Degree",0.0174532925199433],
									ID["EPSG",8822]],
								PARAMETER["Latitude of 1st standard parallel",29.5,
									ANGLEUNIT["Degree",0.0174532925199433],
									ID["EPSG",8823]],
								PARAMETER["Latitude of 2nd standard parallel",45.5,
									ANGLEUNIT["Degree",0.0174532925199433],
									ID["EPSG",8824]],
								PARAMETER["Easting at false origin",0,
									LENGTHUNIT["US survey foot",0.304800609601219],
									ID["EPSG",8826]],
								PARAMETER["Northing at false origin",0,
									LENGTHUNIT["US survey foot",0.304800609601219],
									ID["EPSG",8827]]],
							CS[Cartesian,2],
								AXIS["(E)",east,
									ORDER[1],
									LENGTHUNIT["US survey foot",0.304800609601219,
										ID["EPSG",9003]]],
								AXIS["(N)",north,
									ORDER[2],
									LENGTHUNIT["US survey foot",0.304800609601219,
										ID["EPSG",9003]]]]`
const WktUSACEProjFt37_5 string = `PROJCS["USA_Contiguous_Albers_Equal_Area_Conic_USGS_version",
								GEOGCS["GCS_North_American_1983",
								DATUM["D_North_American_1983",
								SPHEROID["GRS_1980",6378137.0,298.257222101]],
								PRIMEM["Greenwich",0.0],UNIT["Degree",0.0174532925199433]],
								PROJECTION["Albers"],PARAMETER["False_Easting",0.0],
								PARAMETER["False_Northing",0.0],
								PARAMETER["Central_Meridian",-96.0],
								PARAMETER["Standard_Parallel_1",29.5],
								PARAMETER["Standard_Parallel_2",45.5],
								PARAMETER["Latitude_Of_Origin",37.5],
								UNIT["Foot_US",0.3048006096012192]]`
