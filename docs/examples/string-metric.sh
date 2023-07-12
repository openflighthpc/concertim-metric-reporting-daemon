#!/bin/bash

set -e
set -o pipefail

# The base URL against which relative URLs are constructed.
CONCERTIM_HOST=${CONCERTIM_HOST:-command.concertim.alces-flight.com}
BASE_URL=${BASE_URL:="https://${CONCERTIM_HOST}/mrd"}

# This script creates a single string metric for a single device.

# The Concertim ID for the device that the metric is being reported for. The
# rack and device API contains an endpoint to list valid ids.  See the example
# scripts in the ct-visualisation-app repository.
DEVICE_ID=${1:-1}

# The name of the metric being reported.
METRIC=${2:-caffeine.more}

# The value of the metric being reported.
VALUE=${3:-yes}


# An auth token is required for creating metrics.  One can be generated with
# the `ct-visualisation-app/docs/api/get-auth-token.sh` script and exported as
# the environment variable AUTH_TOKEN.
if [ -z "${AUTH_TOKEN}" ] ; then
  echo "$(basename $0) AUTH_TOKEN not set" >&2
  exit 1
fi


# Use `jq` to construct a JSON body request.
BODY=$(jq --null-input \
  --arg datatype "string" \
  --arg name ${METRIC} \
  --arg value ${VALUE} \
  --arg units "" \
  --arg slope "zero" \
  --arg ttl 3600 \
  '
{
  "type": $datatype,
  "name": $name,
  "value": $value,
  "units": $units,
  "slope": $slope,
  "ttl": $ttl|tonumber
}
'
)


# Finally, we make the request to create the metric.
curl -s -k \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -X PUT "${BASE_URL}/${DEVICE_ID}/metrics" \
  -d "${BODY}"
