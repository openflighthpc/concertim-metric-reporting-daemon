#!/bin/bash

set -e
set -o pipefail

# The base URL against which relative URLs are constructed.
CONCERTIM_HOST=${CONCERTIM_HOST:-command.concertim.alces-flight.com}
BASE_URL=${BASE_URL:="https://${CONCERTIM_HOST}/mrd"}

# This script creates multiple int32 metrics for a single host.

# The name of the host should match the name of the one of the devices created
# via the device API. The rack and device API contains an endpoint to list
# valid names.  See the example scripts in the ct-visualisation-app repository.
HOST=${1:-comp001}

# An auth token is required for creating metrics.  One can be generated with
# the `ct-visualisation-app/docs/api/get-auth-token.sh` script and exported as
# the environment variable AUTH_TOKEN.
if [ -z "${AUTH_TOKEN}" ] ; then
  echo "$(basename $0) AUTH_TOKEN not set" >&2
  exit 1
fi


report_metric() {
  local name value body
  name="$1"
  value="$2"

  body=$(jq --null-input \
    --arg datatype "int32" \
    --arg name "${name}" \
    --arg value "${value}" \
    --arg units " " \
    --arg slope "both" \
    --arg ttl 3600 \
    '{"type": $datatype, "name": $name, "value": $value|tonumber, "units": $units, "slope": $slope, "ttl": $ttl|tonumber}'
  )

  curl -s -k \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer ${AUTH_TOKEN}" \
    -X PUT "${BASE_URL}/${HOST}/metrics" \
    -d "${body}"
}


# Reporting multiple metrics for a host involves making a separate request to
# report each metric.
#
# Here we collect all metrics into an associative array, loop over the array
# and report them.  This assumes that they have the same, data type, units,
# slope and ttl.  If not, the loop would need to be slightly more involved.
declare -A metrics=(
  [caffeine.level]=$(shuf -i 12-24 -n 1)
  [caffeine.consumption]=$(shuf -i 24-30 -n 1)
  [caffeine.capacity]=32
)

for m in "${!metrics[@]}" ; do
  report_metric "$m" "${metrics[$m]}"
done
