import datetime
import logging

import boto3
import geopandas as gpd
import imageio
import matplotlib.pyplot as plt
import numpy as np
import xarray as xr
from dotenv import find_dotenv, load_dotenv
from matplotlib.cm import Spectral_r
from shapely.geometry.base import BaseGeometry

MM_TO_IN = 25.4

load_dotenv(find_dotenv())
session = boto3.session.Session()
s3_client = session.client("s3")

logging.root.handlers = []
logging.basicConfig(
    level=logging.INFO,
    format="""{"time": "%(asctime)s" , "level": "%(levelname)s", "message": "%(message)s"}""",
    handlers=[logging.StreamHandler()],
)


def generate_zarr_file_paths(start: str, duration: int, base_url: str) -> list:
    """Generates zarr file paths for precip data extraction"""

    date_fmt = "%Y/%Y%m%d%H"
    # Attempt to parse the 'start' string
    try:
        start_dt = datetime.datetime.strptime(start, "%Y-%m-%d")
    except ValueError:
        raise ValueError("The 'start' parameter should be in the 'YYYY-MM-DD' format.")
    start_dt = datetime.datetime.strptime(start, "%Y-%m-%d")
    end_dt = start_dt + datetime.timedelta(hours=duration)

    zarrfiles = []
    while start_dt < end_dt:
        zarr_path = base_url + start_dt.strftime(date_fmt) + ".zarr"
        zarrfiles.append(zarr_path)
        start_dt += datetime.timedelta(hours=1)

    return zarrfiles


def fetch_watershed_geom(
    watershed_file: str,
    combine_all_geoms: bool = False,
) -> BaseGeometry:
    """Fetches geometry object from watershed geojson"""

    logging.info("Getting watershed geometry")
    watershed_geom = gpd.read_file(watershed_file, driver="GeoJSON")
    # Check that the geodataframe isn't empty
    if watershed_geom.empty:
        logging.error("Geodataframe is empty")
        raise ValueError("Geodataframe is empty")
    watershed_geom = watershed_geom.explode()
    # Check that the geodataframe contains valid geometries
    if not watershed_geom.geometry.is_valid.all():
        logging.error("Invalid geometries present.")
        raise ValueError("Invalid geometries present in the geodataframe.")

    if combine_all_geoms:
        geom = watershed_geom.geometry.unary_union
    else:
        geom = watershed_geom.loc[0].geometry

    return geom


def read_precip_dataset(
    zarrfiles: list,
    clip_poly: any,
    clip_to_watershed: bool = False,
    convert_to_inches: bool = True,
    mm_to_in: float = MM_TO_IN,
) -> xr.Dataset:
    """Reads in precip dataset from zarr filepaths and processes dataset accordingly"""

    ds = xr.open_mfdataset(zarrfiles, engine="zarr", consolidated=True)
    ds = ds.rio.write_crs(4326, inplace=True)
    if clip_to_watershed:
        original_shape = ds.dims
        ds = ds.rio.clip(clip_poly, drop=True, all_touched=False)
        if ds.dims == original_shape:
            logging.error("Dataset shape didnt change after clipping to the watershed.")
    if convert_to_inches:
        ds["APCP_surface"] = ds["APCP_surface"] / mm_to_in
    return ds


# def shift_storm_center(geom, x: float = 0, y: float = 0):
#     """placeholder"""
#     shift_vector = Point(x, y)
#     shifted_polygon = translate(Polygon(geom[0]), shift_vector.x, shift_vector.y)
#     return shifted_polygon


def make_gif(
    temp_dir: str,
    ds: xr.Dataset,
    file_name: str,
    watershed_poly: gpd.GeoDataFrame,
    buffer: float = 0.25,
):
    """Maps data and creates GIF from each map"""

    # Extract the geometry object from the GeoSeries
    if isinstance(watershed_poly, (gpd.GeoSeries, gpd.GeoDataFrame)):
        watershed_geom = watershed_poly.geometry.iloc[0]
    else:
        watershed_geom = watershed_poly
    # Get the bounds of the watershed
    minx, miny, maxx, maxy = watershed_geom.bounds
    minx -= buffer
    maxx += buffer
    miny -= buffer
    maxy += buffer
    cmap = Spectral_r.copy()
    cmap.set_under(color="white")

    ds = ds.where(ds["APCP_surface"] > 0)
    vmax = np.nanmax(ds["APCP_surface"].values)
    vmin = 0.00

    image_files = []

    for i in ds.time:
        x = ds.sel(time=i).APCP_surface
        quadmesh = x.plot(
            cmap=cmap, vmin=vmin, vmax=vmax, cbar_kwargs={"label": "Precip (in.)"}
        )

        # get the matplotlib axes object
        ax = quadmesh.axes
        # Set x and y limits to match watershed bounds
        ax.set_xlim(minx, maxx)
        ax.set_ylim(miny, maxy)
        # Add watershed polygon to the plot
        watershed_x, watershed_y = watershed_geom.exterior.xy
        ax.plot(watershed_x, watershed_y, color="black")  # change color if necessary
        ax.axes.get_xaxis().set_visible(False)
        ax.axes.get_yaxis().set_visible(False)

        dt_str = np.datetime_as_string(i.time.values, unit="h").replace("T", " ")
        ax.axes.set_title(dt_str)
        png = f"{temp_dir}/{dt_str.replace(' ','')}.png"
        image_files.append(png)
        plt.savefig(png)
        plt.close()

    # Sort image files by date
    image_files = sorted(image_files)

    # Create a list of image arrays from image files
    images = [imageio.imread(filename) for filename in image_files]

    # Create the output GIF file name
    output_file = f"{temp_dir}/{file_name}.gif"
    # gif_imgs.append(output_file)

    # Save the list of images as a GIF
    imageio.mimsave(output_file, images)
    return output_file


# import os
# import shutil
# import tempfile

# start = "2009-09-20"
# duration = 24

# zarfiles = generate_zarr_file_paths(start, duration, "s3://tempest/transforms/aorc/precipitation/")

# watershed_poly = fetch_watershed_geom(WATERSHED_FILE)

# ds = read_precip_dataset(zarfiles, watershed_poly)

# temp_dir = tempfile.mkdtemp()


# output_name = f"storm-{start}"
# gif = make_gif(temp_dir, ds, output_name, watershed_poly)

# # Placeholder
# shutil.copy(gif, os.path.join(os.getcwd(), f"{output_name}.gif"))
# shutil.rmtree(temp_dir)
