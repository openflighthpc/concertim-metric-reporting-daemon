#!/bin/bash

set -e
set -o pipefail

# The base URL against which relative URLs are constructed.
# BASE_URL="http://localhost:3000"
BASE_URL="https://command.concertim.alces-flight.com/mrd"


# This script creates a single metric for a single host.  The host is given
# below.  This name needs to be known to Concertim and needs to be for
# something that can have metrics assigned to it, such as a rack, chassis,
# device, power strip or sensor.  Valid names for the demo data include:
# `Rack-1`, `HP-Blade-01`, `comp001`, `pdu01` or `temp01`.
#
# The rack and device API will contain an endpoint to list valid names.  For
# now you can obtain them from the Concertim UI.
HOST="comp001"


# To make testing this API easier, it exports an unauthenticated endpoint with
# which an auth token can be created.  Obviously this is a security issue and
# this end point will be removed in a future version.
#
# The intended mechanism to create auth tokens is via the Concertim UI and
# re-use them on each request.  Here we do what's easiest and create a new one
# for each request.
AUTH_TOKEN=$(curl -s -k -X POST "${BASE_URL}/token" -d '{}' | jq -r .token)


# Use `jq` to construct a JSON body request.
#
# The name is prefixed with `ct.mrd`.  There is no need for such a prefix,
# however prefixing with `ct.` groups the metric with other user-defined
# metrics on the device's metric page.
#
# We use `shuf` to generate a random number, in this case between 12 and 24
# inclusive.
BODY=$(jq --null-input \
  --arg datatype "int32" \
  --arg name "ct.mrd.caffeine.level" \
  --arg value "$(shuf -i 12-24 -n 1)" \
  --arg units " " \
  --arg slope "both" \
  --arg ttl 3600 \
  '{"type": $datatype, "name": $name, "value": $value|tonumber, "units": $units, "slope": $slope, "ttl": $ttl|tonumber}'
)


# Finally, we make the request to create the metric.
curl -s -k \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${AUTH_TOKEN}" \
  -X PUT "${BASE_URL}/${HOST}/metrics" \
  -d "${BODY}"

