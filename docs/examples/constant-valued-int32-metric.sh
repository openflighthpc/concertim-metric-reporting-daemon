#!/bin/bash

set -e
set -o pipefail
# set -x

# The base URL against which relative URLs are constructed.
CONCERTIM_HOST=${CONCERTIM_HOST:-command.concertim.alces-flight.com}
BASE_URL=${BASE_URL:="https://${CONCERTIM_HOST}/mrd"}

# This script creates a single int32 metric for a single device.

# The Concertim ID for the device that the metric is being reported for. The
# rack and device API contains an endpoint to list valid ids.  See the example
# scripts in the ct-visualisation-app repository.
DEVICE_ID=${1:-1}

# The name of the metric we are reporting.
METRIC=${2:-caffeine.capacity}

# This is a constant metric.  It wont (or at least isn't expected to) change
# over time.  Suitable for perhaps kernel version numbers.
#
# In addition to always reporting the same value, the metric is treated as
# constant by having a "slope" of "zero".
VALUE=10

# The units of the metric we are reporting.
UNITS=""


# An auth token is required for creating metrics.  One can be generated with
# the `ct-visualisation-app/docs/api/get-auth-token.sh` script and exported as
# the environment variable AUTH_TOKEN.
if [ -z "${AUTH_TOKEN}" ] ; then
  echo "$(basename $0) AUTH_TOKEN not set" >&2
  exit 1
fi


# Use `jq` to construct a JSON body request.
BODY=$(jq --null-input \
  --arg datatype "int32" \
  --arg name ${METRIC} \
  --arg value ${VALUE} \
  --arg units "${UNITS}" \
  --arg slope "zero" \
  --arg ttl 3600 \
  '
{
  "type": $datatype,
  "name": $name,
  "value": $value|tonumber,
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
