import argparse
import os
import requests
import sys
import time

parser = argparse.ArgumentParser(description='Grab a batch of results from OONI\'s RESTful interface')
parser.add_argument("--get-arguments", help="Arguments to append to the HTTP GET request for resources. Used to filter the search.", default="")
parser.add_argument("--output-directory", help="Where to write results", default="./")
args = parser.parse_args()

# Constants
interface_url = "https://measurements.ooni.torproject.org/api/v1/files?"
retries = 10

# State maintained for backoff
current_pause = .1

def request_with_backoff(url):
	global current_pause
	global last_success
	global retries
	for i in range(retries):
		print(current_pause)
		sys.stdout.flush()
		time.sleep(current_pause)
		r = requests.get(url)
		if r.status_code < 300:
			current_pause = max(0., current_pause - .01)
			break
		else:
			current_pause = max(.01, current_pause * 2)
	if r.status_code >= 300:
		print("Error requesting from url: {} - {}".format(url, r.status_code), file=sys.stderr)
		exit(1)
	return r

def main():
	global args
	global interface_url
	if "./" != args.output_directory and not os.path.exists(args.output_directory):
		os.makedirs(args.output_directory)
	current_url = interface_url 
	# repeat until we get all of our results
	while current_url != None:
		r = request_with_backoff(current_url)
		result = r.json()
		metadata = result['metadata']
		current_url = metadata['next_url']
		print("Got metadata for page {} of {}: {} total items".format(metadata['current_page'], metadata['pages'], metadata['count']))
		sys.stdout.flush()
		for result in result['results']:
			downloaded = request_with_backoff(result['download_url'])
			index = result['index']
			print("Got data for element {}".format(index))
			sys.stdout.flush()
			with open(os.path.join(args.output_directory, "{}.json".format(index)), 'w') as f:
				f.write(downloaded.text)

if __name__ == "__main__":
	main()
