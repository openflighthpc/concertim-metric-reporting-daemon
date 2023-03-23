#!/bin/bash

set -e
set -o pipefail

# The base URL against which relative URLs are constructed.
CONCERTIM_HOST=${CONCERTIM_HOST:-command.concertim.alces-flight.com}
BASE_URL=${BASE_URL:=-"https://${CONCERTIM_HOST}/mrd"}


# This script creates a single metric for a single host.  The host is given
# below.  This name needs to be known to Concertim and needs to be for
# something that can have metrics assigned to it, such as a rack, chassis,
# device, power strip or sensor.  Valid names for the demo data include:
# `Rack-1`, `HP-Blade-01`, `comp001`, `pdu01` or `temp01`.
#
# The rack and device API will contain an endpoint to list valid names.  For
# now you can obtain them from the Concertim UI.
HOST=${1:-comp001}


# An auth token is required for creating metrics.  One can be generated with
# the `ct-visualisation-app/docs/api/get-auth-token.sh` script and exported as
# the environment variable AUTH_TOKEN.
if [ -z "${AUTH_TOKEN}" ] ; then
  echo "$(basename $0) AUTH_TOKEN not set" >&2
  exit 1
fi


# Use `jq` to construct a JSON body request.
#
# The name is prefixed with `ct.mrd`.  There is no need for such a prefix,
# however prefixing with `ct.` groups the metric with other user-defined
# metrics on the device's metric page.
BODY=$(jq --null-input \
  --arg datatype "string" \
  --arg name "ct.mrd.caffeine.more" \
  --arg value "yes" \
  --arg units " " \
  --arg slope "zero" \
  --arg ttl 3600 \
  '{"type": $datatype, "name": $name, "value": $value, "units": $units, "slope": $slope, "ttl": $ttl|tonumber}'
)


# Finally, we make the request to create the metric.
curl -s -k \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -X PUT "${BASE_URL}/${HOST}/metrics" \
  -d "${BODY}"
