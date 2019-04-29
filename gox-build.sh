#!/bin/sh
set -e

#
# build releases using https://github.com/mitchellh/gox
#

OUTPUT_DIR=release
OUTPUT_TEMPLATE="$OUTPUT_DIR/{{.Dir}}_{{.OS}}_{{.Arch}}"

if [ -n $1 ]; then
	OUTPUT_TEMPLATE="$OUTPUT_DIR/{{.Dir}}_$1_{{.OS}}_{{.Arch}}"
fi

mkdir $OUTPUT_DIR
gox -ldflags "-s -w -X github.com/fgma/rester/cmd.versionRevision=`git rev-parse --short HEAD`" \
  -output ${OUTPUT_TEMPLATE}
  -rebuild

rm -f ${OUTPUT_DIR}/rester_*_*.bz2
for f in ${OUTPUT_DIR}/rester_*_* ; do
    echo "Compressing $f"
    bzip2 -f "$f"
done
