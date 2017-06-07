import argparse
import requests

interface_url = "https://measurements.ooni.torproject.org/api/v1/files?"

parser = argparse.ArgumentParser(description='Grab a batch of results from OONI\'s RESTful interface')
parser.add_argument("--get-arguments", help="Arguments to append to the HTTP GET request for resources. Used to filter the search.", default="")
parser.add_argument("--batch-size", help="Batch size for requests", default=100, type=int)
parser.add_argument("--output-directory", help="Where to write results", default="./")
args = parser.parse_args()

def main():
	global args
	global interface_url
	pass

if __name__ == "__main__":
	main()
