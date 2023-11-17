#!/bin/bash

# Probably want to replace this with supervisor or something.

if [ $# -gt 0 ] ; then
  exec "$@"
else
  GMETAD_CONFIG_FILE=${GMETAD_CONFIG_FILE:-docker/gmetad.conf}
  /usr/sbin/gmetad -c "${GMETAD_CONFIG_FILE}"
  MRD_CONFIG_FILE=${MRD_CONFIG_FILE:-config/config.dev.yml}
  $(go env GOPATH)/bin/air -- --config-file "${MRD_CONFIG_FILE}" &

  # Wait for any process to exit.
  wait -n

  # Exit with the exit code of whichever process exited first.
  exit $?
fi
