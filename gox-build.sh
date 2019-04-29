#!/bin/sh
set -e

#
# build releases using https://github.com/mitchellh/gox
#
gox -ldflags "-s -w -X github.com/fgma/rester/cmd.versionRevision=`git rev-parse --short HEAD`" -rebuild

rm ./rester_*_*.bz2
for f in ./rester_*_* ; do
    echo "Compressing $f"
    bzip2 -f "$f"
done
