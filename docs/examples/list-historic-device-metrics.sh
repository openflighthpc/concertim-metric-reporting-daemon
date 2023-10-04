#!/bin/bash

set -e
set -o pipefail
set -x

# The base URL against which relative URLs are constructed.
CONCERTIM_HOST=${CONCERTIM_HOST:-command.concertim.alces-flight.com}
BASE_URL=${BASE_URL:="https://${CONCERTIM_HOST}/mrd"}

# This script lists the historic metric values for the given device between the given timestamps.

# The ID of the device we are querying.
DEVICE_ID=${1:-1}
# The name of the metric we are querying.
METRIC=${2:-caffeine.level}

if type ruby >/dev/null 2>&1 ; then
	# If ruby is installed let's display the metrics for the last hour by
	# default.  Not that the times are given as an integer number of seconds
	# since the 1970-01-01-00:00:00 UTC.
	START=${3:-$(ruby -e 'puts (Time.now.utc - 60*60).to_i')}
	END=${4:-$(ruby -e 'puts (Time.now.utc).to_i')}
else
	START=${3:-1696431210}
	END=${4:-1696431300}
fi


# An auth token is required for creating metrics.  One can be generated with
# the `ct-visualisation-app/docs/api/get-auth-token.sh` script and exported as
# the environment variable AUTH_TOKEN.
if [ -z "${AUTH_TOKEN}" ] ; then
  echo "$(basename $0) AUTH_TOKEN not set" >&2
  exit 1
fi

curl -s -k \
  -H 'Accept: application/json' \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -X GET "${BASE_URL}/device/${DEVICE_ID}/metrics/${METRIC}/historic/${START}/${END}"
