#!/bin/bash

# Probably want to replace this with supervisor or something.

if [ $# -gt 0 ] ; then
  exec "$@"
else
  CONFIG_FILE=${CONFIG_FILE:-/opt/concertim/etc/metric-reporting-daemon/config.yml}
  /app/ct-metric-reporting-daemon --config-file "${CONFIG_FILE}"
fi
