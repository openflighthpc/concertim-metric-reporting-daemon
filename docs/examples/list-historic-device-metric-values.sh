#!/bin/bash

set -e
set -o pipefail
# set -x

# The base URL against which relative URLs are constructed.
CONCERTIM_HOST=${CONCERTIM_HOST:-command.concertim.alces-flight.com}
BASE_URL=${BASE_URL:="https://${CONCERTIM_HOST}/mrd"}

# This script lists the historic metric values for the given device between the given timestamps.

# The ID of the device we are querying.
DEVICE_ID=${1:-1}
# The name of the metric we are querying.
METRIC=${2:-caffeine.level}
# The duration we are querying: last hour; last day or last quarter.
DURATION=${3:-hour}

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
  -X GET "${BASE_URL}/devices/${DEVICE_ID}/metrics/${METRIC}/historic/last/${DURATION}"

# Instead of specifying the last hour/day/quarter, you can specify start and
# end time stamps.
#
#  START=1696431210
#  END=1696431300
#  curl -s -k \
#     -H 'Accept: application/json' \
#     -H "Authorization: Bearer ${AUTH_TOKEN}" \
#     -X GET "${BASE_URL}/metrics/${METRIC}/historic/${START}/${END}"
