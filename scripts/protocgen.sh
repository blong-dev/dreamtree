#!/usr/bin/env bash

set -e

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find . -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
      buf generate --template buf.gen.gogo.yaml $file
      # pulsar descriptors (api/) MUST regenerate in lockstep: autocli renders
      # queries and resolves msg @types through them — gogo-only regen makes
      # new fields/msgs invisible to the CLI (found the hard way in the
      # upgrade-1 rehearsal: migrated e_cap_mult didn't render).
      buf generate --template buf.gen.pulsar.yaml $file
  done
done

cd ..

cp -r github.com/blong-dev/dreamtree/* ./
rm -rf go.dreamtree.xyz github.com
if [ -d dreamtree ]; then
  mkdir -p api/dreamtree
  cp -r dreamtree/* api/dreamtree/
  rm -rf dreamtree
fi