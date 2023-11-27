#!/bin/bash

# Probably want to replace this with supervisor or something.

if [ $# -gt 0 ] ; then
  exec "$@"
else
  CONFIG_FILE=${CONFIG_FILE:-config/config.dev.yml}
  $(go env GOPATH)/bin/air -- --config-file "${CONFIG_FILE}"
fi
