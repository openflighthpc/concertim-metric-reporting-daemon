#!/bin/bash

# Probably want to replace this with supervisor or something.

if [ $# -gt 0 ] ; then
  exec "$@"
else
  /usr/sbin/gmetad -c /opt/concertim/etc/metric-reporting-daemon/gmetad.conf
  /app/ct-metric-reporting-daemon \
    --config-file /opt/concertim/etc/metric-reporting-daemon/config.yml &

  # Wait for any process to exit.
  wait -n

  # Exit with the exit code of whichever process exited first.
  exit $?
fi
