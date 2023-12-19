import sys

from make_gif import PLUGIN_PARAMS, main
from papipyplug import parse_input, plugin_logger, print_results

if __name__ == "__main__":
    # start plugin logger
    plugin_logger()

    # Read, parse, and verify input parameters
    input_params = parse_input(sys.argv, PLUGIN_PARAMS)

    result = main(input_params)

    print_results(result)
