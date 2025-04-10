#!/bin/bash

# call go mod tidy on all the examples
# you should call it from the same directory as this script

set -euo pipefail

for modfile in $(find . -name 'go.mod' -print0 | xargs -0)
do
    moddir=$(dirname "$modfile")
    cd "$moddir"
    echo "$moddir"
    go get -u ./...; go mod tidy
    cd - >/dev/null
done
