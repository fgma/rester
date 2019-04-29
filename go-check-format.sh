#!/bin/sh

GOFMT_ERRORS=`gofmt -l . | grep -v '^vendor/'`

if [ -n "$GOFMT_ERRORS" ]; then
  echo $GOFMT_ERRORS
  exit 1
fi
