#!/bin/bash

# Probably want to replace this with supervisor or something.

if [ ! -f /opt/concertim/etc/metric-reporting-daemon.yml ] ; then
# Move the config file to a volume, so it can be edited if we wish.
# XXX Could this be done by the new ansible playbook?
cp /app/config/config.prod.yml \
  /opt/concertim/etc/metric-reporting-daemon.yml
fi

if [ $# -gt 0 ] ; then
  exec "$@"
else
  /usr/sbin/gmetad -c /etc/ganglia/gmetad.conf
  /app/ct-metric-reporting-daemon \
    --config-file /opt/concertim/etc/metric-reporting-daemon.yml &

  # Wait for any process to exit.
  wait -n

  # Exit with the exit code of whichever process exited first.
  exit $?
fi
