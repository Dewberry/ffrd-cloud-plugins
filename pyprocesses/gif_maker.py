import logging
import os
import shutil
import tempfile

import boto3
from pyprocesses.utils.gifs import (
    fetch_watershed_geom,
    generate_zarr_file_paths,
    make_gif,
    read_precip_dataset,
)
from dotenv import find_dotenv, load_dotenv

try:
    load_dotenv(find_dotenv())
except:
    logging.debug("no .env file found")

session = boto3.session.Session()
s3_client = session.client("s3")

PLUGIN_PARAMS = {
    "required": [
        "start_date",
        "duration",
        "precip_source_location",
        "watershed_file_location",
        "output_bucket",
    ],
    "optional": ["gif_output_prefix"],
}


def main(params: dict):
    start, duration = params["start_date"], params["duration"]
    precip_source, watershed_file, s3_bucket = (
        params["precip_source_location"],
        params["watershed_file_location"],
        params["output_bucket"],
    )

    if params.get("gif_output_prefix") is not None:
        output_prefix = params["gif_output_prefix"]
    else:
        output_prefix = None

    logging.info("Starting....")
    zarfiles = generate_zarr_file_paths(start, duration, precip_source)

    watershed_poly = fetch_watershed_geom(watershed_file)

    logging.info("Reading precipitation dataset")
    ds = read_precip_dataset(zarfiles, watershed_poly)
    temp_dir = tempfile.mkdtemp()
    try:
        output_name = f"storm-{start}"
        logging.info("creating gif")
        gif = make_gif(temp_dir, ds, output_name, watershed_poly)
        if output_prefix:
            s3_object_name = f"{output_prefix}/{output_name}.gif"
        else:
            s3_object_name = f"{output_name}.gif"

        s3_client.upload_file(gif, s3_bucket, s3_object_name)
        result = {"gif_s3_location": f"s3://{s3_bucket}/{s3_object_name}"}
        logging.info(f"Uploaded {gif} to {s3_bucket}/{s3_object_name}")
    except Exception as e:
        logging.error(
            f"An error occurred saving {s3_object_name} to s3 bucket '{s3_bucket}' with error: {e}"
        )
        raise
    finally:
        if os.path.exists(temp_dir):
            shutil.rmtree(temp_dir)
    return result


# main(
#     {
#         "start_date": "2009-09-17",
#         "duration": 72,
#         "precip_source_location": "s3://tempest/transforms/aorc/precipitation/",
#         "watershed_file_location": "/vsis3/tempest/watersheds/kanawha/kanawha-basin.geojson",
#         "output_bucket": "hms-bucket",
#     }
# )
